# ACRE Domain Skills & Prompt Customization (`--skill`)

This directory contains modular domain skills, team-specific conventions, and business logic guidelines that can be injected into ACRE during incident remediation using the `--skill` flag:

```bash
./acre --ticket ../tickets/ENG-0001.json --repo /path/to/repo --runs-dir ../runs --skill skills/sample_skill.md --pr
```

---

## How `--skill` Works

When you pass `--skill <path>` to ACRE:
1. **Single File**: If `--skill` points to a specific `.md` file (e.g. `--skill skills/sample_skill.md`), ACRE reads and injects the document under `## Custom Domain Skills & Engineering Guidelines` in the OpenCode agent prompt.
2. **Directory**: If `--skill` points to a folder (e.g. `--skill skills/`), ACRE automatically scans the directory, extracts all `.md` skill files, and injects them into the prompt with corresponding headers.

---

## Why Use Custom Skills?

* **No Go Code Modifications**: Developers and teams can adjust business rules, codebase conventions, and architectural constraints by modifying simple Markdown files rather than editing or recompiling `prompt.go`.
* **Repository-Specific or Team-Specific**: You can maintain dedicated skill files for different repositories or microservices (e.g. `billing_rules.md`, `auth_guidelines.md`, `logging_standards.md`).
* **Version-Controlled**: Skill files can be checked into git along with your source code for continuous refinement.

---

## Authoring Custom Skills

To create a new skill, copy [`sample_skill.md`](sample_skill.md) or create a new `.md` file defining:
* **Business Invariants**: Rules and constraints specific to your domain (e.g., currency precision, tenant isolation).
* **Architecture Rules**: Layering standards, dependency injection policies, and error handling patterns.
* **Testing Conventions**: Unit test naming and isolation guidelines.
* **Prohibited Anti-Patterns**: Explicit anti-patterns the agent should never introduce.
