package prompt

import (
	"strings"
	"testing"

	"acre/snyk"
	"acre/ticket"
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

func TestGenerate_IncidentRemediationPrompt(t *testing.T) {
	ticket := &ticket.Ticket{
		TicketID: "ENG-1234",
		Summary:  "Fix null pointer in payment gateway service",
		Description: "Payment gateway throws NullReferenceException when metadata is null.",
	}

	// 1. Normal Remediation Mode
	pNormal := Generate(ticket, "/workspace/service", false)
	if !strings.Contains(pNormal, "Auto-Detect Language & Solution Compilation / Build Verification") {
		t.Errorf("Expected prompt to contain language build detection guidelines")
	}
	if !strings.Contains(pNormal, "Targeted Tests & Regression Verification") {
		t.Errorf("Expected prompt to contain targeted test execution guidelines")
	}
	if !strings.Contains(pNormal, "remediation_details.json") {
		t.Errorf("Expected prompt to contain remediation_details.json schema")
	}

	// 2. Recommendations / Test Mode
	pRecs := Generate(ticket, "/workspace/service", true)
	if !strings.Contains(pRecs, "RECOMMENDATIONS-ONLY Mode") {
		t.Errorf("Expected prompt to contain RECOMMENDATIONS-ONLY directive")
	}
	if strings.Contains(pRecs, "Targeted Tests & Regression Verification") {
		t.Errorf("Expected recommendations mode not to include full test execution section")
	}
}
