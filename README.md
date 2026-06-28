# ACRE

An Incident Driven Automatic Code Remediation Engine.

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

## Setup

1. Configure credentials inside a `.env` file right next to the binary or at the root:
   ```env
   GITHUB_TOKEN=your_personal_access_token
   ```
2. Build the orchestrator:
   ```bash
   cd orchestrator
   go build -o acre main.go
   ```
3. Run the orchestrator on a target repository:
   ```bash
   ./acre --ticket ../tickets/ENG-0001.json --repo ../CodeBase/eShop-main --runs-dir ../runs --pr
   ```

