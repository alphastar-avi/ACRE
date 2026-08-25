package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"acre/opencode"
	"acre/prompt"
	"acre/report"
	"acre/snyk"
)

// RunSnyk executes the updated automated Snyk vulnerability remediation workflow.
func RunSnyk(snykJsonPath, repoPath, reportDir, refPath string) error {
	printSnykHeader()

	logBuffer := &strings.Builder{}
	opencodeLogBuffer := &strings.Builder{}

	logPrintln := func(format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		fmt.Println(msg)
		logBuffer.WriteString(stripANSI(msg) + "\n")
	}

	if snykJsonPath == "" {
		return fmt.Errorf("missing path to normalized Snyk JSON file (e.g. Output<datetime>.json)")
	}

	// 1. Read Normalized Findings JSON
	logPrintln("%s[STEP 1/4]%s %sLoading normalized Snyk JSON findings...%s", Cyan, Reset, Bold, Reset)
	logPrintln("   Normalized JSON Path: %s", snykJsonPath)

	data, err := os.ReadFile(snykJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read normalized Snyk JSON at %s: %w", snykJsonPath, err)
	}

	var findings []snyk.NormalizedFinding
	if err := json.Unmarshal(data, &findings); err != nil {
		return fmt.Errorf("failed to parse normalized Snyk JSON at %s: %w", snykJsonPath, err)
	}

	absRepo := repoPath
	if absRepo == "" {
		absRepo = "."
	}
	absRepoPath, err := filepath.Abs(absRepo)
	if err == nil {
		absRepo = absRepoPath
	}
	logPrintln("   Target Repository: %s", absRepo)

	logPrintln("   %s[LOADED]%s Total normalized finding(s) targeted: %d", Green, Reset, len(findings))
	for idx, f := range findings {
		logPrintln("   - Finding %d: ruleId=%s | title=%s | file=%s:%d", idx+1, f.RuleID, f.Title, f.File, f.Line)
	}
	logPrintln("")

	// 2. Reference Guidance Check (--REF)
	logPrintln("%s[STEP 2/4]%s %sChecking Reference (--REF) instructions...%s", Cyan, Reset, Bold, Reset)
	if refPath != "" {
		if _, statErr := os.Stat(refPath); statErr != nil {
			logPrintln("   %s[WARNING]%s Specified REF path not found: %s", Yellow, Reset, refPath)
		} else {
			logPrintln("   %s[REF LOADED]%s Found reference instructions at: %s", Green, Reset, refPath)
		}
	} else {
		logPrintln("   %s[INFO]%s No --REF reference path supplied.", Yellow, Reset)
	}
	logPrintln("")

	// 3. Construct Prompt & Execute OpenCode
	logPrintln("%s[STEP 3/4]%s %sExecuting OpenCode Remediation Agent...%s", Cyan, Reset, Bold, Reset)
	snykPrompt := prompt.GenerateSnykPrompt(findings, absRepo, refPath)

	logPrintln("   Invoking OpenCode agent...")
	opencodeOut, err := opencode.Run(snykPrompt, absRepo)
	opencodeLogBuffer.WriteString(opencodeOut + "\n")

	if err != nil {
		logPrintln("   %s[WARNING]%s OpenCode execution completed with warning/error: %v", Yellow, Reset, err)
	} else {
		logPrintln("   %s[RUN COMPLETE]%s OpenCode agent execution finished.", Green, Reset)
	}
	logPrintln("")

	// 4. Verification Scan
	logPrintln("%s[STEP 4/4]%s %sExecuting Verification Scan with Snyk Code Test...%s", Cyan, Reset, Bold, Reset)
	postScanBytes, verifyErr := snyk.RunTestJSON(absRepo)
	var remainingFindings []snyk.NormalizedFinding
	if verifyErr != nil {
		logPrintln("   %s[WARNING]%s Verification scan execution error: %v", Yellow, Reset, verifyErr)
	} else if len(postScanBytes) > 0 {
		remainingFindings, _ = snyk.NormalizeSarifJSON(postScanBytes, "")
	}

	// Calculate resolved targeted findings using robust findingId matching and fallback rules
	resolvedCount := 0
	var unresolvedFindings []snyk.NormalizedFinding
	matchedRemaining := make(map[int]bool)

	for _, target := range findings {
		found, matchIdx := matchTargetFinding(target, remainingFindings, matchedRemaining)
		if found {
			matchedRemaining[matchIdx] = true
			unresolvedFindings = append(unresolvedFindings, target)
		} else {
			resolvedCount++
		}
	}

	logPrintln("   %s[VERIFICATION COMPLETE]%s Target findings: %d | Resolved: %d | Unresolved: %d",
		Green, Reset, len(findings), resolvedCount, len(unresolvedFindings))
	logPrintln("")

	// Read remediation_details.json from repo
	detailsPath := filepath.Join(absRepo, "remediation_details.json")
	var details report.RemediationDetails
	hasDetails := false

	if detailsBytes, err := os.ReadFile(detailsPath); err == nil {
		if jsonErr := json.Unmarshal(detailsBytes, &details); jsonErr == nil {
			hasDetails = true
		}
		_ = os.Remove(detailsPath) // Clean up JSON after reading
	}

	var detailsPtr *report.RemediationDetails
	if hasDetails {
		detailsPtr = &details
	}

	return generateSnykReports(reportDir, absRepo, findings, unresolvedFindings, resolvedCount, detailsPtr, snykPrompt, opencodeLogBuffer.String(), logBuffer.String())
}

