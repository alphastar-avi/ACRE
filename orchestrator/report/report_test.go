package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(reportStr, "## Final Status") {
		t.Errorf("Expected report to contain Final Status section")
	}
	if !strings.Contains(reportStr, "* **Model Used:**") {
		t.Errorf("Expected report to contain Model Used")
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

func TestGenerate_DetailedFinalStatusWithTimingAndTokens(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "acre_timing_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	startTime := time.Date(2026, 8, 27, 1, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 8, 27, 1, 2, 30, 0, time.UTC)

	data := Data{
		Ticket: &ticket.Ticket{
			TicketID: "ENG-1000",
			Summary:  "Incident with timing and token tracking",
		},
		Prompt:         "Prompt data text",
		OpenCodeOutput: "OpenCode output logs\nTotal tokens: 12,450 (Prompt: 10,000 | Completion: 2,450)",
		RepositoryPath: "/workspace/repo",
		Details: &RemediationDetails{
			UnderstoodIssue: "Sample issue",
			Solved:          true,
		},
		HasDetails: true,
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   endTime.Sub(startTime),
		ModelUsed:  "opencode/claude-3-7-sonnet",
		TokenUsage: "12,450 tokens (Input: 10,000 | Output: 2,450)",
	}

	runDir, err := Generate(tempDir, data)
	if err != nil {
		t.Fatalf("Report generation failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(runDir, "remediation_report.md"))
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	reportStr := string(content)
	if !strings.Contains(reportStr, "* **Model Used:** `opencode/claude-3-7-sonnet`") {
		t.Errorf("Expected Model Used in report")
	}
	if !strings.Contains(reportStr, "* **Start Time:** 2026-08-27 01:00:00 UTC") {
		t.Errorf("Expected Start Time in report")
	}
	if !strings.Contains(reportStr, "* **End Time:** 2026-08-27 01:02:30 UTC") {
		t.Errorf("Expected End Time in report")
	}
	if !strings.Contains(reportStr, "* **Total Duration:** 2m 30s") {
		t.Errorf("Expected Total Duration in report")
	}
	if !strings.Contains(reportStr, "* **Token Usage:** 12,450 tokens") {
		t.Errorf("Expected Token Usage in report")
	}
}
