# ACRE

An Incident-Driven Automatic Code Remediation Engine.

<img width="1470" height="220" alt="Screenshot 2026-06-29 at 4 03 01 AM" src="https://github.com/user-attachments/assets/ce4813ad-fe67-4d87-9642-f4d332b3e0a5" />

## Features

* **Ticket Ingestion**: Loads and parses structured JSON incident reports.
* **Domain Skills & Custom Rules (`--skill`)**: Ingests custom `.md` skills or skill directories (e.g. `--skill skills/sample_skill.md` or `--skill skills/`) to enforce team coding rules, architecture standards, and domain invariants without editing Go code.
* **Open Knowledge Format (OKF) Integration**: Automatically detects and ingests OKF v0.1 specification directories (e.g. `OKF/<repoName>/`) containing structured markdown files and YAML metadata.
* **Prompt Construction**: Programmatically builds detailed diagnostic prompts for OpenCode with codebase styling, backend focus scopes, and senior engineering guidelines.
* **OpenCode CLI Integration**: Executes OpenCode non-interactively using the `--dangerously-skip-permissions` sandbox to repair the code.
* **Dynamic Build & Test Runners**: Scans target codebases to detect solutions (`.sln`/`.slnx`) and test projects, executing clean builds (`dotnet build <sln>`) and test suites (`dotnet test <sln>`).
* **Snyk SARIF Extraction & Normalization (`--snykjson`)**: Runs `snyk code test --json`, parses SARIF 2.1.0 schema, normalizes findings into a unified JSON format, supports filtering by `--ruleid`, and outputs clean normalized JSON files (`snykOutput/Output<datetime>.json`) or directly to stdout (`--cli`).
* **Targeted Snyk Security Remediation (`--snyk`)**: Ingests normalized Snyk finding JSON files (`Output<datetime>.json`), accepts reference instruction files (`--REF`), guides OpenCode to apply security fixes, auto-detect build tools, compile projects, and verify resolution per `ruleId` in real-time.
* **Self-Healing Loop**: Automatically detects compilation/test failures and feeds errors back to OpenCode to retry.
* **PR Generation**: Under the `--pr` flag, commits and pushes changes to a ticket-specific branch.
* **Remediation Report**: Generates timestamped run reports (`report.md`, `prompt.md`, `opencode_output.md`, `logs.md`).

---

## Prerequisites

