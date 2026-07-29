# ACRE

An Incident Driven Automatic Code Remediation Engine.

<img width="1470" height="220" alt="Screenshot 2026-06-29 at 4 03 01 AM" src="https://github.com/user-attachments/assets/ce4813ad-fe67-4d87-9642-f4d332b3e0a5" />

## Features

* **Ticket Ingestion**: Loads and parses structured JSON incident reports.
* **Open Knowledge Format (OKF) Integration**: Automatically detects and ingests OKF v0.1 specification directories (e.g. `OKF/<repoName>/`) containing structured markdown files and YAML metadata, appending them to the diagnostic prompt to guide targeting.
* **Prompt Construction**: Programmatically builds detailed diagnostic prompts for OpenCode with codebase styling, backend focus scopes, and senior engineering guidelines.
* **OpenCode CLI Integration**: Executes OpenCode non-interactively using the `--dangerously-skip-permissions` sandbox to repair the code.
* **Dynamic Build & Test Runners**: Scans the target codebase to detect solutions (`.sln`/`.slnx`) and test projects, executing clean builds (`dotnet build <sln>`) and test suites (`dotnet test <sln>`) while bypassing docker-compose errors, falling back to console runners where appropriate.
* **Self-Healing Loop**: Automatically detects compilation/test failures and feeds the errors back to OpenCode to retry (up to 3 times).
* **PR Generation on Both Success and Failure**: Under the `--pr` flag, commits and pushes changes to a ticket-specific branch. If validation fails, it still submits a PR but tags the title and body with `(Validation Failed)` warnings so engineers can inspect intermediate states.
* **Remediation Report**: Generates timestamped run reports with command logs, stdout/stderr logs, structured analysis, final outcome states, and pull request URLs.
* **Robust Credential Loading**: Searches for `.env` files in both the current working directory and directly adjacent to the compiled `acre` binary.
* **OKF Documentation Generator**: Supports scanning codebases and automatically generating/updating detailed, conformant OKF v0.1 documentation directories under `OKF/` via the `--okf` CLI command.

## Prerequisites

