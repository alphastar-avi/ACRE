package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"acre/okf"
	"acre/opencode"
	"acre/prompt"
	"acre/report"
	"acre/snyk"
)

// RunSnyk executes the automated Snyk vulnerability remediation workflow as a wrapper.
func RunSnyk(snykRepoPath, reportDir, okfPath string, debugMaxBunches int) error {
	printSnykHeader()

	logBuffer := &strings.Builder{}
	opencodeLogBuffer := &strings.Builder{}

	logPrintln := func(format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		fmt.Println(msg)
		logBuffer.WriteString(stripANSI(msg) + "\n")
	}

	absRepo, err := filepath.Abs(snykRepoPath)
	if err != nil {
		return fmt.Errorf("invalid snyk repository path: %w", err)
	}

	// 1. Initial Snyk Scan
	logPrintln("%s[STEP 1/5]%s %sScanning repository with Snyk Code Test...%s", Cyan, Reset, Bold, Reset)
	logPrintln("   Target Repository: %s", absRepo)

	_, initialReport, scanErr := snyk.RunTest(absRepo)
	if scanErr != nil {
		logPrintln("   %s[WARNING]%s Snyk scan execution warning/error: %v", Yellow, Reset, scanErr)
	}

	logPrintln("   %s[INITIAL SNYK SCAN COMPLETE]%s Total Issues: %d | Open Issues: %d",
		Green, Reset, initialReport.TotalIssues, initialReport.OpenIssues)
	logPrintln("   Breakdown -> HIGH: %d | MEDIUM: %d | LOW: %d | CRITICAL: %d",
		initialReport.Counts["HIGH"], initialReport.Counts["MEDIUM"], initialReport.Counts["LOW"], initialReport.Counts["CRITICAL"])
	logPrintln("")

	// 2. OKF Check & User Prompt
	logPrintln("%s[STEP 2/5]%s %sValidating Open Knowledge Format (OKF) Context...%s", Cyan, Reset, Bold, Reset)
	var okfAbsPath, okfIndex string
	var hasOKF bool

	if okfPath != "" {
		if _, statErr := os.Stat(okfPath); os.IsNotExist(statErr) {
			fmt.Printf("   %s[PROMPT]%s Cant find OKF, should i proceed without one? Y/N: ", Yellow, Reset)
			reader := bufio.NewReader(os.Stdin)
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(strings.ToUpper(ans))
			if strings.HasPrefix(ans, "N") {
				logPrintln("   %s[ABORTED]%s User chose not to proceed without OKF.", Red, Reset)
				return fmt.Errorf("aborted by user: OKF not found at %s", okfPath)
			}
			logPrintln("   %s[INFO]%s Proceeding without pre-existing OKF. OKF documentation will be created by OpenCode upon successful remediation.", Yellow, Reset)
		} else {
			okfAbsPath, okfIndex, hasOKF = okf.LoadOKF(okfPath, absRepo)
		}
	} else {
		okfAbsPath, okfIndex, hasOKF = okf.LoadOKF("", absRepo)
	}

	if hasOKF {
		logPrintln("   %s[OKF LOADED]%s Found OKF documentation at: %s", Green, Reset, okfAbsPath)
	} else {
		logPrintln("   %s[INFO]%s No OKF index found. OpenCode will create OKF documentation during remediation.", Yellow, Reset)
	}
	logPrintln("")

	// 3. Filter & Group Vulnerabilities into Similarity Bunches
	logPrintln("%s[STEP 3/5]%s %sGrouping Target Vulnerabilities into Similarity Bunches...%s", Cyan, Reset, Bold, Reset)
	cfg := snyk.DefaultConfig()

	bunches := snyk.GroupIssues(initialReport.Issues, cfg.TargetSeverities)
	if len(bunches) == 0 {
		logPrintln("   %s[SUCCESS]%s No targeted vulnerabilities found for remediation!", Green, Reset)
		return generateSnykReports(reportDir, initialReport, initialReport, nil, "", "", logBuffer.String())
	}

	totalBunches := len(bunches)
	if debugMaxBunches > 0 && debugMaxBunches < totalBunches {
		logPrintln("   %s[DEBUG MODE]%s Processing top %d bunch(es) out of %d total bunches.", Yellow, Reset, debugMaxBunches, totalBunches)
		bunches = bunches[:debugMaxBunches]
	} else {
		logPrintln("   Prepared %d similarity bunch(es) of vulnerabilities for OpenCode.", totalBunches)
	}

	var targetIssues []snyk.Issue
	for _, b := range bunches {
		targetIssues = append(targetIssues, b...)
	}
	logPrintln("   Total targeted findings in selected bunch(es): %d", len(targetIssues))
	logPrintln("")

	// 4. Construct Prompt & Execute OpenCode
	logPrintln("%s[STEP 4/5]%s %sExecuting OpenCode Remediation Agent...%s", Cyan, Reset, Bold, Reset)
	snykPrompt := prompt.GenerateSnykPrompt(targetIssues, absRepo, okfAbsPath, okfIndex)

	logPrintln("   Invoking OpenCode (non-interactive)...")
	opencodeOut, err := opencode.Run(snykPrompt, absRepo)
	opencodeLogBuffer.WriteString(opencodeOut + "\n")

	if err != nil {
		logPrintln("   %s[WARNING]%s OpenCode execution completed with warning/error: %v", Yellow, Reset, err)
	} else {
		logPrintln("   %s[RUN COMPLETE]%s OpenCode agent execution finished.", Green, Reset)
	}
	logPrintln("")

	// 5. Final Snyk Verification Scan & Report Generation
	logPrintln("%s[STEP 5/5]%s %sExecuting Final Snyk Code Test Verification Scan...%s", Cyan, Reset, Bold, Reset)
	_, finalReport, verifyErr := snyk.RunTest(absRepo)
	if verifyErr != nil {
		logPrintln("   %s[WARNING]%s Final Snyk verification scan error: %v", Yellow, Reset, verifyErr)
	}

	logPrintln("   %s[FINAL SNYK SCAN COMPLETE]%s Total Issues: %d | Open Issues: %d",
		Green, Reset, finalReport.TotalIssues, finalReport.OpenIssues)
	logPrintln("   Breakdown -> HIGH: %d | MEDIUM: %d | LOW: %d | CRITICAL: %d",
		finalReport.Counts["HIGH"], finalReport.Counts["MEDIUM"], finalReport.Counts["LOW"], finalReport.Counts["CRITICAL"])

	resolvedDelta := initialReport.OpenIssues - finalReport.OpenIssues
	if resolvedDelta > 0 {
		logPrintln("   %s[REMEDIATION SUCCESS]%s Successfully eliminated %d vulnerability finding(s)!", Green, Reset, resolvedDelta)
	} else {
		logPrintln("   %s[INFO]%s Scan complete. Open issues count after fix: %d.", Yellow, Reset, finalReport.OpenIssues)
	}
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

	return generateSnykReports(reportDir, initialReport, finalReport, detailsPtr, snykPrompt, opencodeLogBuffer.String(), logBuffer.String())
}

