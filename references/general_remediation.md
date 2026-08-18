# General SAST Security Remediation Guide

Universal guidelines for resolving common security vulnerabilities identified by Snyk and other SAST analyzers across multiple programming languages.

---

## 1. Path Traversal (`CWE-22`, `*/PT`, `*/PathTraversal`)

### Root Cause
User input is concatenated into file system paths (`File.ReadAllBytes(path)`, `os.Open(path)`, `fs.readFile(path)`), allowing `../` sequences to access arbitrary files on the server.

### Secure Remediation Pattern:
1. **Filename Only**: If the user only specifies a file name, strip directory separators:
   ```csharp
   string safeFilename = Path.GetFileName(userInput);
   string fullPath = Path.Combine(baseDirectory, safeFilename);
   ```
2. **Canonical Containment Check**: If subdirectories are allowed, verify that the canonical resolved path starts with the base directory:
   ```csharp
   string fullPath = Path.GetFullPath(Path.Combine(baseDirectory, userInput));
   if (!fullPath.StartsWith(Path.GetFullPath(baseDirectory) + Path.DirectorySeparatorChar))
   {
       throw new UnauthorizedAccessException("Path traversal detected.");
   }
   ```

---

## 2. Hardcoded Credentials / Secrets (`CWE-798`, `*/HardcodedCredentials`, `*/Secret`)

### Root Cause
Passwords, API keys, private keys, or connection strings are declared as literal strings in source code.

### Secure Remediation Pattern:
1. Read secrets from environment variables or configuration providers:
   ```csharp
   string apiKey = Environment.GetEnvironmentVariable("PAYMENT_API_KEY") 
       ?? Configuration["Payment:ApiKey"] 
       ?? throw new InvalidOperationException("API key not configured.");
   ```
2. Replace hardcoded test credentials in sample files with standard placeholder tokens (`"REDACTED"` or `"YOUR_API_KEY"`).

---

## 3. Cross-Site Request Forgery (`CWE-352`, `*/Csrf`, `*/CSRF`)

### Root Cause
State-modifying HTTP endpoints (`POST`, `PUT`, `DELETE`) accept requests without validating anti-forgery tokens.

### Secure Remediation Pattern:
1. In ASP.NET Core: Add `[ValidateAntiForgeryToken]` or `[AutoValidateAntiforgeryToken]` attributes on controllers/actions.
2. In Express.js: Apply `csurf` or double-submit cookie middleware.
3. In Spring Boot: Ensure `http.csrf()` protection is enabled for non-API web sessions.

---

## 4. Server-Side Request Forgery (`CWE-918`, `*/Ssrf`, `*/SSRF`)

### Root Cause
User input controls the target destination URL in backend HTTP requests (`HttpClient.GetAsync(url)`, `fetch(url)`).

### Secure Remediation Pattern:
1. **Allowlist Domains**: Validate the URI against an explicit allowlist of authorized hostnames:
   ```csharp
   if (!Uri.TryCreate(inputUrl, UriKind.Absolute, out var uri) || 
       !AllowedHosts.Contains(uri.Host.ToLowerInvariant()))
   {
       throw new ArgumentException("Unauthorized target host.");
   }
   ```
2. **Block Internal IP Ranges**: Reject requests targeting loopback (`127.0.0.1`, `localhost`) or private IP ranges (`10.0.0.0/8`, `192.168.0.0/16`, `172.16.0.0/12`, `169.254.169.254`).

---

## 5. Command Injection (`CWE-78`, `*/CommandInjection`)

### Root Cause
User input is passed directly to system shells (`Process.Start("cmd.exe /c " + input)`, `exec(input)`).

### Secure Remediation Pattern:
1. **Avoid Shell Execution**: Run the executable directly without invoking a system shell.
2. **Use Structured Argument Arrays**: Pass arguments as separate array elements rather than concatenating into a single command line string:
   ```csharp
   var startInfo = new ProcessStartInfo
   {
       FileName = "git",
       ArgumentList = { "log", "-n", "1", "--oneline" },
       RedirectStandardOutput = true
   };
   ```
