package snyk

import (
	"path/filepath"
	"strings"
)

// GroupIssues filters issues by target severities and groups similar vulnerabilities into bunches.
// Similarity is determined by finding title (vulnerability category) and codebase component/module path.
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

	// Map key: "Title::ModuleDir"
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

	return bunches
}

// extractModuleDir extracts top-level module directory from file path e.g. "MediaModule/..." -> "MediaModule"
func extractModuleDir(pathStr string) string {
	clean := filepath.ToSlash(pathStr)
	parts := strings.Split(clean, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}
