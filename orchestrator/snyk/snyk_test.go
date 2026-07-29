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

const userLaptopSampleOutput = `
Testing C:\Users\asivak976\OneDrive - Comcast\Desktop\Concepts\ACRE ...

Open Issues

 ✗ [LOW] Path Traversal
   Finding ID: 6c5bbe7d-f7f4-48da-989d-bceec70b0ef8
   Path: JiraExtractor/main.go, line 268
   Info: Unsanitized input from a CLI argument flows into os.WriteFile, where it is used as a path. This may result in a Path Traversal vulnerability and allow an attacker to write arbitrary files.

 ✗ [MEDIUM] Server-Side Request Forgery (SSRF)
   Finding ID: fa9dca22-95d9-487c-96a3-29b89acdb69d
   Path: JiraExtractor/main.go, line 191
   Info: Unsanitized input from a CLI argument flows into net.http.NewRequest, where it is used as an URL to perform a request. This may result in a Server-Side Request Forgery vulnerability.

╭─────────────────────────────────────────────────────────────────────────────────────╮
│ Test Summary                                                                        │
│                                                                                     │
│   Total issues:   2                                                                 │
│   Ignored issues: 0 [ 0 HIGH  0 MEDIUM  0 LOW ]                                     │
│   Open issues:    2 [ 0 HIGH  1 MEDIUM  1 LOW ]                                     │
╰─────────────────────────────────────────────────────────────────────────────────────╯
`

const ansiAndAbsPathSample = `
Testing C:\Users\asivak976\OneDrive - Comcast\Desktop\Concepts\ACRE ...

Open Issues

 \x1b[31m✗ [CRITICAL] SQL Injection\x1b[0m
   Finding ID: abc-123
   Path: C:\Users\asivak976\OneDrive - Comcast\Desktop\Concepts\ACRE\src\Data\Repository.cs:88
   Info: Unsanitized input used in SQL query execution.

 [HIGH] Use of Hardcoded Credentials
   Finding ID: def-456
   Path: src/Config/Secrets.cs [line 15]
   Info: Do not hardcode secrets in code.
`

func TestParseOutput(t *testing.T) {
	report := ParseOutput(sampleSnykOutput)

	if len(report.Issues) != 3 {
		t.Fatalf("Expected 3 issues parsed, got %d", len(report.Issues))
	}

	if report.TotalIssues != 259 {
		t.Errorf("Expected TotalIssues 259, got %d", report.TotalIssues)
	}

	lowIssue := report.Issues[0]
	if lowIssue.Path != "MDM/MDM.Web/Features/CentralLoginMigration/Hangfire/MigrateCentralLoginUsersJob.cs" {
		t.Errorf("Unexpected path: %s", lowIssue.Path)
	}
	if lowIssue.LineNumber != 64 {
		t.Errorf("Expected line 64, got %d", lowIssue.LineNumber)
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

func TestParseOutputUserSample(t *testing.T) {
	report := ParseOutput(userLaptopSampleOutput)

	if len(report.Issues) != 2 {
		t.Fatalf("Expected 2 issues parsed, got %d", len(report.Issues))
	}

	issue1 := report.Issues[0]
	if issue1.Path != "JiraExtractor/main.go" {
		t.Errorf("Expected path 'JiraExtractor/main.go', got '%s'", issue1.Path)
	}
	if issue1.LineNumber != 268 {
		t.Errorf("Expected line 268, got %d", issue1.LineNumber)
	}

	issue2 := report.Issues[1]
	if issue2.Path != "JiraExtractor/main.go" {
		t.Errorf("Expected path 'JiraExtractor/main.go', got '%s'", issue2.Path)
	}
	if issue2.LineNumber != 191 {
		t.Errorf("Expected line 191, got %d", issue2.LineNumber)
	}
	if issue2.Severity != "MEDIUM" {
		t.Errorf("Expected MEDIUM severity, got %s", issue2.Severity)
	}
}

func TestParseANSIAndAlternativeFormats(t *testing.T) {
	report := ParseOutput(ansiAndAbsPathSample)

	if len(report.Issues) != 2 {
		t.Fatalf("Expected 2 issues parsed, got %d", len(report.Issues))
	}

	crit := report.Issues[0]
	if crit.Severity != "CRITICAL" {
		t.Errorf("Expected CRITICAL severity, got %s", crit.Severity)
	}
	if crit.LineNumber != 88 {
		t.Errorf("Expected line 88, got %d", crit.LineNumber)
	}
	if crit.Path != "src/Data/Repository.cs" {
		t.Errorf("Expected trimmed relative path 'src/Data/Repository.cs', got '%s'", crit.Path)
	}

	high := report.Issues[1]
	if high.Severity != "HIGH" {
		t.Errorf("Expected HIGH severity, got %s", high.Severity)
	}
	if high.LineNumber != 15 {
		t.Errorf("Expected line 15, got %d", high.LineNumber)
	}
	if high.Path != "src/Config/Secrets.cs" {
		t.Errorf("Expected path 'src/Config/Secrets.cs', got '%s'", high.Path)
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
