package opencode

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Run executes the opencode CLI with the given prompt in the target repository.
func Run(prompt string, repoPath string) (string, error) {
	opencodePath, err := exec.LookPath("opencode")
	if err != nil {
		return "", fmt.Errorf("opencode executable not found in PATH: %w", err)
	}

	args := []string{
		"run",
		"--dir", repoPath,
		"--dangerously-skip-permissions",
	}

	if model := os.Getenv("OPENCODE_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, prompt)

	// Execute the opencode command non-interactively with auto-approvals.
	cmd := exec.Command(opencodePath, args...)
	cmd.Stdin = strings.NewReader(prompt)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	output := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		output += "\n-- STDERR --\n" + stderrBuf.String()
	}

	if err != nil {
		return output, fmt.Errorf("opencode execution failed: %w", err)
	}

	return output, nil
}
