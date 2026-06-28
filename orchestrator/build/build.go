package build

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Run executes 'dotnet build' in the target repository.
func Run(repoPath string) (exitCode int, stdout string, stderr string) {
	args := []string{"build"}
	if sln := findSolutionFile(repoPath); sln != "" {
		args = append(args, sln)
	}

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
