package snyk

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// Issue represents a single Snyk vulnerability finding.
type Issue struct {
	FindingID  string `json:"finding_id"`
	Severity   string `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Title      string `json:"title"`
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Info       string `json:"info"`
	RawText    string `json:"raw_text"`
}

// SnykReport represents the parsed summary and findings from a Snyk code test.
type SnykReport struct {
	ProjectPath string         `json:"project_path"`
	TotalIssues int            `json:"total_issues"`
	OpenIssues  int            `json:"open_issues"`
	Counts      map[string]int `json:"counts"` // Counts by severity: HIGH, MEDIUM, LOW, CRITICAL
	Issues      []Issue        `json:"issues"`
	RawOutput   string         `json:"raw_output"`
}

// RunTest executes 'snyk code test "<targetPath>"' and returns the output and parsed report.
func RunTest(targetPath string) (string, *SnykReport, error) {
	snykExec, err := exec.LookPath("snyk")
	if err != nil {
		// Fallback check on standard windows/unix paths if exec.LookPath doesn't find it in subshell
		snykExec = "snyk"
	}

	cmd := exec.Command(snykExec, "code", "test", targetPath)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	output := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		output += "\n" + stderrBuf.String()
	}

	// Parse the output regardless of exit code (Snyk exits with non-zero when vulnerabilities are found)
	report := ParseOutput(output)
	report.ProjectPath = targetPath

	return output, report, nil
}
