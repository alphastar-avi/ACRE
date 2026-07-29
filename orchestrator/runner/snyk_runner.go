package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"acre/build"
	"acre/okf"
	"acre/opencode"
	"acre/prompt"
	"acre/report"
	"acre/snyk"
)

// RunSnyk executes the automated Snyk vulnerability remediation workflow.
func RunSnyk(snykRepoPath, reportDir, okfPath string, debugMaxBunches int) error {
	printSnykHeader()

	logBuffer := &strings.Builder{}
	opencodeLogBuffer := &strings.Builder{}
	promptLogBuffer := &strings.Builder{}

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

	logPrintln("   %s[SNYK SCAN COMPLETE]%s Total Issues Found: %d | Open Issues: %d",
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
			logPrintln("   %s[INFO]%s Proceeding without pre-existing OKF file. OKF will be created upon successful fixes.", Yellow, Reset)
		} else {
			okfAbsPath, okfIndex, hasOKF = okf.LoadOKF(okfPath, absRepo)
		}
	} else {
		okfAbsPath, okfIndex, hasOKF = okf.LoadOKF("", absRepo)
	}

	if hasOKF {
		logPrintln("   %s[OKF LOADED]%s Found OKF documentation at: %s", Green, Reset, okfAbsPath)
	} else {
		logPrintln("   %s[INFO]%s No OKF index found. Snyk fixes will proceed with built-in guidelines.", Yellow, Reset)
	}
	logPrintln("")

	// 3. Group Vulnerabilities
	logPrintln("%s[STEP 3/5]%s %sGrouping High-Severity Vulnerabilities into Similarity Bunches...%s", Cyan, Reset, Bold, Reset)
	cfg := snyk.DefaultConfig()

	bunches := snyk.GroupIssues(initialReport.Issues, cfg.TargetSeverities)
	totalBunches := len(bunches)

	if totalBunches == 0 {
		logPrintln("   %s[SUCCESS]%s No high-severity vulnerabilities found targeting remediation!", Green, Reset)
		return generateSnykReports(reportDir, initialReport, 0, 0, nil, promptLogBuffer.String(), opencodeLogBuffer.String(), logBuffer.String())
	}

	if debugMaxBunches > 0 && debugMaxBunches < totalBunches {
		logPrintln("   %s[DEBUG MODE]%s Processing first %d bunch(es) out of %d total bunches.", Yellow, Reset, debugMaxBunches, totalBunches)
		bunches = bunches[:debugMaxBunches]
	} else {
		logPrintln("   Identified %d similarity bunch(es) of high-severity vulnerabilities.", totalBunches)
	}
	logPrintln("")

	// 4. Remediation Loops
	logPrintln("%s[STEP 4/5]%s %sExecuting Remediation & Verification Loops...%s", Cyan, Reset, Bold, Reset)

	type BunchResult struct {
		BunchIndex   int          `json:"bunch_index"`
		Title        string       `json:"title"`
		Issues       []snyk.Issue `json:"issues"`
		SnykVerified bool         `json:"snyk_verified"`
		BuildVerified bool        `json:"build_verified"`
		Approach     string       `json:"approach"`
		CodeChanges  []string     `json:"code_changes"`
		ErrorReason  string       `json:"error_reason"`
	}

	var results []BunchResult
	var totalFixedIssues int

	for bIdx, bunch := range bunches {
		bunchNum := bIdx + 1
		bunchTitle := fmt.Sprintf("[%s] %s (%s)", bunch[0].Severity, bunch[0].Title, extractModuleDir(bunch[0].Path))
		logPrintln("%s--------------------------------------------------%s", Dim, Reset)
		logPrintln("%s[BUNCH %d/%d]%s %s%s%s", Cyan, bunchNum, len(bunches), Reset, Bold, bunchTitle, Reset)
		logPrintln("   Targeting %d vulnerability finding(s):", len(bunch))
		for _, iss := range bunch {
			logPrintln("   - Line %d in %s (ID: %s)", iss.LineNumber, iss.Path, iss.FindingID)
		}

		bResult := BunchResult{
			BunchIndex: bunchNum,
			Title:      bunchTitle,
			Issues:     bunch,
		}

		snykFixVerified := false
		var lastFeedback string

		// Snyk Fix Loop (Max cfg.MaxSnykFixRetries)
		for sAttempt := 1; sAttempt <= cfg.MaxSnykFixRetries; sAttempt++ {
			logPrintln("   %s[FIX ATTEMPT %d/%d]%s Invoking OpenCode to remediate vulnerabilities...",
				Yellow, sAttempt, cfg.MaxSnykFixRetries, Reset)

			snykPrompt := prompt.GenerateSnykPrompt(bunch, absRepo, okfAbsPath, okfIndex, sAttempt, lastFeedback)
			promptLogBuffer.WriteString(fmt.Sprintf("# System Prompt (Bunch %d, Attempt %d)\n\n%s\n\n---\n\n", bunchNum, sAttempt, snykPrompt))

			opencodeOut, err := opencode.Run(snykPrompt, absRepo)
			opencodeLogBuffer.WriteString(fmt.Sprintf("=== BUNCH %d ATTEMPT %d ===\n%s\n\n", bunchNum, sAttempt, opencodeOut))

			if err != nil {
				logPrintln("   %s[WARNING]%s OpenCode execution exited with error: %v", Yellow, Reset, err)
			} else {
				logPrintln("   %s[RUN COMPLETE]%s OpenCode modification completed.", Green, Reset)
			}

			// Re-test Snyk
			logPrintln("   %s[VERIFY]%s Re-testing codebase with Snyk...", Yellow, Reset)
			_, reReport, _ := snyk.RunTest(absRepo)

			// Check if targeted issues in this bunch still exist
			remainingCount := countTargetedIssuesRemaining(bunch, reReport.Issues)

			if remainingCount == 0 {
				logPrintln("   %s[SNYK VERIFIED]%s All %d vulnerability finding(s) in this bunch resolved!", Green, Reset, len(bunch))
				snykFixVerified = true
				break
			} else {
				logPrintln("   %s[SNYK RE-TEST FAIL]%s %d of %d findings still present in Snyk scan.", Red, Reset, remainingCount, len(bunch))
				lastFeedback = fmt.Sprintf("Snyk re-test failed: %d of %d targeted vulnerabilities are still present. Please adjust your fix to eliminate all targeted findings.", remainingCount, len(bunch))
			}
		}

		bResult.SnykVerified = snykFixVerified

		if !snykFixVerified {
			bResult.ErrorReason = fmt.Sprintf("Failed to resolve Snyk findings after %d attempts.", cfg.MaxSnykFixRetries)
			logPrintln("   %s[BUNCH FAILED]%s %s", Red, Reset, bResult.ErrorReason)
			results = append(results, bResult)
			continue
		}

		// Build Verification Loop (Max cfg.MaxBuildRetries)
		buildVerified := false
		buildCommand := build.GetCommandString(absRepo)

		for bAttempt := 1; bAttempt <= cfg.MaxBuildRetries; bAttempt++ {
			logPrintln("   %s[BUILD VERIFY %d/%d]%s Compiling repository solution (%s)...", Yellow, bAttempt, cfg.MaxBuildRetries, Reset, buildCommand)
			bCode, bOut, bErr := build.Run(absRepo)
			if bCode == 0 {
				logPrintln("   %s[BUILD SUCCESS]%s Solution compilation succeeded!", Green, Reset)
				buildVerified = true
				break
			} else {
				logPrintln("   %s[BUILD FAIL]%s Exit code %d. Passing compilation errors to OpenCode...", Red, Reset, bCode)
				buildPrompt := fmt.Sprintf("The security fix built, but compilation failed with exit code %d:\n```\n%s\n%s\n```\nPlease fix the compilation error while preserving the security fix.", bCode, bOut, bErr)
				_, _ = opencode.Run(buildPrompt, absRepo)
			}
		}

		bResult.BuildVerified = buildVerified

		// Read remediation_details.json from repo if created by OpenCode
		detailsPath := filepath.Join(absRepo, "remediation_details.json")
		if detailsBytes, err := os.ReadFile(detailsPath); err == nil {
			var details report.RemediationDetails
			if jsonErr := json.Unmarshal(detailsBytes, &details); jsonErr == nil {
				bResult.Approach = details.Approach
				for _, cc := range details.CodeChanges {
					bResult.CodeChanges = append(bResult.CodeChanges, fmt.Sprintf("%s: %s", cc.File, cc.Description))
				}
			}
			_ = os.Remove(detailsPath)
		}

		if buildVerified && snykFixVerified {
			logPrintln("   %s[BUNCH SUCCESS]%s Bunch %d fully resolved and verified!", Green, Reset, bunchNum)
			totalFixedIssues += len(bunch)

			// Update OKF with remediation findings
			logPrintln("   %s[OKF UPDATE]%s Logging remediation knowledge into OKF...", Cyan, Reset)
			if err := okf.UpdateOKFWithRemediation(absRepo, okfPath, bunchTitle, bunch, bResult.Approach); err != nil {
				logPrintln("   %s[WARNING]%s Failed to update OKF: %v", Yellow, Reset, err)
			} else {
				logPrintln("   %s[OKF UPDATED]%s Remediation knowledge saved to OKF/snyk_remediations.md", Green, Reset)
			}
		} else {
			bResult.ErrorReason = "Build compilation failed after security fix."
			logPrintln("   %s[BUNCH FAILED]%s %s", Red, Reset, bResult.ErrorReason)
		}

		results = append(results, bResult)
	}

	// 5. Generate Reports
	logPrintln("")
	logPrintln("%s[STEP 5/5]%s %sGenerating Final Reports in %s...%s", Cyan, Reset, Bold, reportDir, Reset)
	return generateSnykReports(reportDir, initialReport, totalFixedIssues, len(results), results, promptLogBuffer.String(), opencodeLogBuffer.String(), logBuffer.String())
}