// matchTargetFinding matches a target finding against post-scan findings.
// It prioritizes findingId matching to ensure line shifts caused by code edits are not falsely treated as fixes.
func matchTargetFinding(target snyk.NormalizedFinding, remaining []snyk.NormalizedFinding, matched map[int]bool) (bool, int) {
	// 1. Primary match: Match by stable findingId
	if target.FindingID != "" {
		for i, post := range remaining {
			if !matched[i] && post.FindingID != "" && post.FindingID == target.FindingID {
				return true, i
			}
		}
	}

	targetFile := normalizeFilePath(target.File)

	// 2. Secondary match: Exact RuleID + File + Line
	for i, post := range remaining {
		if !matched[i] && post.RuleID == target.RuleID && normalizeFilePath(post.File) == targetFile {
			if post.Line == target.Line {
				return true, i
			}
		}
	}

	// 3. Fallback match: Same RuleID + File with line proximity (e.g. line shifted due to code edits)
	bestIdx := -1
	minLineDiff := 1000000
	for i, post := range remaining {
		if !matched[i] && post.RuleID == target.RuleID && normalizeFilePath(post.File) == targetFile {
			diff := post.Line - target.Line
			if diff < 0 {
				diff = -diff
			}
			if diff < minLineDiff {
				minLineDiff = diff
				bestIdx = i
			}
		}
	}

	if bestIdx != -1 {
		return true, bestIdx
	}

	return false, -1
}

func normalizeFilePath(p string) string {
	clean := strings.ReplaceAll(p, "\\", "/")
	clean = strings.TrimPrefix(clean, "./")
	return strings.ToLower(strings.TrimSpace(clean))
}

