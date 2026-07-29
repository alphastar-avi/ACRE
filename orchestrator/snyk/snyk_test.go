package snyk

import (
	"testing"
)

const sampleSnykOutput = `
PS C:\WINDOWS\system32> snyk code test "C:\Users\asivak976\source\repos\Plartform Backup\Platform"

Testing C:\Users\asivak976\source\repos\Platform ...

Open Issues

 ✗ [LOW] Exposure of Private Personal Information to an Unauthorized Actor
   Finding ID: 9eeb0d00-65bd-4ee6-ba1f-af9970b0bb46
   Path: MDM/MDM.Web/Features/CentralLoginMigration/Hangfire/MigrateCentralLoginUsersJob.cs, line 64
   Info: Private information from sensitive data source flows into global::Microsoft.Extensions.Logging.ILogger.LogInformation. This may disclose private information to an attacker.

 ✗ [MEDIUM] Server-Side Request Forgery (SSRF)
   Finding ID: 30f1adc8-87f4-41a8-a13e-3652f8db95ee
   Path: MDM/Integration/CustomFields/OneStrata.CustomFields.Common/Integrations/RDM/PickListRdmClient.cs, line 72
   Info: Unsanitized input from data flows into PostAsync.

 ✗ [HIGH] Hardcoded Non-Cryptographic Secret
   Finding ID: 045b05e7-6029-4413-8028-4a82d4c7c965
   Path: client/src/mocks/handlers/externalIntegration/data.ts, line 241
   Info: Avoid hardcoding values that are meant to be secret.

╭───────────────────────────────────────────────────────────────────────────────────╮
│ Test Summary                                                                      │
│                                                                                   │
│   Total issues:   259                                                             │
│   Ignored issues: 0 [ 0 HIGH  0 MEDIUM  0 LOW ]                                   │
│   Open issues:    259 [ 1 HIGH  7 MEDIUM  5LOW ]                               │
╰───────────────────────────────────────────────────────────────────────────────────╯
`

func TestParseOutput(t *testing.T) {
	report := ParseOutput(sampleSnykOutput)

	if len(report.Issues) != 3 {
		t.Fatalf("Expected 3 issues parsed, got %d", len(report.Issues))
	}

	if report.TotalIssues != 259 {
		t.Errorf("Expected TotalIssues 259, got %d", report.TotalIssues)
	}

	highIssue := report.Issues[2]
	if highIssue.Severity != "HIGH" {
		t.Errorf("Expected HIGH severity, got %s", highIssue.Severity)
	}
	if highIssue.FindingID != "045b05e7-6029-4413-8028-4a82d4c7c965" {
		t.Errorf("Unexpected finding ID: %s", highIssue.FindingID)
	}
	if highIssue.LineNumber != 241 {
		t.Errorf("Expected line 241, got %d", highIssue.LineNumber)
	}
}

func TestGroupIssues(t *testing.T) {
	report := ParseOutput(sampleSnykOutput)
	targetSevs := map[string]bool{"HIGH": true}

	bunches := GroupIssues(report.Issues, targetSevs)

	if len(bunches) != 1 {
		t.Fatalf("Expected 1 bunch for HIGH severity, got %d", len(bunches))
	}

	if len(bunches[0]) != 1 {
		t.Fatalf("Expected 1 issue in bunch, got %d", len(bunches[0]))
	}

	if bunches[0][0].Severity != "HIGH" {
		t.Errorf("Expected HIGH severity issue in bunch")
	}
}
