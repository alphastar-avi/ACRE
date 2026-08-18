# C# / .NET SQL Injection Remediation Guide (`csharp/Sqli`)

* **Rule IDs:** `csharp/Sqli`, `csharp/SqlInjection`
* **CWE:** `CWE-89` (Improper Neutralization of Special Elements used in an SQL Command)
* **Compatibility:** Applies to both **legacy .NET Framework** (classic ADO.NET, `System.Data.SqlClient`, `SqlHelper`, `DataTable`/`DataSet`, ASP.NET MVC/Web API) and **modern .NET** (.NET 6/8/9, `Microsoft.Data.SqlClient`, Dapper, EF Core).

---

## 1. Core Principles: Style & Business Logic Preservation

1. **Do NOT Break Signatures or Logic**: Maintain existing method signatures, return types, repository interfaces, and data transformation logic.
2. **Match Existing Codebase Style**:
   - In legacy ADO.NET / `SqlHelper` codebases, use classic `SqlParameter` arrays with appropriate `SqlDbType` and parameter collections.
   - In ORM / Dapper codebases, use framework-native parameter objects.
3. **Targeted Surgical Edits**: Only modify the dynamic string assembly to introduce parameter placeholders and pass typed parameters. Do NOT perform unnecessary broad refactoring.

---

## 2. Why Snyk Taint Analysis Flags SQL Injection

Snyk Code performs static Abstract Syntax Tree (AST) taint tracking from HTTP/untrusted sources to database sinks (`SqlCommand.CommandText`, `SqlHelper.Execute...()`, `DbSet.FromSqlRaw()`).

### Important SAST Engine Constraints:
* **Custom String Sanitization Does NOT Clear Taint**: Calling custom methods like `input.Replace("'", "''")` or `ValidateSqlInput(input)` will **never** satisfy Snyk's taint engine. Snyk traces the original string object as tainted.
* **Regex Extraction Does NOT Clear Taint**: Calling `Regex.Match(tainted).Value` or rebuilding a string character-by-character is still tracked as originating from an untrusted source.
* **Only Parameterization or Server Constants Clear Taint**: Snyk's rule engine only recognizes typed parameters (`SqlParameter`, Dapper parameters) or hardcoded server-side constants as valid sanitizers.

---

## 2. Remediation Patterns

### Pattern A: Standard Parameterized Queries

#### ❌ Vulnerable (String Interpolation / Concatenation):
```csharp
string sql = $"SELECT * FROM Orders WHERE CustomerId = '{customerId}' AND Status = '{status}'";
return await sqlHelper.ExecuteDataSetAsync(CommandType.Text, sql);
```

#### ✅ Secure (SqlParameter Collection):
```csharp
string sql = "SELECT * FROM Orders WHERE CustomerId = @CustomerId AND Status = @Status";
var parameters = new[]
{
    new SqlParameter("@CustomerId", SqlDbType.VarChar, 50) { Value = customerId },
    new SqlParameter("@Status", SqlDbType.VarChar, 20) { Value = status }
};
return await sqlHelper.ExecuteDataSetAsync(CommandType.Text, sql, parameters);
```

---

### Pattern B: Dynamic Stored Procedure / Table Names (Allowlist Mapping)

When an API accepts a stored procedure name or table name from an HTTP request, `SqlParameter` cannot be used for the procedure name itself. Passing the raw string directly into `CommandText` will trigger Snyk.

#### ❌ Vulnerable (Passing User String Directly):
```csharp
public async Task<IActionResult> ExecuteProc([FromBody] SqlModel model)
{
    // Regex or character validation still leaves model.CommandText tainted in Snyk:
    if (!Regex.IsMatch(model.CommandText, @"^[\w\.]+$")) return BadRequest();
    return Ok(await _sqlHelper.ExecuteNonQueryAsync(CommandType.StoredProcedure, model.CommandText));
}
```

#### ✅ Secure (Server-Side Allowlist Dictionary):
```csharp
private static readonly Dictionary<string, string> PermittedProcedures = new(StringComparer.OrdinalIgnoreCase)
{
    { "GetOrderSummary", "dbo.usp_GetOrderSummary" },
    { "UpdateOrderStatus", "dbo.usp_UpdateOrderStatus" },
    { "CalculateTotal", "dbo.usp_CalculateTotal" }
};

public async Task<IActionResult> ExecuteProc([FromBody] SqlModel model)
{
    if (string.IsNullOrWhiteSpace(model.CommandText) ||
        !PermittedProcedures.TryGetValue(model.CommandText, out var canonicalProcName))
    {
        return BadRequest("Unknown or unauthorized stored procedure name.");
    }

    // Pass the safe server-side dictionary constant to the SQL sink:
    return Ok(await _sqlHelper.ExecuteNonQueryAsync(CommandType.StoredProcedure, canonicalProcName));
}
```

---

### Pattern C: Dynamic `IN (...)` List Clauses

#### ❌ Vulnerable (`string.Join` in SQL):
```csharp
string sql = string.Format("SELECT * FROM Insertions WHERE Id IN ({0})", string.Join(",", ids));
```

#### ✅ Secure (Generated Parameter Placeholders):
```csharp
var idList = ids.ToList();
if (idList.Count == 0) return Enumerable.Empty<Insertion>();

var paramNames = idList.Select((_, i) => $"@Id{i}").ToList();
string sql = $"SELECT * FROM Insertions WHERE Id IN ({string.Join(",", paramNames)})";

var parameters = new List<SqlParameter>();
for (int i = 0; i < idList.Count; i++)
{
    parameters.Add(new SqlParameter($"@Id{i}", SqlDbType.Int) { Value = idList[i] });
}

return await _sqlHelper.ExecuteDataSetAsync(CommandType.Text, sql, parameters.ToArray());
```

---

### Pattern D: Dapper / Entity Framework Core

#### ❌ Vulnerable:
```csharp
// EF Core raw string
var users = await context.Users.FromSqlRaw($"SELECT * FROM Users WHERE Email = '{email}'").ToListAsync();

// Dapper string interpolation
var users = await connection.QueryAsync<User>($"SELECT * FROM Users WHERE Email = '{email}'");
```

#### ✅ Secure:
```csharp
// EF Core interpolated parameterization (safe)
var users = await context.Users.FromSqlInterpolated($"SELECT * FROM Users WHERE Email = {email}").ToListAsync();

// Dapper parameter object (safe)
var users = await connection.QueryAsync<User>("SELECT * FROM Users WHERE Email = @Email", new { Email = email });
```