func generateSnykReports(reportDir string, repoPath string, initialFindings, unresolvedFindings []snyk.NormalizedFinding, resolvedCount int, details *report.RemediationDetails, promptOutput, opencodeOutput, logOutput string) error {
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	var reportBuilder strings.Builder
	reportBuilder.WriteString("# ACRE Snyk Code Test Remediation Report\n\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Date:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	reportBuilder.WriteString(fmt.Sprintf("* **Target Repository:** `%s`\n\n", repoPath))

	reportBuilder.WriteString("## Initial Vulnerability Findings Summary\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Targeted Findings Count:** %d\n\n", len(initialFindings)))
	if len(initialFindings) > 0 {
		reportBuilder.WriteString("| Finding ID | Rule ID | Title | File | Line | Level |\n")
		reportBuilder.WriteString("| ---------- | ------- | ----- | ---- | ---- | ----- |\n")
		for _, f := range initialFindings {
			fid := f.FindingID
			if fid == "" {
				fid = "-"
			}
			reportBuilder.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | `%s` | %d | %s |\n", fid, f.RuleID, f.Title, f.File, f.Line, f.Level))
		}
		reportBuilder.WriteString("\n")
	}

	reportBuilder.WriteString("## Verification Summary (After Remediation & Snyk Scan)\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Resolved Findings Count:** %d\n", resolvedCount))
	reportBuilder.WriteString(fmt.Sprintf("* **Remaining Unresolved Findings:** %d\n\n", len(unresolvedFindings)))

	if len(unresolvedFindings) > 0 {
		reportBuilder.WriteString("### Remaining Unresolved Findings\n")
		reportBuilder.WriteString("| Finding ID | Rule ID | Title | File | Line | Level |\n")
		reportBuilder.WriteString("| ---------- | ------- | ----- | ---- | ---- | ----- |\n")
		for _, f := range unresolvedFindings {
			fid := f.FindingID
			if fid == "" {
				fid = "-"
			}
			reportBuilder.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | `%s` | %d | %s |\n", fid, f.RuleID, f.Title, f.File, f.Line, f.Level))
		}
		reportBuilder.WriteString("\n")
	}

	reportBuilder.WriteString("## OpenCode Vulnerability Analysis & Fix Details\n")
	if details != nil {
		reportBuilder.WriteString(fmt.Sprintf("* **Confidence Score:** %d/100\n", details.ConfidenceScore))
		if details.ConfidenceJustification != "" {
			reportBuilder.WriteString(fmt.Sprintf("* **Confidence Justification:** %s\n", details.ConfidenceJustification))
		}
		reportBuilder.WriteString(fmt.Sprintf("* **Solved Status:** %t\n\n", details.Solved))

		if details.UnderstoodIssue != "" {
			reportBuilder.WriteString(fmt.Sprintf("### Understood Vulnerabilities\n%s\n\n", details.UnderstoodIssue))
		}
		if details.PotentialIssue != "" {
			reportBuilder.WriteString(fmt.Sprintf("### Core Root Cause Analysis\n%s\n\n", details.PotentialIssue))
		}
		if details.Approach != "" {
			reportBuilder.WriteString(fmt.Sprintf("### Applied Security Remediation Approach\n%s\n\n", details.Approach))
		}

		reportBuilder.WriteString("### Applied Code Modifications\n")
		if len(details.CodeChanges) > 0 {
			for _, cc := range details.CodeChanges {
				reportBuilder.WriteString(fmt.Sprintf("* **File:** `%s`\n  * **Modification:** %s\n", cc.File, cc.Description))
			}
			reportBuilder.WriteString("\n")
		} else {
			reportBuilder.WriteString("_No source code files were marked for modification._\n\n")
		}

		if details.Recommendations != "" {
			reportBuilder.WriteString(fmt.Sprintf("### Extra Engineering Recommendations\n%s\n\n", details.Recommendations))
		}
	} else {
		reportBuilder.WriteString("> [!WARNING]\n> OpenCode completed execution but did not write `remediation_details.json`.\n\n")
	}

	reportBuilder.WriteString("## Output Artifacts & Execution Logs\n")
	reportBuilder.WriteString("- System prompt fed to OpenCode is saved in `prompt.md`.\n")
	reportBuilder.WriteString("- Raw output from OpenCode execution is saved in `opencode_output.md`.\n")
	reportBuilder.WriteString("- CLI trace log is saved in `logs.md`.\n")

	_ = os.WriteFile(filepath.Join(reportDir, "report.md"), []byte(reportBuilder.String()), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "prompt.md"), []byte(promptOutput), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "opencode_output.md"), []byte(opencodeOutput), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "logs.md"), []byte(logOutput), 0644)

	fmt.Printf("\n%s%s==================================================\n", Bold, Green)
	fmt.Printf(" ACRE Snyk Remediation Run Completed!\n")
	fmt.Printf(" Final Reports Saved to: %s\n", reportDir)
	fmt.Printf("==================================================%s\n\n", Reset)

	return nil
}

func printSnykHeader() {
	header := `
   ___   ______ ___  ___   ___  _  ______  __
  / _ | / __// _ \/ _ \ / __// |/ / \ \/ /
 / __ |/ /__/ , _/ ___/_\ \ /    /   \  / 
/_/ |_|\___//_/|_/_/   /___//_/|_/    /_/  
ACRE Snyk Security Remediation Engine v2.0.0
`
	fmt.Printf("%s%s%s\n", Bold, Cyan, header)
}

func stripANSI(str string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (str[i] >= 'a' && str[i] <= 'z') || (str[i] >= 'A' && str[i] <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(str[i])
	}
	return b.String()
}
