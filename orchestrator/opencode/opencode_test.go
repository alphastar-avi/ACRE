package opencode

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRun_StdinExecution(t *testing.T) {
	if _, err := exec.LookPath("opencode"); err != nil {
		t.Skip("opencode CLI not found in PATH, skipping integration test")
		return
	}

	largePrompt := "Respond with 'ACK_OK' only. Ignore this filler: " + strings.Repeat("A", 50000)
	out, err := Run(largePrompt, ".")
	if err != nil {
		t.Fatalf("opencode Run failed with large prompt: %v", err)
	}

	if !strings.Contains(out, "ACK_OK") && !strings.Contains(out, "ACK") {
		t.Logf("Output received from opencode: %s", out)
	}
}
