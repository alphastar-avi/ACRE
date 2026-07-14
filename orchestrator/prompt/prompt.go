package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		builder.WriteString("2. **Inspect YAML Metadata First**: The documentation concept files are located in the directory path provided above. Each file begins with a YAML frontmatter metadata block containing `type`, `title`, `description`, `resource` and `tags`.\n")
		builder.WriteString("   Before reading an entire file, use your file reading tool to view only the first 10-15 lines of a candidate file to check its metadata block. For example, read the top lines of `basket_flow.md` or `testing.md` to see if it matches your target area.\n")
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
