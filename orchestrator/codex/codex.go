package codex

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Run executes the codex CLI with the given prompt in the target repository.
func Run(prompt string, repoPath string) (string, error) {
	// Execute the codex command exactly as requested.
	cmd := exec.Command(
		"codex",
		"exec",
		"-s", "workspace-write",
		"-C", repoPath,
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
		return output, fmt.Errorf("codex execution failed: %w", err)
	}

	return output, nil
}
