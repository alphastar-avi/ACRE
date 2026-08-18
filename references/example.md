# Custom Remediation Reference Template

Use this template to create team-specific or framework-specific security guidelines to pass into ACRE via `--REF`.

---

## Targeted Rule IDs
* `language/VulnerabilityRuleID` (e.g. `csharp/Sqli`, `javascript/XSS`, `python/PathTraversal`)
* Associated CWEs: `CWE-89`, `CWE-79`, `CWE-22`

---

## 1. Vulnerability Description & Root Cause
* Provide a 1-2 sentence description explaining what causes this vulnerability in your codebase.
* Highlight common mistakes (e.g. *"String interpolation in repository methods"* or *"Accepting unvalidated filenames in download endpoints"*).

---

## 2. Core Security Remediation Rule
* State the primary fix pattern (e.g. *"Always use parameterized queries with `SqlParameter` or ORM safe bindings"*).
* State what is **prohibited** (e.g. *"Do NOT use custom string quote-escaping functions like `Replace("'", "''")` as static analysis will still treat them as tainted"*).

---

## 3. Secure Code Examples

### ❌ Insecure Pattern (Before):
```csharp
// Example of vulnerable code
var query = $"SELECT * FROM Users WHERE Username = '{username}'";
var cmd = new SqlCommand(query, conn);
```

### ✅ Secure Pattern (After):
```csharp
// Example of parameterized remediation
var query = "SELECT * FROM Users WHERE Username = @Username";
var cmd = new SqlCommand(query, conn);
cmd.Parameters.Add(new SqlParameter("@Username", SqlDbType.VarChar, 100) { Value = username });
```

---

## 4. Special Architecture Considerations
* **Dynamic Identifiers**: If procedure names or table names must be dynamic, map them through a server-side `Dictionary<string, string>` or strict `enum` rather than passing raw user input.
* **IN-Lists**: For dynamic collections, generate individual parameter placeholders (`@id0, @id1...`) or use table-valued parameters.
* **Unit Tests**: Update any existing repository unit tests that asserted exact SQL string matches to assert parameterized command patterns.
