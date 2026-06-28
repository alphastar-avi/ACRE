package prompt

import (
	"fmt"
	"strings"

	"acre/ticket"
)

// Generate constructs a remediation prompt for OpenCode based on the ticket details.
func Generate(t *ticket.Ticket, repoPath string) string {
	var builder strings.Builder

	builder.WriteString("You are a senior software engineer tasked with fixing a bug in the following repository:\n")
	builder.WriteString(fmt.Sprintf("Repository Path: %s\n\n", repoPath))

	builder.WriteString("## Incident Report\n")
	builder.WriteString(fmt.Sprintf("**Ticket ID:** %s\n", t.TicketID))
	builder.WriteString(fmt.Sprintf("**Summary:** %s\n\n", t.Summary))

	builder.WriteString("### Description\n")
	builder.WriteString(t.Description + "\n\n")

	builder.WriteString("### Expected Behavior\n")
	builder.WriteString(t.ExpectedBehavior + "\n\n")

	builder.WriteString("### Actual Behavior\n")
	builder.WriteString(t.ActualBehavior + "\n\n")

	builder.WriteString("### Steps to Reproduce\n")
	for i, step := range t.StepsToReproduce {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}
	builder.WriteString("\n")

	if t.AdditionalNotes != "" {
		builder.WriteString("### Additional Notes\n")
		builder.WriteString(t.AdditionalNotes + "\n\n")
	}

	builder.WriteString("## Engineering & Implementation Guidelines\n")
	builder.WriteString("1. **Codebase Style Alignment**: Carefully read and mirror the patterns, indentation (spaces vs tabs), brackets, naming conventions, and programming paradigms already present in the codebase. Change only what is strictly necessary.\n")
	builder.WriteString("2. **Regression Testing**: If you modify any logic, locate the corresponding test files. Add or update unit/regression tests in the exact style of the existing test files. Run the test suite within the codebase to ensure nothing is broken.\n")
	builder.WriteString("3. **Senior Engineering Decision-Making**: If running the test suite reveals failures, analyze them carefully. Fix the failures by adjusting *only* the new changes you introduced. Do not modify unrelated, stable parts of the codebase to hide compile or test failures.\n")
	builder.WriteString("4. **No Hallucinations**: If you cannot locate the files related to the issue, cannot determine a safe way to fix the issue, or find that the issue is already resolved, do not generate fake code or touch unrelated files. Instead, proceed to the reporting step below and indicate that the issue could not be resolved.\n")
	builder.WriteString("5. **Backend Focus & Diagnostic Scoping**: 90% of the time, the root cause of these reported issues resides in the backend/server-side logic, controllers, configurations, models, routing, or data parsing. Do not make hypothetical or cosmetic frontend view/HTML/CSS changes unless it is explicitly clear the issue originates there. Focus your diagnostic investigations and modifications on backend business logic and services.\n")
	builder.WriteString("6. **Minimizing Test Modifications**: Do not modify or create new test files if the existing test suite is already sufficient to catch regressions, or if no new test coverage is explicitly requested. Only modify test files if you are adding new functionality that requires new test cases, or if existing test assertions are outdated. Avoid changing tests unnecessarily.\n")
	builder.WriteString("7. **Mandatory Reporting File**: Once you are finished, you MUST create a JSON file named `remediation_details.json` at the root of the repository. Do not leave the workspace without writing this file. It must have the following structure:\n")
	builder.WriteString("```json\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"understood_issue\": \"Detailed explanation of what you understood the issue to be\",\n")
	builder.WriteString("  \"potential_issue\": \"What you identified as the core root cause of the issue\",\n")
	builder.WriteString("  \"approach\": \"Detailed explanation of the approach used to fix the issue (or attempt to resolve it)\",\n")
	builder.WriteString("  \"code_changes\": [\n")
	builder.WriteString("    {\n")
	builder.WriteString("      \"file\": \"relative/path/to/modified/file.cs\",\n")
	builder.WriteString("      \"description\": \"Detailed description of modifications made to this file\"\n")
	builder.WriteString("    }\n")
	builder.WriteString("  ],\n")
	builder.WriteString("  \"recommendations\": \"Clear recommendations for manual engineering intervention if you were unable to solve the issue\",\n")
	builder.WriteString("  \"solved\": true, // set to false if you could not solve or safely fix the issue\n")
	builder.WriteString("  \"wrote_tests\": true // set to true if you created or modified test cases\n")
	builder.WriteString("}\n")
	builder.WriteString("```\n")

	return builder.String()
}
