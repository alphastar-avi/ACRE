package snyk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// NormalizedFinding represents a normalized Snyk vulnerability finding.
type NormalizedFinding struct {
	RuleID    string   `json:"ruleId"`
	Title     string   `json:"title"`
	Level     string   `json:"level"`
	Message   string   `json:"message"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	CWE       []string `json:"cwe"`
	Precision string   `json:"precision"`
}

// SARIF 2.1.0 JSON Structures for Snyk output
type SarifRule struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ShortDescription struct {
		Text string `json:"text"`
	} `json:"shortDescription"`
	Properties struct {
		CWE       []string `json:"cwe"`
		Precision string   `json:"precision"`
	} `json:"properties"`
}

type SarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region struct {
			StartLine int `json:"startLine"`
		} `json:"region"`
	} `json:"physicalLocation"`
}

type SarifResult struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []SarifLocation `json:"locations"`
}

type SarifRun struct {
	Tool struct {
		Driver struct {
			Rules []SarifRule `json:"rules"`
		} `json:"driver"`
	} `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifReport struct {
	Runs []SarifRun `json:"runs"`
}

// Issue represents a single Snyk vulnerability finding (legacy compatibility).
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
	ProjectPath string              `json:"project_path"`
	TotalIssues int                 `json:"total_issues"`
	OpenIssues  int                 `json:"open_issues"`
	Counts      map[string]int      `json:"counts"` // Counts by severity
	Issues      []Issue             `json:"issues"`
	Normalized  []NormalizedFinding `json:"normalized"`
	RawOutput   string              `json:"raw_output"`
}

// RunTest executes 'snyk code test "<targetPath>"' and returns the output and parsed report.
func RunTest(targetPath string) (string, *SnykReport, error) {
	snykExec, err := exec.LookPath("snyk")
	if err != nil {
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

	report := ParseOutput(output)
	report.ProjectPath = targetPath

	return output, report, nil
}

// RunTestJSON executes 'snyk code test --json "<targetPath>"' and returns the JSON output.
func RunTestJSON(targetPath string) ([]byte, error) {
	snykExec, err := exec.LookPath("snyk")
	if err != nil {
		snykExec = "snyk"
	}

	cmd := exec.Command(snykExec, "code", "test", "--json", targetPath)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Snyk exits with non-zero when vulnerabilities are found, so ignore exit error
	_ = cmd.Run()

	out := stdoutBuf.Bytes()
	if len(out) == 0 && stderrBuf.Len() > 0 {
		out = stderrBuf.Bytes()
	}
	return out, nil
}

// NormalizeSarifJSON parses raw SARIF JSON output from Snyk and converts it into normalized findings.
func NormalizeSarifJSON(sarifBytes []byte, ruleIdFilter string) ([]NormalizedFinding, error) {
	var sarif SarifReport
	if err := json.Unmarshal(sarifBytes, &sarif); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SARIF JSON: %w", err)
	}

	var findings []NormalizedFinding

	for _, run := range sarif.Runs {
		// 1. Build lookup map from rules
		ruleMap := make(map[string]SarifRule)
		for _, rule := range run.Tool.Driver.Rules {
			ruleMap[rule.ID] = rule
		}

		// 2. Iterate over results
		for _, result := range run.Results {
			// Apply ruleId filter if specified
			if ruleIdFilter != "" && result.RuleID != ruleIdFilter {
				continue
			}

			file := ""
			line := 0
			if len(result.Locations) > 0 {
				file = result.Locations[0].PhysicalLocation.ArtifactLocation.URI
				line = result.Locations[0].PhysicalLocation.Region.StartLine
			}

			cwe := []string{}
			title := ""
			precision := ""

			if rule, found := ruleMap[result.RuleID]; found {
				title = rule.ShortDescription.Text
				if rule.Properties.CWE != nil {
					cwe = rule.Properties.CWE
				}
				precision = rule.Properties.Precision
			}

			findings = append(findings, NormalizedFinding{
				RuleID:    result.RuleID,
				Title:     title,
				Level:     result.Level,
				Message:   result.Message.Text,
				File:      file,
				Line:      line,
				CWE:       cwe,
				Precision: precision,
			})
		}
	}

	if findings == nil {
		findings = []NormalizedFinding{}
	}

	return findings, nil
}

// GenerateSnykJSON executes 'snyk code test --json', writes 'snykOutput.json',
// normalizes findings, creates 'snykOutput' directory if missing, writes 'Output<datetime>.json',
// and if printCLI is true, prints the normalized JSON directly to stdout for CLI consumption.
func GenerateSnykJSON(repoPath string, ruleIdFilter string, printCLI bool) (string, error) {
	if repoPath == "" {
		repoPath = "."
	}
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repo path: %w", err)
	}

	if !printCLI {
		fmt.Printf("[SNYKJSON] Running 'snyk code test --json' on target repository: %s...\n", absRepo)
	}

	jsonBytes, err := RunTestJSON(absRepo)
	if err != nil || len(jsonBytes) == 0 {
		return "", fmt.Errorf("failed to run 'snyk code test --json': %v", err)
	}

	// Save raw output to snykOutput.json in working directory
	if err := os.WriteFile("snykOutput.json", jsonBytes, 0644); err != nil && !printCLI {
		fmt.Printf("[SNYKJSON] Warning: Could not write raw snykOutput.json: %v\n", err)
	} else if !printCLI {
		fmt.Println("[SNYKJSON] Saved raw output to snykOutput.json")
	}

	// Normalize findings
	normalized, err := NormalizeSarifJSON(jsonBytes, ruleIdFilter)
	if err != nil {
		return "", fmt.Errorf("failed to normalize SARIF JSON: %w", err)
	}

	normBytes, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal normalized findings: %w", err)
	}

	if printCLI {
		// Output clean normalized JSON array directly to stdout
		fmt.Println(string(normBytes))
	}

	// Create snykOutput folder at root if it doesn't exist
	outputDir := "snykOutput"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Output<datetime>.json filename
	timestamp := time.Now().Format("20060102_150405")
	outputFilename := fmt.Sprintf("Output%s.json", timestamp)
	outputPath := filepath.Join(outputDir, outputFilename)

	if err := os.WriteFile(outputPath, normBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write normalized file %s: %w", outputPath, err)
	}

	if !printCLI {
		fmt.Printf("[SNYKJSON] Successfully created normalized JSON with %d finding(s) at: %s\n", len(normalized), outputPath)
	}
	return outputPath, nil
}

