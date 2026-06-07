package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"acre/ticket"
)

// Data holds all information needed to generate the report.
type Data struct {
	Ticket         *ticket.Ticket
	Prompt         string
	CodexOutput    string
	BuildExitCode  int
	BuildStdout    string
	BuildStderr    string
	TestExitCode   int
	TestStdout     string
	TestStderr     string
	RepositoryPath string
}

// Generate creates a timestamped run directory and saves all artifacts and the final report.
func Generate(baseRunsDir string, data Data) (string, error) {
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	runDir := filepath.Join(baseRunsDir, timestamp)

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create run directory: %w", err)
	}

	// Save the prompt
	if err := os.WriteFile(filepath.Join(runDir, "prompt.txt"), []byte(data.Prompt), 0644); err != nil {
		return "", err
	}

	// Save codex output
	if err := os.WriteFile(filepath.Join(runDir, "codex_output.log"), []byte(data.CodexOutput), 0644); err != nil {
		return "", err
	}

	// Save build logs
	buildLogs := fmt.Sprintf("Exit Code: %d\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n", data.BuildExitCode, data.BuildStdout, data.BuildStderr)
	if err := os.WriteFile(filepath.Join(runDir, "build.log"), []byte(buildLogs), 0644); err != nil {
		return "", err
	}

	// Save test logs
	testLogs := fmt.Sprintf("Exit Code: %d\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n", data.TestExitCode, data.TestStdout, data.TestStderr)
	if err := os.WriteFile(filepath.Join(runDir, "test.log"), []byte(testLogs), 0644); err != nil {
		return "", err
	}

	// Determine Outcome
	outcome := "Success"
	if data.BuildExitCode != 0 {
		outcome = "Failure (Build Failed)"
	} else if data.TestExitCode != 0 {
		outcome = "Failure (Tests Failed)"
	}

	// Generate Final Markdown Report
	reportContent := fmt.Sprintf(`# Remediation Report

## Ticket Information
* **Ticket ID:** %s
* **Summary:** %s

## Validation
* **Build Result:** %s
* **Test Result:** %s

## Outcome
* **%s**
`,
		data.Ticket.TicketID,
		data.Ticket.Summary,
		getResultString(data.BuildExitCode),
		getResultString(data.TestExitCode),
		outcome,
	)

	reportPath := filepath.Join(runDir, "remediation_report.md")
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return "", err
	}

	return runDir, nil
}

func getResultString(exitCode int) string {
	if exitCode == 0 {
		return "Success"
	}
	return "Failure"
}
