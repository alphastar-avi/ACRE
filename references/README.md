# ACRE Remediation Reference Guides (`--REF`)

This directory contains modular reference guides and secure coding standards that can be passed to ACRE using the `--REF` flag:

```bash
./acre --snyk "snykOutput/Output<datetime>.json" --repo "/path/to/repo" --REF "references"
```

---

## How `--REF` Works

When you supply `--REF` to ACRE:
1. **Single File**: If `--REF` points to a specific `.md` file (e.g. `--REF references/csharp_sqli.md`), ACRE injects the entire file under `## Reference & Standard Recommended Instructions` in the OpenCode remediation prompt.
2. **Directory**: If `--REF` points to a folder (e.g. `--REF references`), ACRE scans the folder, reads all `.md` files, and automatically injects them into the prompt with corresponding headings.

---

## Available Built-in Reference Guides

| File | Description | Applicable Rule IDs |
| :--- | :--- | :--- |
| [`csharp_sqli.md`](csharp_sqli.md) | Gold-standard parameterized query, stored procedure allowlists, and IN-list remediation patterns for C# / .NET | `csharp/Sqli`, `csharp/SqlInjection` |
| [`general_remediation.md`](general_remediation.md) | Universal secure remediation patterns for Path Traversal, Secrets, CSRF, SSRF, and Command Injection | `*/PT`, `*/HardcodedCredentials`, `*/Csrf`, `*/Ssrf`, `*/CommandInjection` |
| [`example.md`](example.md) | Boilerplate template demonstrating how developers can create custom project-specific reference guides | *Custom rules* |

---

## Authoring Custom Reference Guides

To add custom guidelines for your organization or repository, create a `.md` file following the template in [`example.md`](example.md) and pass your reference folder via `--REF`.
