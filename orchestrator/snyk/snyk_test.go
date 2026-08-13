package snyk

import (
	"encoding/json"
	"os"
	"strings"
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

func TestNormalizeSarifJSON(t *testing.T) {
	sampleSarif := `{
		"version": "2.1.0",
		"runs": [{
			"tool": {
				"driver": {
					"rules": [{
						"id": "csharp/PT",
						"shortDescription": { "text": "Path Traversal" },
						"properties": { "cwe": ["CWE-22"], "precision": "very-high" }
					}, {
						"id": "csharp/HardcodedSecret",
						"shortDescription": { "text": "Hardcoded Secret" },
						"properties": { "cwe": ["CWE-547"], "precision": "high" }
					}]
				}
			},
			"results": [{
				"ruleId": "csharp/PT",
				"level": "error",
				"message": { "text": "Unsanitized path used" },
				"locations": [{
					"physicalLocation": {
						"artifactLocation": { "uri": "Controllers/FileController.cs" },
						"region": { "startLine": 45 }
					}
				}]
			}, {
				"ruleId": "csharp/HardcodedSecret",
				"level": "error",
				"message": { "text": "Hardcoded API key" },
				"locations": [{
					"physicalLocation": {
						"artifactLocation": { "uri": "Config/Secrets.cs" },
						"region": { "startLine": 12 }
					}
				}]
			}]
		}]
	}`

	findings, err := NormalizeSarifJSON([]byte(sampleSarif), "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("Expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "csharp/PT" || f1.Title != "Path Traversal" || f1.File != "Controllers/FileController.cs" || f1.Line != 45 {
		t.Errorf("Unexpected finding 1 data: %+v", f1)
	}
	if len(f1.CWE) != 1 || f1.CWE[0] != "CWE-22" {
		t.Errorf("Unexpected CWE: %+v", f1.CWE)
	}

	// Test ruleId filter
	filtered, err := NormalizeSarifJSON([]byte(sampleSarif), "csharp/PT")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered finding, got %d", len(filtered))
	}
	if filtered[0].RuleID != "csharp/PT" {
		t.Errorf("Expected ruleId csharp/PT, got %s", filtered[0].RuleID)
	}
}

func TestNormalizeSarifJSON_FindingIDAndCodeFlow(t *testing.T) {
	sarifWithFlow := `{
		"version": "2.1.0",
		"runs": [{
			"tool": {
				"driver": {
					"rules": [{
						"id": "csharp/Sqli",
						"shortDescription": { "text": "SQL Injection" },
						"properties": { "cwe": ["CWE-89"], "precision": "very-high" }
					}]
				}
			},
			"results": [{
				"ruleId": "csharp/Sqli",
				"level": "warning",
				"message": { "text": "Unsanitized input flows into SQL command." },
				"locations": [{
					"physicalLocation": {
						"artifactLocation": { "uri": "Repositories/OrderRepository.cs" },
						"region": { "startLine": 402, "endLine": 402 }
					}
				}],
				"fingerprints": {
					"snyk/asset/finding/v1": "965c5104-8706-4eb6-9d0b-34fc14e604a7",
					"identity": "965c5104-8706-4eb6-9d0b-34fc14e604a7"
				},
				"codeFlows": [{
					"threadFlows": [{
						"locations": [
							{
								"location": {
									"physicalLocation": {
										"artifactLocation": { "uri": "Helpers/SqlHelper.cs" },
										"region": { "startLine": 105, "endLine": 105 }
									}
								}
							},
							{
								"location": {
									"physicalLocation": {
										"artifactLocation": { "uri": "Repositories/RepositoryHelper.cs" },
										"region": { "startLine": 138, "endLine": 138 }
									}
								}
							},
							{
								"location": {
									"physicalLocation": {
										"artifactLocation": { "uri": "Repositories/OrderRepository.cs" },
										"region": { "startLine": 316, "endLine": 321 }
									}
								}
							}
						]
					}]
				}]
			}]
		}]
	}`

	findings, err := NormalizeSarifJSON([]byte(sarifWithFlow), "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.FindingID != "965c5104-8706-4eb6-9d0b-34fc14e604a7" {
		t.Errorf("Expected findingId '965c5104-8706-4eb6-9d0b-34fc14e604a7', got '%s'", f.FindingID)
	}
	if f.RuleID != "csharp/Sqli" || f.Title != "SQL Injection" || f.File != "Repositories/OrderRepository.cs" || f.Line != 402 {
		t.Errorf("Unexpected finding metadata: %+v", f)
	}
	if len(f.CodeFlow) != 3 {
		t.Fatalf("Expected 3 codeFlow locations, got %d", len(f.CodeFlow))
	}

	if f.CodeFlow[0].File != "Helpers/SqlHelper.cs" || f.CodeFlow[0].StartLine != 105 || f.CodeFlow[0].EndLine != 105 {
		t.Errorf("Unexpected step 0: %+v", f.CodeFlow[0])
	}
	if f.CodeFlow[1].File != "Repositories/RepositoryHelper.cs" || f.CodeFlow[1].StartLine != 138 || f.CodeFlow[1].EndLine != 138 {
		t.Errorf("Unexpected step 1: %+v", f.CodeFlow[1])
	}
	if f.CodeFlow[2].File != "Repositories/OrderRepository.cs" || f.CodeFlow[2].StartLine != 316 || f.CodeFlow[2].EndLine != 321 {
		t.Errorf("Unexpected step 2: %+v", f.CodeFlow[2])
	}
}

func TestNormalizeSarifJSON_EmptyCodeFlowSerializesAsEmptyArray(t *testing.T) {
	finding := NormalizedFinding{
		RuleID:    "csharp/AntiforgeryTokenDisabled",
		Title:     "Anti-forgery token disabled",
		Level:     "note",
		Message:   "Action should use anti-forgery token",
		File:      "Controllers/HomeController.cs",
		Line:      25,
		CWE:       []string{"CWE-352"},
		Precision: "very-high",
		FindingID: "83b28cff-4963-4018-beb0-0b870fec115f",
		CodeFlow:  []CodeFlowLocation{},
	}

	bytes, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("Failed to marshal finding: %v", err)
	}

	jsonStr := string(bytes)
	if !strings.Contains(jsonStr, `"codeFlow":[]`) {
		t.Errorf("Expected JSON to contain '\"codeFlow\":[]', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"findingId":"83b28cff-4963-4018-beb0-0b870fec115f"`) {
		t.Errorf("Expected JSON to contain findingId, got: %s", jsonStr)
	}
}

func TestNormalizeSarifJSON_SampleFileIntegration(t *testing.T) {
	// Look for sample file in orchestrator directory
	samplePaths := []string{
		"../testdata_sample.json",
		"testdata_sample.json",
	}

	var data []byte
	var err error
	for _, p := range samplePaths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if err != nil || len(data) == 0 {
		t.Skip("testdata_sample.json not found on disk, skipping sample file integration test")
		return
	}

	findings, err := NormalizeSarifJSON(data, "")
	if err != nil {
		t.Fatalf("Failed to normalize real sample SARIF: %v", err)
	}

	if len(findings) != 16 {
		t.Fatalf("Expected 16 findings from sample JSON, got %d", len(findings))
	}

	// Verify Finding 0 (CSRF)
	f0 := findings[0]
	if f0.FindingID != "83b28cff-4963-4018-beb0-0b870fec115f" {
		t.Errorf("Expected finding 0 ID '83b28cff-4963-4018-beb0-0b870fec115f', got '%s'", f0.FindingID)
	}
	if f0.RuleID != "csharp/AntiforgeryTokenDisabled" {
		t.Errorf("Expected rule csharp/AntiforgeryTokenDisabled, got %s", f0.RuleID)
	}
	if len(f0.CodeFlow) != 1 {
		t.Errorf("Expected 1 codeFlow location in finding 0, got %d", len(f0.CodeFlow))
	}

	// Verify Finding 8 (SQLi)
	var sqliFinding *NormalizedFinding
	for idx := range findings {
		if findings[idx].FindingID == "965c5104-8706-4eb6-9d0b-34fc14e604a7" {
			sqliFinding = &findings[idx]
			break
		}
	}

	if sqliFinding == nil {
		t.Fatalf("Finding 965c5104-8706-4eb6-9d0b-34fc14e604a7 not found")
	}

	if sqliFinding.RuleID != "csharp/Sqli" {
		t.Errorf("Expected rule csharp/Sqli, got %s", sqliFinding.RuleID)
	}
	if sqliFinding.Title != "SQL Injection" {
		t.Errorf("Expected title 'SQL Injection', got '%s'", sqliFinding.Title)
	}
	if sqliFinding.File != "Sbms.Api.Repository.Print/PrintInsertionRepository.cs" {
		t.Errorf("Expected file 'Sbms.Api.Repository.Print/PrintInsertionRepository.cs', got '%s'", sqliFinding.File)
	}
	if sqliFinding.Line != 402 {
		t.Errorf("Expected line 402, got %d", sqliFinding.Line)
	}
	if len(sqliFinding.CodeFlow) != 20 {
		t.Errorf("Expected 20 deduplicated code flow locations for SQLi finding, got %d", len(sqliFinding.CodeFlow))
	}

	// Check first and last codeFlow locations
	firstLoc := sqliFinding.CodeFlow[0]
	if firstLoc.File != "Sbms.Api.Repository/SqlHelper.cs" || firstLoc.StartLine != 105 {
		t.Errorf("Unexpected first codeFlow loc: %+v", firstLoc)
	}

	lastLoc := sqliFinding.CodeFlow[len(sqliFinding.CodeFlow)-1]
	if lastLoc.File != "Sbms.Api.Repository/SqlHelper.cs" || lastLoc.StartLine != 88 {
		t.Errorf("Unexpected last codeFlow loc: %+v", lastLoc)
	}
}

