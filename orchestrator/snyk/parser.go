package snyk

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Regex matching finding header e.g.: ✗ [LOW] Exposure of Private Personal Information...
	findingHeaderRegex = regexp.MustCompile(`(?i)(?:✗\s*)?\[(LOW|MEDIUM|HIGH|CRITICAL)\]\s+(.+)`)
	findingIDRegex     = regexp.MustCompile(`(?i)Finding ID:\s*([a-f0-9\-]+)`)
	pathWithLineRegex  = regexp.MustCompile(`(?i)^\s*Path:\s*(.+?),\s*line\s*(\d+)\s*$`)
	pathOnlyRegex      = regexp.MustCompile(`(?i)^\s*Path:\s*(.+)\s*$`)
	infoRegex          = regexp.MustCompile(`(?i)^\s*Info:\s*(.+)`)

	// Regex matching summary table counts
	totalIssuesRegex       = regexp.MustCompile(`(?i)Total issues:\s*(\d+)`)
	openIssuesRegex        = regexp.MustCompile(`(?i)Open issues:\s*(\d+)`)
	severityBreakdownRegex = regexp.MustCompile(`(?i)(\d+)\s*(HIGH|MEDIUM|LOW|CRITICAL)`)
)

// ParseOutput parses raw output string from `snyk code test` into a structured SnykReport.
func ParseOutput(rawOutput string) *SnykReport {
	report := &SnykReport{
		Counts: map[string]int{
			"CRITICAL": 0,
			"HIGH":     0,
			"MEDIUM":   0,
			"LOW":      0,
		},
		Issues:    []Issue{},
		RawOutput: rawOutput,
	}

	trimmed := strings.TrimSpace(rawOutput)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if parseJSONOutput(trimmed, report) {
			return report
		}
	}

	parsePlainTextOutput(rawOutput, report)
	return report
}

func parseJSONOutput(rawJSON string, report *SnykReport) bool {
	type JSONIssue struct {
		ID       string `json:"id"`
		Severity string `json:"severity"`
		Title    string `json:"title"`
		Rule     string `json:"rule"`
		FilePath string `json:"filePath"`
		Line     int    `json:"line"`
		Info     string `json:"info"`
		Msg      string `json:"message"`
	}
	type JSONResult struct {
		Runs []struct {
			Results []struct {
				RuleID  string `json:"ruleId"`
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
		Issues []JSONIssue `json:"issues"`
	}

	var res JSONResult
	if err := json.Unmarshal([]byte(rawJSON), &res); err == nil && (len(res.Issues) > 0 || len(res.Runs) > 0) {
		for _, jIssue := range res.Issues {
			sev := strings.ToUpper(jIssue.Severity)
			if sev == "" {
				sev = "MEDIUM"
			}
			report.Counts[sev]++
			info := jIssue.Info
			if info == "" {
				info = jIssue.Msg
			}
			report.Issues = append(report.Issues, Issue{
				FindingID:  jIssue.ID,
				Severity:   sev,
				Title:      jIssue.Title,
				Path:       jIssue.FilePath,
				LineNumber: jIssue.Line,
				Info:       info,
			})
		}
		for _, run := range res.Runs {
			for _, r := range run.Results {
				sev := "HIGH"
				report.Counts[sev]++
				path := ""
				line := 0
				if len(r.Locations) > 0 {
					path = r.Locations[0].PhysicalLocation.ArtifactLocation.URI
					line = r.Locations[0].PhysicalLocation.Region.StartLine
				}
				report.Issues = append(report.Issues, Issue{
					FindingID:  r.RuleID,
					Severity:   sev,
					Title:      r.RuleID,
					Path:       path,
					LineNumber: line,
					Info:       r.Message.Text,
				})
			}
		}
		report.TotalIssues = len(report.Issues)
		report.OpenIssues = len(report.Issues)
		return true
	}
	return false
}

func parsePlainTextOutput(rawOutput string, report *SnykReport) {
	lines := strings.Split(rawOutput, "\n")
	var currentIssue *Issue
	var rawBlock strings.Builder
	inInfo := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		trimmedLine := strings.TrimSpace(line)

		// Header match e.g. ✗ [LOW] Exposure of Private Personal Information...
		if headerMatch := findingHeaderRegex.FindStringSubmatch(trimmedLine); headerMatch != nil {
			if currentIssue != nil {
				currentIssue.RawText = rawBlock.String()
				report.Issues = append(report.Issues, *currentIssue)
				report.Counts[currentIssue.Severity]++
			}
			currentIssue = &Issue{
				Severity: strings.ToUpper(headerMatch[1]),
				Title:    strings.TrimSpace(headerMatch[2]),
			}
			rawBlock.Reset()
			rawBlock.WriteString(line + "\n")
			inInfo = false
			continue
		}

		if currentIssue != nil {
			rawBlock.WriteString(line + "\n")

			if match := findingIDRegex.FindStringSubmatch(trimmedLine); match != nil {
				currentIssue.FindingID = strings.TrimSpace(match[1])
				inInfo = false
			} else if match := pathWithLineRegex.FindStringSubmatch(trimmedLine); match != nil {
				currentIssue.Path = strings.TrimSpace(match[1])
				if lNum, err := strconv.Atoi(match[2]); err == nil {
					currentIssue.LineNumber = lNum
				}
				inInfo = false
			} else if match := pathOnlyRegex.FindStringSubmatch(trimmedLine); match != nil {
				// Only set Path if line starts with Path: and current issue path is empty
				if currentIssue.Path == "" {
					currentIssue.Path = strings.TrimSpace(match[1])
				}
				inInfo = false
			} else if match := infoRegex.FindStringSubmatch(trimmedLine); match != nil {
				currentIssue.Info = strings.TrimSpace(match[1])
				inInfo = true
			} else if inInfo && trimmedLine != "" && !strings.HasPrefix(trimmedLine, "╭") && !strings.HasPrefix(trimmedLine, "│") && !strings.HasPrefix(trimmedLine, "╰") && !strings.HasPrefix(trimmedLine, "💡") {
				currentIssue.Info += " " + trimmedLine
			}
		}

		// Parse Summary box at bottom
		if totalMatch := totalIssuesRegex.FindStringSubmatch(trimmedLine); totalMatch != nil {
			if count, err := strconv.Atoi(totalMatch[1]); err == nil {
				report.TotalIssues = count
			}
		}
		if openMatch := openIssuesRegex.FindStringSubmatch(trimmedLine); openMatch != nil {
			if count, err := strconv.Atoi(openMatch[1]); err == nil {
				report.OpenIssues = count
			}
		}
	}

	if currentIssue != nil {
		currentIssue.RawText = rawBlock.String()
		report.Issues = append(report.Issues, *currentIssue)
		report.Counts[currentIssue.Severity]++
	}

	if report.TotalIssues == 0 {
		report.TotalIssues = len(report.Issues)
	}
	if report.OpenIssues == 0 {
		report.OpenIssues = len(report.Issues)
	}

	// Parse breakdown from summary text if present e.g. Open issues: 259 [ 1 HIGH  7 MEDIUM  5 LOW ]
	matches := severityBreakdownRegex.FindAllStringSubmatch(rawOutput, -1)
	for _, m := range matches {
		if len(m) == 3 {
			count, err := strconv.Atoi(m[1])
			sev := strings.ToUpper(m[2])
			if err == nil {
				report.Counts[sev] = count
			}
		}
	}
}
