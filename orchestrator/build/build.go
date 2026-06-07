package build

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// Run executes 'dotnet build' in the target repository.
func Run(repoPath string) (exitCode int, stdout string, stderr string) {
	cmd := exec.Command("dotnet", "build")
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
