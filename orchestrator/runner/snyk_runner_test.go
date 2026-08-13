package runner

import (
	"testing"

	"acre/snyk"
)

func TestMatchTargetFinding_FindingIDMatchWithLineShift(t *testing.T) {
	target := snyk.NormalizedFinding{
		FindingID: "965c5104-8706-4eb6-9d0b-34fc14e604a7",
		RuleID:    "csharp/Sqli",
		File:      "Repositories/PrintRepository.cs",
		Line:      402,
	}

	// Post scan has the finding, but line shifted from 402 to 450 after code edits
	remaining := []snyk.NormalizedFinding{
		{
			FindingID: "965c5104-8706-4eb6-9d0b-34fc14e604a7",
			RuleID:    "csharp/Sqli",
			File:      "Repositories/PrintRepository.cs",
			Line:      450,
		},
	}

	matched := make(map[int]bool)
	found, idx := matchTargetFinding(target, remaining, matched)

	if !found {
		t.Fatalf("Expected finding to be matched even with line shift from 402 to 450")
	}
	if idx != 0 {
		t.Fatalf("Expected match at index 0, got %d", idx)
	}
}

func TestMatchTargetFinding_ResolvedFinding(t *testing.T) {
	target := snyk.NormalizedFinding{
		FindingID: "965c5104-8706-4eb6-9d0b-34fc14e604a7",
		RuleID:    "csharp/Sqli",
		File:      "Repositories/PrintRepository.cs",
		Line:      402,
	}

	// Post scan contains only a different vulnerability
	remaining := []snyk.NormalizedFinding{
		{
			FindingID: "different-id-123",
			RuleID:    "csharp/HardcodedSecret",
			File:      "Config/Secrets.cs",
			Line:      50,
		},
	}

	matched := make(map[int]bool)
	found, _ := matchTargetFinding(target, remaining, matched)

	if found {
		t.Fatalf("Expected target finding to be marked as resolved (not found in remaining)")
	}
}

func TestMatchTargetFinding_FallbackRuleAndFileMatch(t *testing.T) {
	// Target has empty FindingID (legacy format)
	target := snyk.NormalizedFinding{
		FindingID: "",
		RuleID:    "csharp/Sqli",
		File:      "Repositories/PrintRepository.cs",
		Line:      402,
	}

	// Remaining finding in same file with shifted line
	remaining := []snyk.NormalizedFinding{
		{
			FindingID: "",
			RuleID:    "csharp/Sqli",
			File:      "Repositories/PrintRepository.cs",
			Line:      410,
		},
	}

	matched := make(map[int]bool)
	found, idx := matchTargetFinding(target, remaining, matched)

	if !found {
		t.Fatalf("Expected finding to match via fallback ruleId + file")
	}
	if idx != 0 {
		t.Fatalf("Expected match at index 0, got %d", idx)
	}
}
