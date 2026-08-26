package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"acre/okf"
	"acre/runner"
	"acre/snyk"
)

func loadEnv() {
	// Try loading from current working directory
	if err := parseEnvFile(".env"); err == nil {
		return
	}
	// Try loading from executable directory
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		_ = parseEnvFile(filepath.Join(exeDir, ".env"))
	}
}

func parseEnvFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
	return nil
}

func main() {
	loadEnv()

	okfRepoPath := flag.String("okf", "", "Path to the target repository to scan and generate OKF v0.1 documentation for")
	okfScope := flag.String("scope", "", "Optional relative path within the repository to focus OKF documentation scanning on (e.g. src/Services/Basket)")
	ticketPath := flag.String("ticket", "", "Path to the incident ticket JSON file")
	repoPath := flag.String("repo", "", "Path to the target repository")
	runsDir := flag.String("runs-dir", "", "Path to the runs directory to store reports")
	enablePR := flag.Bool("pr", false, "Create a Git branch, push, and open a PR if successful")
	enableRecs := flag.Bool("r", false, "Analyze codebase, identify root cause, write structured recommendations report, and open a PR without changing codebase files")
	enableTest := flag.Bool("test", false, "Test mode: analyze codebase, compile solution, generate report & manual PR URL without git operations or running regression tests")

	skillPathLower := flag.String("skill", "", "Optional path to custom skill markdown file (.md) or skills directory to inject domain rules and business logic")
	skillPathUpper := flag.String("SKILL", "", "Optional path to custom skill markdown file (.md) or skills directory to inject domain rules and business logic")
	okfPathLower := flag.String("okf-path", "", "Optional path to explicit OKF documentation folder or file for incident remediation")
	okfPathUpper := flag.String("OKF-PATH", "", "Optional path to explicit OKF documentation folder or file for incident remediation")
	okfPathShort := flag.String("OKF", "", "Optional path to explicit OKF documentation folder or file for incident remediation")
	okfDest := flag.String("dest", "", "Optional target destination directory for generated OKF documentation (e.g. OKF/Smartstore)")
	okfTarget := flag.String("target", "", "Optional target destination directory for generated OKF documentation (e.g. OKF/Smartstore)")

	// Snyk workflow flags
	enableSnykJson := flag.Bool("snykjson", false, "Run 'snyk code test --json', normalize findings SARIF, and output to snykOutput/Output<datetime>.json")
	ruleIdFlag := flag.String("ruleid", "", "Optional rule ID filter for --snykjson (e.g. csharp/PT)")
	cliFlag := flag.Bool("cli", false, "Print normalized findings JSON directly to stdout for CLI consumption")
	snykJsonPath := flag.String("snyk", "", "Path to normalized Snyk JSON file (Output<datetime>.json) for remediation")
	reportPath := flag.String("report", "", "Path to directory to output final Snyk remediation reports (report.md, opencode_output.md, logs.md)")
	refPathUpper := flag.String("REF", "", "Optional path to reference directory or .md file")
	refPathLower := flag.String("ref", "", "Optional path to reference directory or .md file")

	flag.Parse()

	// Handle --snykjson flag first (does not require opencode CLI for JSON extraction)
	if *enableSnykJson {
		targetRepo := *repoPath
		if targetRepo == "" {
			targetRepo = "."
		}
		outputPath, err := snyk.GenerateSnykJSON(targetRepo, *ruleIdFlag, *cliFlag)
		if err != nil {
			log.Fatalf("Snyk JSON generation failed: %v", err)
		}
		if !*cliFlag {
			log.Printf("Normalized Snyk JSON successfully generated at: %s", outputPath)
		}
		os.Exit(0)
	}

	// Check if opencode is installed in system PATH for agent workflows
	if _, err := exec.LookPath("opencode"); err != nil {
		log.Fatalf("Error: 'opencode' executable not found in system PATH. ACRE requires the OpenCode CLI to be installed. Please install it first (e.g. 'brew install opencode').")
	}

	// If --snyk flag is specified, run Snyk remediation pipeline
	if *snykJsonPath != "" {
		outReportDir := *reportPath
		if outReportDir == "" {
			outReportDir = "runs/snyk_report"
		}
		ref := *refPathUpper
		if ref == "" {
			ref = *refPathLower
		}

		err := runner.RunSnyk(*snykJsonPath, *repoPath, outReportDir, ref)
		if err != nil {
			log.Fatalf("Snyk Vulnerability Remediation failed: %v", err)
		}
		os.Exit(0)
	}

	// If --okf flag is specified, run the documentation indexer and exit
	if *okfRepoPath != "" {
		targetOut := *okfDest
		if targetOut == "" {
			targetOut = *okfTarget
		}
		err := okf.Generate(*okfRepoPath, *okfScope, targetOut)
		if err != nil {
			log.Fatalf("OKF Generation failed: %v", err)
		}
		os.Exit(0)
	}

	if *ticketPath == "" || *repoPath == "" {
		log.Println("Error: Missing required arguments (--ticket and --repo are required).")
		flag.Usage()
		os.Exit(1)
	}

	runs := *runsDir
	if runs == "" {
		runs = "runs"
	}

	skill := *skillPathLower
	if skill == "" {
		skill = *skillPathUpper
	}

	okfPath := *okfPathLower
	if okfPath == "" {
		okfPath = *okfPathUpper
	}
	if okfPath == "" {
		okfPath = *okfPathShort
	}

	err := runner.Run(*ticketPath, *repoPath, runs, *enablePR, *enableRecs, *enableTest, skill, okfPath)
	if err != nil {
		log.Fatalf("ACRE execution failed: %v", err)
	}
}
