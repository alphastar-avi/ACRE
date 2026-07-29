package snyk

import (
	"path/filepath"
	"sort"
	"strings"
)

// GroupIssues filters issues by target severities and groups similar vulnerabilities into bunches.
// Bunches are sorted by severity priority (CRITICAL > HIGH > MEDIUM > LOW).
func GroupIssues(issues []Issue, targetSeverities map[string]bool) [][]Issue {
	var filtered []Issue
	for _, issue := range issues {
		sev := strings.ToUpper(issue.Severity)
		if targetSeverities[sev] {
			filtered = append(filtered, issue)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	// Group key: "Title::ModuleDir"
	groupedMap := make(map[string][]Issue)
	var keys []string

	for _, issue := range filtered {
		modDir := extractModuleDir(issue.Path)
		key := issue.Title + "::" + modDir
		if _, exists := groupedMap[key]; !exists {
			keys = append(keys, key)
		}
		groupedMap[key] = append(groupedMap[key], issue)
	}

	maxBunchSize := 10
	var bunches [][]Issue

	for _, key := range keys {
		items := groupedMap[key]
		// Chunk large groups into bunches of at most maxBunchSize
		for i := 0; i < len(items); i += maxBunchSize {
			end := i + maxBunchSize
			if end > len(items) {
				end = len(items)
			}
			bunches = append(bunches, items[i:end])
		}
	}

	// Sort bunches by highest severity rank descending
	sort.SliceStable(bunches, func(i, j int) bool {
		sevI := 0
		if len(bunches[i]) > 0 {
			sevI = severityRank(bunches[i][0].Severity)
		}
		sevJ := 0
		if len(bunches[j]) > 0 {
			sevJ = severityRank(bunches[j][0].Severity)
		}
		return sevI > sevJ
	})

	return bunches
}

func severityRank(sev string) int {
	switch strings.ToUpper(sev) {
	case "CRITICAL":
		return 4
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	case "LOW":
		return 1
	default:
		return 0
	}
}

func extractModuleDir(pathStr string) string {
	clean := filepath.ToSlash(pathStr)
	parts := strings.Split(clean, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}
