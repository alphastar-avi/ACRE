package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	ticketPath := flag.String("ticket", "", "Path to the incident ticket JSON file")
	repoPath := flag.String("repo", "", "Path to the target repository")
	runsDir := flag.String("runs-dir", "", "Path to the runs directory to store reports")
	enablePR := flag.Bool("pr", false, "Create a Git branch, push, and open a PR if successful")
	flag.Parse()

	if *ticketPath == "" || *repoPath == "" || *runsDir == "" {
		log.Println("Error: Missing required arguments.")
		flag.Usage()
		os.Exit(1)
	}

	err := runner.Run(*ticketPath, *repoPath, *runsDir, *enablePR)
	if err != nil {
		log.Fatalf("ACRE execution failed: %v", err)
	}
}
