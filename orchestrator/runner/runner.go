package runner

import (
	"fmt"

	"acre/build"
	"acre/codex"
	"acre/prompt"
	"acre/report"
	"acre/test"
	"acre/ticket"
)

// ANSI styles
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Green     = "\033[32m"
	Red       = "\033[31m"
	Yellow    = "\033[33m"
	Cyan      = "\033[36m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	BGDark    = "\033[48;5;235m"
)

// Run executes the full incident remediation pipeline with self-healing retries.
func Run(ticketPath, repoPath, runsDir string) error {
	printHeader()

	// 1. Load Ticket
	fmt.Printf("%s[%s]%s Loading incident ticket %s...\n", Cyan, "1/6", Reset, ticketPath)
	t, err := ticket.Load(ticketPath)
	if err != nil {
		return fmt.Errorf("failed to load ticket: %w", err)
	}
	fmt.Printf("   %sID:%s %s\n", Bold, Reset, t.TicketID)
	fmt.Printf("   %sSummary:%s %s\n\n", Bold, Reset, t.Summary)

	// 2. Generate Prompt
	fmt.Printf("%s[%s]%s Generating initial remediation prompt...\n", Cyan, "2/6", Reset)
	p := prompt.Generate(t, repoPath)
	fmt.Printf("   Initial prompt generated (%d chars).\n\n", len(p))

	currentPrompt := p
	maxRetries := 3
	var codexOut string
	var buildCode, testCode int
	var buildOut, buildErr, testOut, testErr string
	var runSuccess bool

	// Self-healing loop
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("%s[%s]%s %sRemediation Attempt %d/%d%s\n", Cyan, "LOOP", Reset, Bold, attempt, maxRetries, Reset)
		fmt.Printf("   %sStep A:%s Executing Codex CLI (non-interactive)...\n", Yellow, Reset)
		
		codexOut, err = codex.Run(currentPrompt, repoPath)
		if err != nil {
			fmt.Printf("   %s[Warning]%s Codex execution exited with code/error: %v\n", Yellow, Reset, err)
		} else {
			fmt.Printf("   %s[Success]%s Codex modification run completed.\n", Green, Reset)
		}

		// Run Build
		fmt.Printf("   %sStep B:%s Compiling repository (dotnet build)...\n", Yellow, Reset)
		buildCode, buildOut, buildErr = build.Run(repoPath)
		if buildCode != 0 {
			fmt.Printf("   %s[Fail]%s Build failed with exit code %d.\n", Red, Reset, buildCode)
			
			// Provide feedback for self-healing
			currentPrompt = fmt.Sprintf("%s\n\n## Feedback (Attempt %d)\nYour previous modification failed to build with the following error:\n```\n%s\n%s\n```\nPlease correct your modification to resolve this build error.", p, attempt, buildOut, buildErr)
			fmt.Printf("   %s[Self-Healing]%s Appended build errors. Retrying...\n\n", Magenta, Reset)
			continue
		}
		fmt.Printf("   %s[Success]%s Build succeeded.\n", Green, Reset)

		// Run Tests
		fmt.Printf("   %sStep C:%s Running regression tests (dotnet run -- --run-tests)...\n", Yellow, Reset)
		testCode, testOut, testErr = test.Run(repoPath)
		if testCode != 0 {
			fmt.Printf("   %s[Fail]%s Tests failed with exit code %d.\n", Red, Reset, testCode)
			
			// Provide feedback for self-healing
			currentPrompt = fmt.Sprintf("%s\n\n## Feedback (Attempt %d)\nYour previous modification successfully built, but tests failed with the following output:\n```\n%s\n%s\n```\nPlease adjust your modifications to pass the regression tests.", p, attempt, testOut, testErr)
			fmt.Printf("   %s[Self-Healing]%s Appended test failures. Retrying...\n\n", Magenta, Reset)
			continue
		}

		fmt.Printf("   %s[Success]%s All tests passed!\n\n", Green, Reset)
		runSuccess = true
		break
	}

	if !runSuccess {
		fmt.Printf("%s[Outcome] Remediation Failed after %d attempts.%s\n\n", Red, maxRetries, Reset)
	} else {
		fmt.Printf("%s[Outcome] Remediation Successfully Completed!%s\n\n", Green, Reset)
	}

	// 6. Report
	fmt.Printf("%s[%s]%s Packaging final run report...\n", Cyan, "6/6", Reset)
	reportData := report.Data{
		Ticket:         t,
		Prompt:         p,
		CodexOutput:    codexOut,
		BuildExitCode:  buildCode,
		BuildStdout:    buildOut,
		BuildStderr:    buildErr,
		TestExitCode:   testCode,
		TestStdout:     testOut,
		TestStderr:     testErr,
		RepositoryPath: repoPath,
	}

	runDir, err := report.Generate(runsDir, reportData)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("\n%s%s==================================================\n", Bold, Green)
	fmt.Printf(" ACRE Run Finished Successfully!\n")
	fmt.Printf(" Run Report Saved to: %s\n", runDir)
	fmt.Printf("==================================================%s\n", Reset)

	return nil
}

func printHeader() {
	header := `
   ___   ______ ___  ___
  / _ | / __// _ \/ _ \
 / __ |/ /__/ , _/ ___/
/_/ |_|\___//_/|_/_/    
Automatic Code Remediation Engine v1.0
`
	fmt.Printf("%s%s%s\n", Bold, Green, header)
}
