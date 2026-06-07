# ACRE

An Incident Driven Automatic Code Remediation Engine

## Features

* **Ticket Ingestion**: Loads and parses structured JSON incident reports.
* **Prompt Construction**: Programmatically builds detailed diagnostic prompts for Codex.
* **Codex CLI Integration**: Executes Codex non-interactively using the workspace-write sandbox to repair the code.
* **Build Validation**: Automatically runs dotnet build to verify compilation.
* **Test Validation**: Runs project-specific regression tests to verify fixes.
* **Self-Healing Loop**: Automatically detects compilation/test failures and feeds the errors back to Codex to retry (up to 3 times).
* **Remediation Report**: Generates timestamped run reports with stdout/stderr logs and final outcome states.

## Setup

1. Clone the repository.
2. Build the orchestrator:
   ```bash
   cd orchestrator
   go build -o acre main.go
   ```
3. Run the orchestrator:
   ```bash
   ./acre --ticket ../tickets/ENG-0001.json --repo ../CodeBase/ApiRateLimiter --runs-dir ../runs
   ```
