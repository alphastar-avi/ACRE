package test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Run executes the correct test command in the target repository.
func Run(repoPath string) (exitCode int, stdout string, stderr string) {
	_, args := findTestCommand(repoPath)

	cmd := exec.Command("dotnet", args...)
	cmd.Dir = repoPath

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	} else {
		exitCode = 0
	}

	return exitCode, stdout, stderr
}

// GetCommandString returns the command that will be executed for reporting.
func GetCommandString(repoPath string) string {
	cmdName, _ := findTestCommand(repoPath)
	return cmdName
}

func findTestCommand(repoPath string) (string, []string) {
	hasTestProjects := false
	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csproj") {
			if strings.Contains(strings.ToLower(info.Name()), "test") {
				hasTestProjects = true
			}
		}
		return nil
	})

	if hasTestProjects {
		sln := findSolutionFile(repoPath)
		if sln != "" {
			return "dotnet test " + sln, []string{"test", sln}
		}
		return "dotnet test", []string{"test"}
	}

	return "dotnet run -- --run-tests", []string{"run", "--", "--run-tests"}
}

func findSolutionFile(repoPath string) string {
	files, err := os.ReadDir(repoPath)
	if err != nil {
		return ""
	}
	var slnFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if strings.HasSuffix(name, ".sln") || strings.HasSuffix(name, ".slnx") {
			slnFiles = append(slnFiles, name)
		}
	}
	if len(slnFiles) == 0 {
		return ""
	}
	// Prefer solution files that do not contain "everything" or "all"
	for _, sln := range slnFiles {
		lower := strings.ToLower(sln)
		if !strings.Contains(lower, "everything") && !strings.Contains(lower, "all") {
			return sln
		}
	}
	return slnFiles[0]
}
