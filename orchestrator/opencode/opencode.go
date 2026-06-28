package opencode

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Run executes the opencode CLI with the given prompt in the target repository.
func Run(prompt string, repoPath string) (string, error) {
	// Execute the opencode command non-interactively with auto-approvals.
	cmd := exec.Command(
		"opencode",
		"run",
		"--dir", repoPath,
		"--dangerously-skip-permissions",
		"--model", "opencode/big-pickle",
		prompt,
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	output := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		output += "\n-- STDERR --\n" + stderrBuf.String()
	}

	if err != nil {
		return output, fmt.Errorf("opencode execution failed: %w", err)
	}

	return output, nil
}