func countTargetedIssuesRemaining(targeted []snyk.Issue, current []snyk.Issue) int {
	currentIDs := make(map[string]bool)
	currentPaths := make(map[string]bool)

	for _, c := range current {
		if c.FindingID != "" {
			currentIDs[c.FindingID] = true
		}
		if c.Path != "" {
			currentPaths[fmt.Sprintf("%s:%d", c.Path, c.LineNumber)] = true
		}
	}

	remaining := 0
	for _, t := range targeted {
		if t.FindingID != "" && currentIDs[t.FindingID] {
			remaining++
		} else if currentPaths[fmt.Sprintf("%s:%d", t.Path, t.LineNumber)] {
			remaining++
		}
	}
	return remaining
}

func generateSnykReports(reportDir string, initialReport *snyk.SnykReport, fixedCount, processedBunches int, results interface{}, promptOutput, opencodeOutput, logOutput string) error {
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Write report.md
	var reportBuilder strings.Builder
	reportBuilder.WriteString("# ACRE Snyk Code Test Vulnerability Remediation Report\n\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Date:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	reportBuilder.WriteString(fmt.Sprintf("* **Target Repository:** `%s`\n\n", initialReport.ProjectPath))

	reportBuilder.WriteString("## Initial Vulnerability Summary\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Total Issues:** %d\n", initialReport.TotalIssues))
	reportBuilder.WriteString(fmt.Sprintf("* **Open Issues:** %d\n", initialReport.OpenIssues))
	reportBuilder.WriteString(fmt.Sprintf("* **HIGH Severity:** %d\n", initialReport.Counts["HIGH"]))
	reportBuilder.WriteString(fmt.Sprintf("* **MEDIUM Severity:** %d\n", initialReport.Counts["MEDIUM"]))
	reportBuilder.WriteString(fmt.Sprintf("* **LOW Severity:** %d\n", initialReport.Counts["LOW"]))
	reportBuilder.WriteString(fmt.Sprintf("* **CRITICAL Severity:** %d\n\n", initialReport.Counts["CRITICAL"]))

	reportBuilder.WriteString("## Remediation Results\n")
	reportBuilder.WriteString(fmt.Sprintf("* **Bunches Processed:** %d\n", processedBunches))
	reportBuilder.WriteString(fmt.Sprintf("* **High Severity Issues Resolved:** %d\n\n", fixedCount))

	reportBuilder.WriteString("### Execution Logs & Raw Outputs\n")
	reportBuilder.WriteString("- System prompts fed to OpenCode are saved in `prompt.md`.\n")
	reportBuilder.WriteString("- Raw OpenCode outputs are saved in `opencode_output.md`.\n")
	reportBuilder.WriteString("- Orchestrator progress logs are saved in `logs.md`.\n")

	_ = os.WriteFile(filepath.Join(reportDir, "report.md"), []byte(reportBuilder.String()), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "prompt.md"), []byte(promptOutput), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "opencode_output.md"), []byte(opencodeOutput), 0644)
	_ = os.WriteFile(filepath.Join(reportDir, "logs.md"), []byte(logOutput), 0644)

	fmt.Printf("\n%s%s==================================================\n", Bold, Green)
	fmt.Printf(" Snyk Remediation Finished!\n")
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
	ansiRegex := regexpANSI()
	return ansiRegex.ReplaceAllString(str, "")
}

func extractModuleDir(pathStr string) string {
	clean := filepath.ToSlash(pathStr)
	parts := strings.Split(clean, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}

func regexpANSI() *regexpCompiler {
	return &regexpCompiler{}
}

type regexpCompiler struct{}

func (r *regexpCompiler) ReplaceAllString(src string, repl string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(src); i++ {
		if src[i] == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (src[i] >= 'a' && src[i] <= 'z') || (src[i] >= 'A' && src[i] <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(src[i])
	}
	return b.String()
}
