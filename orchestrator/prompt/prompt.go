package prompt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acre/snyk"
	"acre/ticket"
)

// Generate constructs a remediation prompt for OpenCode based on the ticket details.
func Generate(t *ticket.Ticket, repoPath string, enableRecs bool) string {
	var builder strings.Builder

	builder.WriteString("You are a senior software engineer tasked with fixing a bug in the following repository:\n")
	builder.WriteString(fmt.Sprintf("Repository Path: %s\n\n", repoPath))

	// Look for OKF codebase index context (directory conforming to OKF v0.1 or legacy file)
	repoName := filepath.Base(repoPath)
	okfDirPaths := []string{
		filepath.Join("OKF", repoName),
		filepath.Join("..", "OKF", repoName),
	}
	var okfAbsPath string
	var indexContent string
	var foundOKF bool

	for _, dirPath := range okfDirPaths {
		info, err := os.Stat(dirPath)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(dirPath)
			if err == nil {
				okfAbsPath = abs
				indexPath := filepath.Join(dirPath, "index.md")
				data, readErr := os.ReadFile(indexPath)
				if readErr == nil {
					indexContent = string(data)
					foundOKF = true
					break
				}
			}
		}
	}

	if foundOKF {
		builder.WriteString("## Codebase Context & Index (Open Knowledge Format)\n")
		builder.WriteString("This codebase uses the Open Knowledge Format (OKF) v0.1 to manage architectural and domain knowledge.\n")
		builder.WriteString(fmt.Sprintf("The absolute path to the OKF documentation folder on this system is: %s\n\n", okfAbsPath))
		builder.WriteString("### Root index.md Document:\n")
		builder.WriteString("```markdown\n")
		builder.WriteString(indexContent)
		builder.WriteString("\n```\n\n")
		builder.WriteString("### 📖 OKF Progressive Disclosure Guidelines:\n")
		builder.WriteString("To ensure efficiency, minimize token cost, and prevent context lag, use the following strategy to discover and read documentation:\n")
		builder.WriteString("1. **Analyze the Root Index**: Start by reviewing the `Navigation Graph` and `Key Entry Points` in the `index.md` above to identify which documentation concept files might be relevant to the bug.\n")
		builder.WriteString("2. **Inspect Full YAML Metadata Block First**: The documentation concept files are located in the directory path provided above. Each file begins with a YAML frontmatter metadata block delimited by `---` containing `type`, `title`, `description`, `resource` and `tags`.\n")
		builder.WriteString("   Before reading an entire document body, read the full YAML frontmatter block (between opening and closing `---`) at the top of candidate files (e.g., `core_flow.md` or `testing.md`) to verify its target domain.\n")
		builder.WriteString("3. **Disclose on Demand**: If and only if the metadata confirms the file is highly relevant to the problem (e.g. describes the flow or layers where the bug occurred, or contains specific build/test instructions), proceed to read the rest of the file. Otherwise, skip it to keep the context clean.\n\n")
	} else {
		// Fallback to legacy single file OKF if directory/index.md isn't found
		okfPaths := []string{
			filepath.Join("OKF", repoName+".md"),
			filepath.Join("..", "OKF", repoName+".md"),
		}
		for _, p := range okfPaths {
			if data, err := os.ReadFile(p); err == nil {
				builder.WriteString("## Codebase Context & Index (Open Knowledge Format)\n")
				builder.WriteString(string(data))
				builder.WriteString("\n")
				break
			}
		}
	}

	builder.WriteString("## Incident Report\n")
	builder.WriteString(fmt.Sprintf("**Ticket ID:** %s\n", t.TicketID))
	builder.WriteString(fmt.Sprintf("**Summary:** %s\n\n", t.Summary))

	builder.WriteString("### Description\n")
	builder.WriteString(t.Description + "\n\n")

	if t.AcceptanceCriteria != "" {
		builder.WriteString("### Acceptance Criteria\n")
		builder.WriteString(t.AcceptanceCriteria + "\n\n")
	}

	if len(t.Comments) > 0 {
		builder.WriteString("### Comments & Discussion (Oldest to Newest)\n")
		for _, comment := range t.Comments {
			builder.WriteString(fmt.Sprintf("- **%s** (%s):\n  %s\n", comment.Author, comment.Created, comment.Body))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## Engineering & Implementation Guidelines\n")
	if enableRecs {
		builder.WriteString("1. **RECOMMENDATIONS-ONLY Mode**: Do NOT modify or edit any source files in the repository. Your sole task is to analyze the codebase, diagnose the root cause, design a potential fix, and write the details to `remediation_details.json`.\n")
		builder.WriteString("2. **Detailed Analysis**: In your JSON report, describe in detail the files that *should* be changed and the specific edits needed under the `code_changes` list.\n")
	} else {
		builder.WriteString("1. **Codebase Style Alignment**: Carefully read and mirror the patterns, indentation (spaces vs tabs), brackets, naming conventions, and programming paradigms already present in the codebase. Change only what is strictly necessary.\n")
		builder.WriteString("2. **Regression Testing**: If you modify any logic, locate the corresponding test files. Add or update unit/regression tests in the exact style of the existing test files. Run the test suite within the codebase to ensure nothing is broken.\n")
	}
	builder.WriteString("3. **Senior Engineering Decision-Making**: Analyze compilation/test patterns carefully. Do not introduce hypothetical or cosmetic frontend/HTML/CSS changes unless it is explicitly clear the issue originates there. Focus on backend business logic and services.\n")
	builder.WriteString("4. **No Hallucinations**: If you cannot locate the files related to the issue, cannot determine a safe way to fix the issue, or find that the issue is already resolved, indicate that in the report details.\n")
	builder.WriteString("5. **Confidence Rating**: You must assess your diagnosis and potential fix with a confidence score (from 0 to 100) and provide a short justification in the report.\n")
	builder.WriteString("6. **Mandatory Reporting File**: Once you are finished, you MUST create a JSON file named `remediation_details.json` at the root of the repository. Do not leave the workspace without writing this file. It must have the following structure:\n")
	builder.WriteString("```json\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"understood_issue\": \"Detailed explanation of what you understood the issue to be\",\n")
	builder.WriteString("  \"potential_issue\": \"What you identified as the core root cause of the issue\",\n")
	builder.WriteString("  \"approach\": \"Detailed explanation of the approach used to fix the issue (or attempt to resolve it)\",\n")
	builder.WriteString("  \"code_changes\": [\n")
	builder.WriteString("    {\n")
	builder.WriteString("      \"file\": \"relative/path/to/modified/file.cs\",\n")
	builder.WriteString("      \"description\": \"Detailed description of modifications made or needed for this file\"\n")
	builder.WriteString("    }\n")
	builder.WriteString("  ],\n")
	builder.WriteString("  \"recommendations\": \"Clear recommendations for manual engineering intervention if you were unable to solve the issue\",\n")
	builder.WriteString("  \"confidence_score\": 90, // integer percentage representing your confidence in the diagnosis and fix (0 to 100)\n")
	builder.WriteString("  \"confidence_justification\": \"A short, concise justification for your confidence score\",\n")
	builder.WriteString("  \"solved\": true, // set to false if you could not solve or safely fix the issue\n")
	builder.WriteString("  \"wrote_tests\": true // set to true if you created or modified test cases\n")
	builder.WriteString("}\n")
	builder.WriteString("```\n")

	return builder.String()
}

// GenerateSnykPrompt constructs a clear, robust remediation prompt for OpenCode to resolve targeted Snyk vulnerabilities.
func GenerateSnykPrompt(findings []snyk.NormalizedFinding, repoPath string, refPath string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Fix the %d normalized Snyk security vulnerability finding(s) listed below in the target repository at `%s`.\n\n", len(findings), repoPath))
	builder.WriteString("AUTOMATED NON-INTERACTIVE DIRECTIVE:\n")
	builder.WriteString("- You are executing in fully automated non-interactive remediation mode inside ACRE.\n")
	builder.WriteString("- DO NOT ask questions, request user confirmation, or ask for Snyk reports or tech stack clarification.\n")
	builder.WriteString("- ALL targeted Snyk findings (findingId, ruleId, title, level, message, file, line, cwe, precision, codeFlow) are explicitly provided below in JSON format.\n")
	builder.WriteString("- IMMEDIATELY inspect the repository files, trace data flows via `codeFlow` when available, edit the source code to remediate the findings, verify project compilation/build, re-verify with Snyk code test for the targeted ruleId(s), write `remediation_details.json`, and exit!\n\n")

	builder.WriteString(fmt.Sprintf("Repository Path: %s\n\n", repoPath))

	// Include reference guidance if provided via --REF
	if refPath != "" {
		info, err := os.Stat(refPath)
		if err == nil {
			builder.WriteString("## Reference & Standard Recommended Instructions\n")
			builder.WriteString(fmt.Sprintf("Reference Location: %s\n", refPath))
			if !info.IsDir() {
				if content, rErr := os.ReadFile(refPath); rErr == nil {
					builder.WriteString("```markdown\n")
					builder.WriteString(string(content))
					builder.WriteString("\n```\n\n")
				}
			} else {
				// Search for .md files in the reference directory
				entries, _ := os.ReadDir(refPath)
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
						p := filepath.Join(refPath, entry.Name())
						if content, rErr := os.ReadFile(p); rErr == nil {
							builder.WriteString(fmt.Sprintf("### Reference File: %s\n", entry.Name()))
							builder.WriteString("```markdown\n")
							builder.WriteString(string(content))
							builder.WriteString("\n```\n\n")
						}
					}
				}
			}
		}
	}

	builder.WriteString("## Targeted Snyk Normalized Findings\n")
	builder.WriteString("Below is the exact normalized JSON finding structure passed for remediation:\n\n")

	normBytes, err := json.MarshalIndent(findings, "", "  ")
	if err == nil {
		builder.WriteString("```json\n")
		builder.WriteString(string(normBytes))
		builder.WriteString("\n```\n\n")
	} else {
		for idx, f := range findings {
			builder.WriteString(fmt.Sprintf("%d. findingId: %s | ruleId: %s | title: %s | file: %s | line: %d | level: %s | message: %s\n",
				idx+1, f.FindingID, f.RuleID, f.Title, f.File, f.Line, f.Level, f.Message))
		}
		builder.WriteString("\n")
	}

	acreExe := getAcreExecutablePath()

	builder.WriteString("## Execution, Build & Verification Workflow\n")
	builder.WriteString("1. **Targeted Remediation & Taint Data-Flow Analysis (`codeFlow`)**:\n")
	builder.WriteString("   - For each finding, analyze the provided `ruleId`, `title`, `cwe` classification, and `message` to determine the precise vulnerability category and appropriate security remediation pattern.\n")
	builder.WriteString("   - Adhere to industry-standard secure coding guidelines for the target language and framework.\n")
	builder.WriteString("   - **Snyk Static Taint Analysis Rules (Breaking Taint Chains)**:\n")
	builder.WriteString("     * Snyk uses AST data-flow tracking from source to sink. Custom runtime string checks, regex validation on raw inputs, character-by-character string rebuilding, or custom quote-escaping functions do NOT break static taint chains in SAST engines.\n")
	builder.WriteString("     * **Command & Query Sinks (Injection Flaws)**: ALWAYS replace dynamic string concatenation or interpolation with native parameterized queries / prepared statements. Never pass concatenated untrusted strings into execution sinks.\n")
	builder.WriteString("     * **Dynamic Identifiers (Procedure, Table, Column Names)**: If sink requires a dynamic identifier, map inputs through a fixed server-side constant allowlist or lookup map so the value passed to the sink is a safe constant from your code, not user input.\n")
	builder.WriteString("     * **Path & File Sinks (Path Traversal)**: Use strict allowlists, base filename extraction, or verify canonical paths start with the target base directory.\n")
	builder.WriteString("     * **Type Constraints**: Parse inputs into strong primitive types (integers, UUIDs/GUIDs, booleans, enums) before using them.\n")
	builder.WriteString("   - If `codeFlow` is provided, trace the sequential path of data/taint from the source (where untrusted data enters) through intermediary helper classes/methods to the vulnerable sink (where it is evaluated or executed). Apply the fix at the most effective and secure point along the flow (e.g. source input validation, intermediary type constraints, or sink parameterization).\n")
	builder.WriteString("   - If `codeFlow` is empty, inspect the target `file` at `line` and surrounding context.\n")
	builder.WriteString("   - Apply minimal, robust security fixes matching the `ruleId`, `title`, and `message` while preserving existing business logic and code style.\n\n")
	builder.WriteString("2. **Module Build & Compilation Verification**:\n")
	builder.WriteString("   - Analyze the stack of the target repository and detect the native solution build command (e.g. `dotnet build` for C#/.NET, `npm run build` or `npx tsc` for TS/JS, `go build` for Go, `mvn compile` or `gradle build` for Java).\n")
	builder.WriteString("   - Execute the build command to ensure your code changes compile cleanly without errors.\n")
	builder.WriteString("   - Ensure your modified files introduce ZERO syntax, type, or compiler errors (e.g. `error CS...` in C#, TypeScript compilation errors, Go syntax errors).\n")
	builder.WriteString("   - If the repository has pre-existing environment issues (such as missing external system SDKs or pre-existing package restore mismatches in unrelated modules) that are completely unrelated to your edits, verify that your modified files compile with 0 syntax/type errors and proceed.\n\n")
	builder.WriteString("3. **Snyk Targeted Verification Scan**:\n")
	builder.WriteString("   - To verify if the vulnerability for a target `ruleId` has been eliminated, execute the ACRE CLI verification command from your shell:\n")
	builder.WriteString(fmt.Sprintf("     `\"%s\" --snykjson --repo . --ruleid <ruleId> --cli`\n", acreExe))
	builder.WriteString(fmt.Sprintf("     *(PowerShell note: If invoking a quoted path in PowerShell, prefix with `&`: `& \"%s\" --snykjson --repo . --ruleid <ruleId> --cli`)*\n", acreExe))
	builder.WriteString("   - The `--cli` flag runs the Snyk scan on the target repository, normalizes the output, and prints the remaining JSON findings directly to terminal output.\n")
	builder.WriteString(fmt.Sprintf("   - Continue iterating code fixes and build verification until the command returns `[]` (0 remaining findings for that `ruleId`).\n"))
	builder.WriteString("   - *(Fallback: You can also run `snyk code test --json` directly in the target repository to inspect remaining SARIF findings matching the target `ruleId` or `findingId`).* \n\n")
	builder.WriteString("4. **Anti-Hallucination Rule**:\n")
	builder.WriteString("   - If you cannot safely resolve a vulnerability without breaking critical functionality, DO NOT hallucinate fixes or make dummy edits.\n")
	builder.WriteString("   - Set `\"solved\": false` in `remediation_details.json`, explain the details under `\"recommendations\"`, and exit.\n\n")
	builder.WriteString("5. **Mandatory Reporting File (`remediation_details.json`)**:\n")
	builder.WriteString("   - Create or update `remediation_details.json` at the root of the repository before completing execution. Schema:\n")
	builder.WriteString("```json\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"understood_issue\": \"Summary of Snyk security findings addressed\",\n")
	builder.WriteString("  \"potential_issue\": \"Root cause analysis explaining why the vulnerabilities existed\",\n")
	builder.WriteString("  \"approach\": \"Detailed explanation of exact security fixes applied\",\n")
	builder.WriteString("  \"code_changes\": [\n")
	builder.WriteString("    {\n")
	builder.WriteString("      \"file\": \"relative/path/to/modified/file.cs\",\n")
	builder.WriteString("      \"description\": \"Detailed explanation of security changes made in this file\"\n")
	builder.WriteString("    }\n")
	builder.WriteString("  ],\n")
	builder.WriteString("  \"confidence_score\": 95,\n")
	builder.WriteString("  \"confidence_justification\": \"Justification for confidence in fix\",\n")
	builder.WriteString("  \"solved\": true,\n")
	builder.WriteString("  \"wrote_tests\": false\n")
	builder.WriteString("}\n")
	builder.WriteString("```\n")

	return builder.String()
}

func getAcreExecutablePath() string {
	exePath, err := os.Executable()
	if err == nil && exePath != "" {
		exePath = filepath.Clean(exePath)
		if !strings.Contains(exePath, "go-build") && !strings.Contains(exePath, "/tmp/") && !strings.Contains(exePath, "\\Temp\\") {
			return exePath
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for _, name := range []string{"acre.exe", "acre"} {
			candidate := filepath.Join(cwd, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}
	return "acre"
}

