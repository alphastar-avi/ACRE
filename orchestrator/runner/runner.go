package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acre/build"
	"acre/github"
	"acre/opencode"
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
func Run(ticketPath, repoPath, runsDir string, enablePR, enableRecs bool) error {
	printHeader()

	// 1. Load Ticket
	fmt.Printf("%s[%s]%s Loading incident ticket %s...\n", Cyan, "1/6", Reset, ticketPath)
	t, err := ticket.Load(ticketPath)
	if err != nil {
		return fmt.Errorf("failed to load ticket: %w", err)
	}
	fmt.Printf("   %sID:%s %s\n", Bold, Reset, t.TicketID)
	fmt.Printf("   %sSummary:%s %s\n\n", Bold, Reset, t.Summary)

	// Branching integration
	var baseBranch, branchName string
	shouldCreatePR := enablePR || enableRecs
	if shouldCreatePR {
		baseBranch, err = github.GetBaseBranch(repoPath)
		if err != nil {
			return fmt.Errorf("failed to get active base branch: %w", err)
		}
		branchName = t.TicketID // e.g. ENG-0001
		fmt.Printf("   [Git PR] Creating and checking out branch %s from %s...\n", branchName, baseBranch)
		if err := github.CreateBranch(repoPath, branchName, baseBranch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", err)
		}
	}

	// 2. Generate Prompt
	fmt.Printf("%s[%s]%s Generating initial remediation prompt...\n", Cyan, "2/6", Reset)
	p := prompt.Generate(t, repoPath, enableRecs)
	fmt.Printf("   Initial prompt generated (%d chars).\n\n", len(p))

	currentPrompt := p
	maxRetries := 3
	var opencodeOut string
	var buildCode, testCode int
	var buildOut, buildErr, testOut, testErr string
	var runSuccess bool
	var finalPRURL string

	buildCommand := build.GetCommandString(repoPath)
	testCommand := test.GetCommandString(repoPath)

	if enableRecs {
		fmt.Printf("   %sStep A:%s Executing OpenCode CLI for analysis (recommendations only)...\n", Yellow, Reset)
		opencodeOut, err = opencode.Run(currentPrompt, repoPath)
		if err != nil {
			fmt.Printf("   %s[Warning]%s OpenCode execution exited with code/error: %v\n", Yellow, Reset, err)
		} else {
			fmt.Printf("   %s[Success]%s OpenCode analysis run completed.\n", Green, Reset)
		}
		runSuccess = true
	} else {
		// Self-healing loop
		for attempt := 1; attempt <= maxRetries; attempt++ {
			fmt.Printf("%s[%s]%s %sRemediation Attempt %d/%d%s\n", Cyan, "LOOP", Reset, Bold, attempt, maxRetries, Reset)
			fmt.Printf("   %sStep A:%s Executing OpenCode CLI (non-interactive)...\n", Yellow, Reset)
			
			opencodeOut, err = opencode.Run(currentPrompt, repoPath)
			if err != nil {
				fmt.Printf("   %s[Warning]%s OpenCode execution exited with code/error: %v\n", Yellow, Reset, err)
			} else {
				fmt.Printf("   %s[Success]%s OpenCode modification run completed.\n", Green, Reset)
			}

			// Run Build
			fmt.Printf("   %sStep B:%s Compiling repository (%s)...\n", Yellow, Reset, buildCommand)
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
			fmt.Printf("   %sStep C:%s Running regression tests (%s)...\n", Yellow, Reset, testCommand)
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
	}

	if !runSuccess {
		fmt.Printf("%s[Outcome] Remediation Failed after %d attempts.%s\n\n", Red, maxRetries, Reset)
	} else {
		if enableRecs {
			fmt.Printf("%s[Outcome] Recommendations Gathering Completed!%s\n\n", Green, Reset)
		} else {
			fmt.Printf("%s[Outcome] Remediation Successfully Completed!%s\n\n", Green, Reset)
		}
	}

	// 5. Parse remediation_details.json from repo
	detailsPath := filepath.Join(repoPath, "remediation_details.json")
	var details report.RemediationDetails
	hasDetails := false

	if _, err := os.Stat(detailsPath); err == nil {
		if detailsBytes, readErr := os.ReadFile(detailsPath); readErr == nil {
			if jsonErr := json.Unmarshal(detailsBytes, &details); jsonErr == nil {
				hasDetails = true
			}
		}
		_ = os.Remove(detailsPath)
	}

	// Git Branch and PR operations
	if shouldCreatePR {
		var commitMsg string
		var prTitle string
		var prBody string

		if enableRecs {
			// 1. Discard codebase source file modifications
			fmt.Printf("   [Git PR] Discarding codebase source file modifications (recommendations mode)...\n")
			_, _ = github.RunCommand(repoPath, "git", "checkout", "--", ".")
			_, _ = github.RunCommand(repoPath, "git", "clean", "-fd")

			// 2. Generate structured report file recommendations.md at root of target repository
			var builder strings.Builder
			builder.WriteString("# ACRE Incident Analysis & Recommendations\n\n")
			builder.WriteString("## Ticket Information\n")
			builder.WriteString(fmt.Sprintf("* **Ticket ID:** %s\n", t.TicketID))
			builder.WriteString(fmt.Sprintf("* **Summary:** %s\n\n", t.Summary))
			builder.WriteString("### Description\n")
			builder.WriteString(t.Description + "\n\n")
			if t.AcceptanceCriteria != "" {
				builder.WriteString("### Acceptance Criteria\n")
				builder.WriteString(t.AcceptanceCriteria + "\n\n")
			}

			builder.WriteString("## OpenCode Incident Analysis\n")
			if hasDetails {
				builder.WriteString(fmt.Sprintf("* **Confidence Score:** %d/100\n", details.ConfidenceScore))
				builder.WriteString(fmt.Sprintf("* **Justification:** %s\n\n", details.ConfidenceJustification))
				builder.WriteString(fmt.Sprintf("### Understanding of the Issue\n%s\n\n", details.UnderstoodIssue))
				builder.WriteString(fmt.Sprintf("### Core Root Cause Identified\n%s\n\n", details.PotentialIssue))
				builder.WriteString(fmt.Sprintf("### Potential Approach to Fix\n%s\n\n", details.Approach))
				builder.WriteString("### Concise Clear Code Changes Needed\n")
				if len(details.CodeChanges) > 0 {
					for _, c := range details.CodeChanges {
						builder.WriteString(fmt.Sprintf("* **File:** `%s`\n  * **Change:** %s\n", c.File, c.Description))
					}
				} else {
					builder.WriteString("_No files were marked for modification._\n")
				}
				if details.Recommendations != "" {
					builder.WriteString(fmt.Sprintf("\n### Extra Recommendations\n%s\n", details.Recommendations))
				}
			} else {
				builder.WriteString("> [!WARNING]\n> Analysis failed. OpenCode did not write `remediation_details.json`.\n")
			}

			recsFilePath := filepath.Join(repoPath, "recommendations.md")
			_ = os.WriteFile(recsFilePath, []byte(builder.String()), 0644)

			commitMsg = fmt.Sprintf("docs: add incident recommendations for %s", t.TicketID)
			prTitle = fmt.Sprintf("docs: recommendations for incident %s - %s", t.TicketID, t.Summary)
			prBody = builder.String()
		} else {
			// Normal flow
			if runSuccess {
				commitMsg = fmt.Sprintf("fix: resolve incident %s", t.TicketID)
			} else {
				commitMsg = fmt.Sprintf("fix: attempt to resolve incident %s (Validation Failed)", t.TicketID)
			}

			var prBodyBuilder strings.Builder
			if !runSuccess {
				prBodyBuilder.WriteString("> [!WARNING]\n> **Verification Validation Failed**: ACRE completed execution, but the compilation build or regression tests failed. Please inspect the code changes below and review the logs.\n\n")
			}

			if hasDetails {
				prBodyBuilder.WriteString(fmt.Sprintf("## Understood Issue\n%s\n\n", details.UnderstoodIssue))
				prBodyBuilder.WriteString(fmt.Sprintf("## Core Root Cause\n%s\n\n", details.PotentialIssue))
				prBodyBuilder.WriteString(fmt.Sprintf("## Approach Used to Resolve\n%s\n\n", details.Approach))

				prBodyBuilder.WriteString("## Applied Code Changes\n")
				if len(details.CodeChanges) > 0 {
					for _, c := range details.CodeChanges {
						prBodyBuilder.WriteString(fmt.Sprintf("* **File:** `%s` - %s\n", c.File, c.Description))
					}
				} else {
					prBodyBuilder.WriteString("_No files were modified._\n")
				}
				prBodyBuilder.WriteString("\n")

				prBodyBuilder.WriteString(fmt.Sprintf("## Wrote Tests\n* %t\n\n", details.WroteTests))
				
				if details.Recommendations != "" {
					prBodyBuilder.WriteString(fmt.Sprintf("## Recommendations / Diagnostic Notes\n%s\n\n", details.Recommendations))
				}
			} else {
				// Fallback when details are missing (e.g. build failed before writing details)
				prBodyBuilder.WriteString("## Understood Issue\nACRE loaded the ticket and attempted remediation.\n\n")
				prBodyBuilder.WriteString("## Core Root Cause\nFailed during compilation or verification tests.\n\n")
				prBodyBuilder.WriteString("## Approach Used to Resolve\nOpenCode was executed but the codebase did not successfully compile or pass regression tests.\n\n")
				prBodyBuilder.WriteString("## Recommendations\nCheck the build and test logs in the ACRE runs directory to diagnose the compile/verification errors.\n\n")
			}

			if runSuccess {
				prTitle = fmt.Sprintf("fix: resolve incident %s - %s", t.TicketID, t.Summary)
			} else {
				prTitle = fmt.Sprintf("fix: resolve incident %s - %s (Validation Failed)", t.TicketID, t.Summary)
			}
			prBody = prBodyBuilder.String()
		}

		fmt.Printf("   [Git PR] Committing and pushing changes to branch %s...\n", branchName)
		if err := github.CommitAndPush(repoPath, branchName, commitMsg); err != nil {
			fmt.Printf("   %s[Warning]%s Failed to push git changes: %v\n", Yellow, Reset, err)
		} else {
			fmt.Printf("   [Git PR] Creating pull request from %s to %s...\n", branchName, baseBranch)
			prURL, manualURL, prErr := github.CreatePR(repoPath, baseBranch, branchName, prTitle, prBody)
			if prErr != nil {
				fmt.Printf("   %s[Warning]%s Failed to create pull request via API: %v\n", Yellow, Reset, prErr)
			}
			if prURL != "" {
				finalPRURL = prURL
				fmt.Printf("\n   %s[Git PR Success]%s Pull request successfully created:\n   %s\n", Green, Reset, prURL)
			} else {
				finalPRURL = manualURL
				fmt.Printf("\n   %s[Git PR]%s No GITHUB_TOKEN configured or API failed. Click the URL below to create the PR manually:\n   %s\n", Yellow, Reset, manualURL)
			}
		}
		// Switch back to base branch (keep local branch pushed to remote)
		_ = github.CheckoutBranch(repoPath, baseBranch)
	}

	// 6. Report
	fmt.Printf("%s[%s]%s Packaging final run report...\n", Cyan, "6/6", Reset)
	reportData := report.Data{
		Ticket:         t,
		Prompt:         p,
		OpenCodeOutput: opencodeOut,
		BuildCommand:   buildCommand,
		BuildExitCode:  buildCode,
		BuildStdout:    buildOut,
		BuildStderr:    buildErr,
		TestCommand:    testCommand,
		TestExitCode:   testCode,
		TestStdout:     testOut,
		TestStderr:     testErr,
		RepositoryPath: repoPath,
		Details:        &details,
		HasDetails:     hasDetails,
		PullRequestURL: finalPRURL,
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