* **Go Compiler** (version 1.20 or later, required to build ACRE)
* **LLM Coding CLI**: OpenCode (install via `brew install opencode`), Codex, or ClaudCode
* **Target SDK**: .NET SDK (only required for C# repositories), Node.js (for JS/TS), Python, Go, etc.

---

## Setup & Building

1. Configure credentials inside a `.env` file right next to the binary or at the root:
   ```env
   GITHUB_TOKEN=your_personal_access_token
   OPENCODE_MODEL=your_custom_model_name  # Optional. Defaults to "opencode/big-pickle" if not set.
   ```
2. Build the orchestrator:
   ```bash
   cd orchestrator
   go build -o acre main.go
   ```

---

## Incident Remediation Workflow

- **Remediation Mode (with auto-PR)**:
  ```bash
  ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs --pr
  ```
- **Remediation with Custom Domain Skill / Team Rules (`--skill`)**:
  ```bash
  ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs --skill ../skills/sample_skill.md --pr
  ```
- **Recommendations Mode (no code changes, creates `recommendations.md` & opens PR)**:
  ```bash
  ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs -r
  ```
- **Test Mode (`--test` dry-run, no git operations)**:
  ```bash
  ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs --test
  ```

> [!TIP]
> **Custom Skills (`--skill <path>`)**: Pass a single `.md` file or an entire directory of `.md` skill files (e.g. [`skills/sample_skill.md`](skills/sample_skill.md)). ACRE injects these guidelines under `## Custom Domain Skills & Engineering Guidelines` in the agent prompt, allowing team and domain rules to be customized dynamically without modifying `prompt.go`.

---

## Snyk Vulnerability Remediation Workflow

ACRE includes a two-stage automated security vulnerability remediation engine powered by Snyk static code analysis.

### Stage 1: Extraction & SARIF Normalization (`--snykjson`)

Runs `snyk code test --json`, saves raw output to `snykOutput.json`, parses SARIF 2.1.0 schema, and normalizes findings into the target JSON structure:

```json
{
  "ruleId": "csharp/PT",
  "title": "Path Traversal",
  "level": "error",
  "message": "Unsanitized input from CLI argument flows into os.WriteFile",
  "file": "Controllers/FileController.cs",
  "line": 45,
  "cwe": ["CWE-22"],
  "precision": "very-high"
}
```

#### Extract & Save Normalized JSON File
```bash
./acre --snykjson --repo "/path/to/repo"
```
*Output*: Creates `snykOutput/Output<datetime>.json`.

#### Extract with Specific Rule ID Filter
```bash
./acre --snykjson --repo "/path/to/repo" --ruleid csharp/PT
```
*Output*: Filters findings to only include those matching `--ruleid csharp/PT`.

#### Output Directly to Terminal Output (`--cli`)
```bash
./acre --snykjson --repo "/path/to/repo" --ruleid csharp/PT --cli
```
*Output*: Prints the clean normalized JSON array directly to stdout (useful for real-time agent verification).

---

### Stage 2: Remediation & Verification (`--snyk`)

Ingests the normalized `Output<datetime>.json` file, builds a targeted system prompt for OpenCode, and executes the remediation pipeline.

#### Command Usage
```bash
./acre --snyk "snykOutput/Output<datetime>.json" --repo "/path/to/repo" --report "/path/to/report/dir" --REF "references"
```

#### Key Capabilities:
* **Reference Guidance (`--REF`)**: Passes optional reference `.md` files or instruction folders (e.g. `--REF references/csharp_sqli.md` or `--REF references/`). ACRE scans the folder, extracts all `.md` guides, and injects them under `## Reference & Standard Recommended Instructions` in the OpenCode remediation prompt.
* **Built-in Reference Guides (`references/`)**:
  * [`references/csharp_sqli.md`](references/csharp_sqli.md): Complete guide for C# SQL Injection (parameterized queries, procedure allowlists, parameterized IN-lists).
  * [`references/general_remediation.md`](references/general_remediation.md): Universal patterns for Path Traversal, Secrets, CSRF, SSRF, and Command Injection.
  * [`references/example.md`](references/example.md): Template for creating custom team/project-specific security reference guides.
* **Targeted Normalized Context**: Passes exact normalized findings (`findingId`, `ruleId`, `title`, `file`, `line`, `message`, `cwe`, `precision`, `codeFlow`) to OpenCode.
* **Automated Compilation Checks**: Instructs OpenCode to detect the target repository's native solution build tool (e.g. `dotnet build`, `npm run build`, `go build`, `mvn compile`) and compile the project to verify build success with zero compiler errors.
* **Real-time CLI Verification Loop**: OpenCode verifies vulnerability resolution during execution by executing the resolved ACRE binary:
  ```powershell
  & "C:\path\to\acre.exe" --snykjson --repo . --ruleid <ruleId> --cli
  ```
  Iterates until `[]` (0 findings) is returned.
* **Verification Scan**: Post-remediation verification matches remaining findings with line-shift tolerance using stable finding IDs and rule fallbacks.
* **Output Artifacts**: Saves `report.md`, `prompt.md`, `opencode_output.md`, and `logs.md` in `--report`.

---

## OKF Documentation Generator

Supports scanning codebases and generating/updating conformant OKF v0.1 documentation directories under `OKF/`:

```bash
./acre --okf /path/to/target/repo [--scope /path/to/target/repo/src/SubModule]
```

---

## Jira Extractor

A CLI utility under `JiraExtractor/` to fetch and format Jira tickets into structured JSON for ACRE ingestion.

### 1. Environment Configuration (`.env`)
Configure credentials in `.env` (at repository root or next to executable):
```env
JIRA_PAT=your_jira_personal_access_token
JIRA_BASE_URL=https://jira.yourcompany.com
```

### 2. Build & Usage
```bash
cd JiraExtractor
go build -o jira main.go
./jira -ticket <TICKET_ID>
```
*Example*: `./jira -ticket ENG-0001`

### 3. Output Ticket Schema (`<TICKET_ID>.json`)
Outputs formatted ticket to `<TICKET_ID>.json`:
```json
{
  "ticket_id": "ENG-0001",
  "summary": "Fix payment gateway timeout error",
  "description": "Full incident description...",
  "acceptance_criteria": "Acceptance criteria from customfield_11813",
  "comments": [
    {
      "created": "2026-07-07T22:26:39.000+0000",
      "author": "Alice Doe",
      "body": "Initial triage findings..."
    }
  ]
}
```
*Note*: Comments are automatically sorted chronologically (`old → new`).
