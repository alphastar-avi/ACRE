package prompt

import (
	"strings"
	"testing"

	"acre/snyk"
)

func TestGenerateSnykPrompt_VerificationCommand(t *testing.T) {
	findings := []snyk.NormalizedFinding{
		{
			RuleID:    "csharp/Sqli",
			Title:     "SQL Injection",
			Level:     "warning",
			Message:   "Unsanitized query execution",
			File:      "Repositories/PrintRepository.cs",
			Line:      402,
			FindingID: "965c5104-8706-4eb6-9d0b-34fc14e604a7",
			CodeFlow:  []snyk.CodeFlowLocation{},
		},
	}

	p := GenerateSnykPrompt(findings, "C:\\Users\\asivak976\\source\\repos\\DH\\SbmsApi", "")

	if !strings.Contains(p, "--snykjson --repo . --ruleid <ruleId> --cli") {
		t.Errorf("Expected prompt to contain verification command structure")
	}

	if !strings.Contains(p, "PowerShell note") {
		t.Errorf("Expected prompt to include PowerShell invocation guidance")
	}

	if !strings.Contains(p, "965c5104-8706-4eb6-9d0b-34fc14e604a7") {
		t.Errorf("Expected prompt to contain findingId")
	}
}
