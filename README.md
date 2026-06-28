# ACRE

An Incident Driven Automatic Code Remediation Engine.

## Features

* **Ticket Ingestion**: Loads and parses structured JSON incident reports.
* **Prompt Construction**: Programmatically builds detailed diagnostic prompts for OpenCode with codebase styling and senior engineering guidelines.
* **OpenCode CLI Integration**: Executes OpenCode non-interactively using the `--dangerously-skip-permissions` sandbox to repair the code.
* **Build Validation**: Automatically compiles the target repository to verify the fixes.
* **Test Validation**: Runs project-specific regression tests to verify no new regressions are introduced.
* **Self-Healing Loop**: Automatically detects compilation/test failures and feeds the errors back to OpenCode to retry (up to 3 times).
* **Remediation Report**: Generates timestamped run reports with command logs, stdout/stderr logs, structured analysis, and final outcome states.
* **Git & PR Automation**: Automatic branch creation, pushing, and pull request generation with structured summaries when the `--pr` option is set.

## Setup

1. Build the orchestrator:
   ```bash
   cd orchestrator
   go build -o acre main.go
   ```
2. Run the orchestrator:
   ```bash
   ./acre --ticket ../tickets/ENG-0001.json --repo ../CodeBase/ApiRateLimiter --runs-dir ../runs
   ```
3. Run the orchestrator with automated PR generation:
   ```bash
   ./acre --ticket ../tickets/ENG-0001.json --repo ../CodeBase/ApiRateLimiter --runs-dir ../runs --pr
   ```
