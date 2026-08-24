package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"acre/ticket"
)

func TestGenerate_AgentDrivenBuildAndTestReport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "acre_report_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	data := Data{
		Ticket: &ticket.Ticket{
			TicketID: "ENG-9999",
			Summary:  "Test Summary for Incident",
		},
		Prompt:         "Test prompt text",
		OpenCodeOutput: "Test OpenCode output with build and tests passed",
		RepositoryPath: "/workspace/repo",
		Details: &RemediationDetails{
			UnderstoodIssue:         "Understood issue description",
			PotentialIssue:          "Root cause identified",
			Approach:                "Approach applied",
			CodeChanges:             []CodeChange{{File: "src/App.cs", Description: "Fixed null check"}},
			Recommendations:         "No extra recommendations",
			ConfidenceScore:         95,
			ConfidenceJustification: "All unit tests pass",
			Solved:                  true,
			WroteTests:              true,
		},
		HasDetails:     true,
		PullRequestURL: "https://github.com/org/repo/pull/1",
	}

	runDir, err := Generate(tempDir, data)
	if err != nil {
		t.Fatalf("Report generation failed: %v", err)
	}

	// Verify report file was created
	reportFile := filepath.Join(runDir, "remediation_report.md")
	content, err := os.ReadFile(reportFile)
	if err != nil {
		t.Fatalf("Failed to read generated report file: %v", err)
	}

	reportStr := string(content)
	if !strings.Contains(reportStr, "ENG-9999") {
		t.Errorf("Expected report to contain Ticket ID")
	}
	if !strings.Contains(reportStr, "Auto-detected & verified by OpenCode Agent") {
		t.Errorf("Expected report to contain default agent-driven build verification")
	}
	if !strings.Contains(reportStr, "Auto-detected targeted & regression tests verified by OpenCode Agent") {
		t.Errorf("Expected report to contain default agent-driven test verification")
	}
	if !strings.Contains(reportStr, "* **Outcome:** Success") {
		t.Errorf("Expected report outcome to be Success")
	}

	// Verify build.log and test.log exist
	buildLog, err := os.ReadFile(filepath.Join(runDir, "build.log"))
	if err != nil || !strings.Contains(string(buildLog), "OpenCode Agent") {
		t.Errorf("Expected build.log to note agent execution")
	}

	testLog, err := os.ReadFile(filepath.Join(runDir, "test.log"))
	if err != nil || !strings.Contains(string(testLog), "OpenCode Agent") {
		t.Errorf("Expected test.log to note agent execution")
	}
}