* **Go compiler** (version 1.20 or later, required to build ACRE)
* **LLM Coding CLI**: OpenCode (install via `brew install opencode`), Codex, or ClaudCode
* **Target SDK**: .NET SDK (only required for C# repositories), Node.js (for JS/TS), Python, Go, etc.

## Setup

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
   - **Remediation Mode (with auto-PR)**: Modifies files, compiles, runs tests, self-heals, and opens a PR with actual code changes:
     ```bash
     ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs --pr
     ```
   - **Recommendations Mode (no code changes)**: Analyzes codebase, discards source code modifications, generates a structured analysis report (`recommendations.md`) with a confidence score/justification, commits only the report, and opens a PR with details:
     ```bash
     ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs -r
     ```
   - **Test Mode (`--test`)**: Lightweight diagnostic mode. OpenCode analyzes the codebase for a fix, ACRE executes a full solution compilation check, skips git operations (no branch/commit/push) and regression tests, and outputs a prefilled manual GitHub PR link directly in the console:
     ```bash
     ./acre --ticket ../tickets/ENG-0001.json --repo /path/to/target/repo --runs-dir ../runs --test
     ```
4. Run the orchestrator to generate OKF v0.1 documentation for a repository (supports relative or absolute full paths):
   ```bash
   ./acre --okf /Users/avinash/Desktop/blurr/CodeBase/eShop-main
   ```
   To focus documentation scanning and indexing on a specific module subdirectory (supports absolute full path or relative repo path), add the `--scope` parameter:
   ```bash
   ./acre --okf /Users/avinash/Desktop/blurr/CodeBase/eShop-main --scope /Users/avinash/Desktop/blurr/CodeBase/eShop-main/src/Services/Basket
   ```

## Language & CLI Customization

ACRE is designed to be language and tool-agnostic. You can easily adapt it:
* **LLM CLI**: Swap `opencode` for any other CLI coding assistant (e.g., `codex`, `claudcode`) by modifying the command string inside [orchestrator/opencode/opencode.go](file:///Users/avinash/Desktop/blurr/ACRE/orchestrator/opencode/opencode.go).
* **Compiling & Building**: Edit the build command parser inside [orchestrator/build/build.go](file:///Users/avinash/Desktop/blurr/ACRE/orchestrator/build/build.go) to target other compilers (e.g. `npm run build`, `make`, `cargo build`).
* **Regression Testing**: Edit [orchestrator/test/test.go](file:///Users/avinash/Desktop/blurr/ACRE/orchestrator/test/test.go) to target your test runner (e.g. `pytest`, `npm test`, `go test`).

## Jira Extractor

A CLI utility under `JiraExtractor` to download and structure incident tickets from Jira:

### Setup
Configure the following inside your `.env` file:
```env
JIRA_PAT=your_jira_personal_access_token
JIRA_BASE_URL=https://jira.example.com
```

### Usage
Build and run the extractor:
```bash
cd "JiraExtractor"
go build -o jira main.go
./jira -ticket <TICKET_ID>
```
This generates `<TICKET_ID>.json` with structured ticket fields (summary, description, acceptance criteria, and sorted comments) in the current directory and prints them to stdout.

## Snyk Vulnerability Remediation (`--snyk`)

ACRE includes an automated security vulnerability remediation engine powered by Snyk static code analysis (`snyk code test`).

### Features & Capabilities
* **Automated Snyk Parsing**: Parses both standard plaintext output from `snyk code test` (extracting Finding IDs, severity tags `[HIGH]`, `[MEDIUM]`, `[LOW]`, `[CRITICAL]`, line numbers, and file paths) and JSON formatted output.
* **Similarity Vulnerability Grouping**: Groups targeted vulnerabilities by vulnerability category/rule title and module/component directory so related security findings are fixed together in optimal batches (1 to 10+ findings).
* **Configurable Severities & Retries (`snyk.Config`)**: Target severities (`HIGH`, `CRITICAL`, `MEDIUM`, `LOW`) and max retry attempts (`MaxSnykFixRetries`, `MaxBuildRetries`) can be easily configured in Go code.
* **Interactive OKF Validation & Updates**: If `--OKF` path is passed but missing, prompts interactively in the CLI: `"Cant find OKF, should i proceed without one? Y/N"`. Upon successful vulnerability resolution and verification, automatically appends knowledge records to `OKF/snyk_remediations.md`.
* **Dual Self-Healing Verification Loops**:
  1. **Snyk Fix Loop (Max 3 Tries)**: Applies minimal, targeted security fixes via OpenCode, re-tests with `snyk code test`, and verifies if targeted findings are eliminated.
  2. **Build Verification Loop (Max 3 Tries)**: Compiles the solution (`dotnet build`). If compilation fails, feeds build error output back to OpenCode for repair.
* **Complete Report Package**: Outputs detailed artifacts in the `--report` directory:
  - `report.md`: Structured breakdown of initial severities, bunch actions, resolved findings, code modifications, and final summary statistics.
  - `prompt.md`: System prompts fed to OpenCode during execution.
  - `opencode_output.md`: Concatenated raw logs of all OpenCode runs.
  - `logs.md`: Complete execution trace log of the orchestrator CLI.

### Command Usage
```bash
./acre --snyk "/path/to/target/repo" --report "/path/to/report/dir" --OKF "/path/to/okf/folder" --debug 1
```

#### Parameters
* `--snyk <full path>` *(Required)*: Path to the target repository to scan and resolve Snyk vulnerabilities for.
* `--report <full path>` *(Optional)*: Directory path where output reports (`report.md`, `prompt.md`, `opencode_output.md`, `logs.md`) will be saved (defaults to `runs/snyk_report`).
* `--OKF <full path>` *(Optional)*: Path to OKF documentation directory or file.
* `--debug <a>` *(Optional)*: Max number of vulnerability bunches to process (e.g. `--debug 1` stops after remediating 1 bunch; omitted processes all target severity bunches).




