package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"strconv"

	"acre/okf"
	"acre/runner"
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

	// Snyk remediation flags
	snykRepoPath := flag.String("snyk", "", "Path to repository to scan and resolve Snyk code test vulnerabilities for")
	reportPath := flag.String("report", "", "Path to directory to output final Snyk remediation reports (report.md, opencode_output.md, logs.md)")
	okfPathUpper := flag.String("OKF", "", "Optional path to OKF documentation directory or file")
	debugFlag := flag.String("debug", "", "Optional max number of high severity vulnerability bunches to process (e.g. 1 or 'a')")

	flag.Parse()

	// Check if opencode is installed in system PATH
	if _, err := exec.LookPath("opencode"); err != nil {
		log.Fatalf("Error: 'opencode' executable not found in system PATH. ACRE requires the OpenCode CLI to be installed. Please install it first (e.g. 'brew install opencode').")
	}

	// If --snyk flag is specified, run Snyk remediation pipeline
	if *snykRepoPath != "" {
		outReportDir := *reportPath
		if outReportDir == "" {
			outReportDir = "runs/snyk_report"
		}
		debugMaxBunches := 0
		if *debugFlag != "" {
			if num, err := strconv.Atoi(*debugFlag); err == nil {
				debugMaxBunches = num
			} else {
				// If debug is set to a single non-numeric token like "a" or "1", default to 1 bunch
				debugMaxBunches = 1
			}
		}

		err := runner.RunSnyk(*snykRepoPath, outReportDir, *okfPathUpper, debugMaxBunches)
		if err != nil {
			log.Fatalf("Snyk Vulnerability Remediation failed: %v", err)
		}
		os.Exit(0)
	}

	// If --okf flag is specified, run the documentation indexer and exit
	if *okfRepoPath != "" {
		err := okf.Generate(*okfRepoPath, *okfScope)
		if err != nil {
			log.Fatalf("OKF Generation failed: %v", err)
		}
		os.Exit(0)
	}

	if *ticketPath == "" || *repoPath == "" || *runsDir == "" {
		log.Println("Error: Missing required arguments.")
		flag.Usage()
		os.Exit(1)
	}

	err := runner.Run(*ticketPath, *repoPath, *runsDir, *enablePR, *enableRecs, *enableTest)
	if err != nil {
		log.Fatalf("ACRE execution failed: %v", err)
	}
}