func generateSnykReports(reportDir string, initialReport, finalReport *snyk.SnykReport, details *report.RemediationDetails, promptOutput, opencodeOutput, logOutput string) error {
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	resolvedDelta := initialReport.OpenIssues - finalReport.OpenIssues

	// Write report.md
	var reportBuilder strings.Builder
	reportBuilder.WriteString("# ACRE Snyk Code Test Vulnerability Remediation Report\n\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Date:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	reportBuilder.WriteString(fmt.Sprintf("* **Target Repository:** `%s`\n\n", initialReport.ProjectPath))

	reportBuilder.WriteString("## Initial Vulnerability Summary (Before Remediation)\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Total Issues:** %d\n", initialReport.TotalIssues))
	reportBuilder.WriteString(fmt.Sprintf("* **Open Issues:** %d\n", initialReport.OpenIssues))
	reportBuilder.WriteString(fmt.Sprintf("* **HIGH Severity:** %d\n", initialReport.Counts["HIGH"]))
	reportBuilder.WriteString(fmt.Sprintf("* **MEDIUM Severity:** %d\n", initialReport.Counts["MEDIUM"]))
	reportBuilder.WriteString(fmt.Sprintf("* **LOW Severity:** %d\n", initialReport.Counts["LOW"]))
	reportBuilder.WriteString(fmt.Sprintf("* **CRITICAL Severity:** %d\n\n", initialReport.Counts["CRITICAL"]))

	reportBuilder.WriteString("## Verification Summary (After Remediation)\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Total Remaining Open Issues:** %d\n", finalReport.OpenIssues))
	reportBuilder.WriteString(fmt.Sprintf("* **Vulnerabilities Resolved Delta:** %d\n", resolvedDelta))
	reportBuilder.WriteString(fmt.Sprintf("* **HIGH Severity Remaining:** %d\n", finalReport.Counts["HIGH"]))
	reportBuilder.WriteString(fmt.Sprintf("* **MEDIUM Severity Remaining:** %d\n", finalReport.Counts["MEDIUM"]))
	reportBuilder.WriteString(fmt.Sprintf("* **LOW Severity Remaining:** %d\n", finalReport.Counts["LOW"]))
	reportBuilder.WriteString(fmt.Sprintf("* **CRITICAL Severity Remaining:** %d\n\n", finalReport.Counts["CRITICAL"]))

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
ACRE Snyk Security Remediation Engine v1.0
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

