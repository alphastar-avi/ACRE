package prompt

import (
	"fmt"
	"strings"

	"acre/ticket"
)

// Generate constructs a remediation prompt for Codex based on the ticket details.
func Generate(t *ticket.Ticket, repoPath string) string {
	var builder strings.Builder

	builder.WriteString("You are an expert software engineer tasked with fixing a bug in the following repository:\n")
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

	builder.WriteString("## Instructions\n")
	builder.WriteString("1. Investigate the issue based on the provided incident report.\n")
	builder.WriteString("2. Determine the probable root cause in the target repository.\n")
	builder.WriteString("3. Identify the affected files.\n")
	builder.WriteString("4. Modify the code to fix the issue.\n")
	builder.WriteString("5. Create or update tests if necessary to prevent regressions.\n")
	builder.WriteString("6. Minimize unrelated modifications.\n")
	builder.WriteString("7. Focus only on resolving the reported incident.\n")
	builder.WriteString("Please execute the necessary changes using the available workspace-write capabilities.\n")

	return builder.String()
}
