package main

import (
	"flag"
	"log"
	"os"

	"acre/runner"
)

func main() {
	ticketPath := flag.String("ticket", "", "Path to the incident ticket JSON file")
	repoPath := flag.String("repo", "", "Path to the target repository")
	runsDir := flag.String("runs-dir", "", "Path to the runs directory to store reports")
	flag.Parse()

	if *ticketPath == "" || *repoPath == "" || *runsDir == "" {
		log.Println("Error: Missing required arguments.")
		flag.Usage()
		os.Exit(1)
	}

	err := runner.Run(*ticketPath, *repoPath, *runsDir)
	if err != nil {
		log.Fatalf("ACRE execution failed: %v", err)
	}
}
