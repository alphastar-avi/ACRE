package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Jira Cloud (v3 REST API) with Basic Auth & Atlassian Document Format (ADF)
// ---------------------------------------------------------------------------

type JiraAuthor struct {
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
}

type JiraComment struct {
	ID      string          `json:"id"`
	Author  JiraAuthor      `json:"author"`
	Body    json.RawMessage `json:"body"`
	Created string          `json:"created"`
}

type JiraCommentList struct {
	Comments []JiraComment `json:"comments"`
}

type JiraFields struct {
	Summary            string          `json:"summary"`
	Description        json.RawMessage `json:"description"`
	AcceptanceCriteria json.RawMessage `json:"customfield_11813"`
	Comment            JiraCommentList `json:"comment"`
}

type JiraIssueResponse struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Fields JiraFields `json:"fields"`
}

type OutputTicket struct {
	TicketID           string          `json:"ticket_id"`
	Summary            string          `json:"summary"`
	Description        string          `json:"description"`
	AcceptanceCriteria string          `json:"acceptance_criteria"`
	Comments           []OutputComment `json:"comments"`
}

type OutputComment struct {
	Created string `json:"created"`
	Author  string `json:"author"`
	Body    string `json:"body"`
}

// ---------------------------------------------------------------------------
// ADF -> plain text
// ---------------------------------------------------------------------------

type adfNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text"`
	Content []adfNode `json:"content"`
}

func adfToPlainText(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
	}
	var node adfNode
	if err := json.Unmarshal(raw, &node); err != nil {
		return trimmed
	}
	var sb strings.Builder
	walkADF(&node, &sb)
	return strings.TrimSpace(sb.String())
}

func walkADF(n *adfNode, sb *strings.Builder) {
	if n == nil {
		return
	}
	if n.Type == "listItem" {
		sb.WriteString("- ")
	}
	if n.Text != "" {
		sb.WriteString(n.Text)
	}
	for i := range n.Content {
		walkADF(&n.Content[i], sb)
	}
	switch n.Type {
	case "paragraph", "heading", "listItem", "blockquote", "codeBlock", "hardBreak":
		sb.WriteString("\n")
	}
}

// ---------------------------------------------------------------------------
// .env loading
// ---------------------------------------------------------------------------

func findAndLoadEnv() string {
	dir, err := os.Getwd()
	if err == nil {
		for {
			envPath := filepath.Join(dir, ".env")
			if _, err := os.Stat(envPath); err == nil {
				if err := parseEnvFile(envPath); err == nil {
					return envPath
				}
			}
			orchEnvPath := filepath.Join(dir, "orchestrator", ".env")
			if _, err := os.Stat(orchEnvPath); err == nil {
				if err := parseEnvFile(orchEnvPath); err == nil {
					return orchEnvPath
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := parseEnvFile(envPath); err == nil {
				return envPath
			}
		}
		orchEnvPath := filepath.Join(exeDir, "orchestrator", ".env")
		if _, err := os.Stat(orchEnvPath); err == nil {
			if err := parseEnvFile(orchEnvPath); err == nil {
				return orchEnvPath
			}
		}
	}
	return ""
}

func parseEnvFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
	return nil
}

func getenvFirst(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	ticketFlag := flag.String("ticket", "", "Jira ticket ID (e.g. FWS-67484)")
	flag.Parse()

	if *ticketFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -ticket <ID> is required.")
		flag.Usage()
		os.Exit(1)
	}

	ticketID := strings.TrimSpace(*ticketFlag)

	envPath := findAndLoadEnv()
	if envPath == "" {
		fmt.Fprintln(os.Stderr, "Warning: No .env file loaded. Using existing environment variables.")
	} else {
		fmt.Printf("Loaded environment from: %s\n", envPath)
	}

	token := getenvFirst("JIRA_API_TOKEN", "JIRA_PAT")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: JIRA_API_TOKEN (or legacy JIRA_PAT) not found in environment variables.")
		os.Exit(1)
	}

	baseUrl := strings.TrimRight(getenvFirst("JIRA_BASE_URL"), "/")
	if baseUrl == "" {
		fmt.Fprintln(os.Stderr, "Error: JIRA_BASE_URL not found in environment variables.")
		os.Exit(1)
	}

	email := getenvFirst("JIRA_EMAIL")

	client := &http.Client{Timeout: 30 * time.Second}
	outputFileName := fmt.Sprintf("%s.json", ticketID)

	if err := extractTicket(client, baseUrl, email, ticketID, token, outputFileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func extractTicket(
	client *http.Client,
	baseUrl string,
	email string,
	ticketID string,
	token string,
	outputFileName string,
) error {

	issuePath := fmt.Sprintf("%s/rest/api/3/issue/%s", baseUrl, ticketID)
	query := "?fields=summary,description,comment,customfield_11813"
	uri := issuePath + query

	fmt.Printf("=== REQUEST URL ===\n%s\n\n", uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if email != "" {
		req.SetBasicAuth(email, token)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if v := resp.Header.Get("X-Seraph-Loginreason"); v != "" {
			fmt.Fprintf(os.Stderr, "X-Seraph-Loginreason: %s (check base URL / token)\n", v)
		}
		return fmt.Errorf("Jira API returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	var jiraIssue JiraIssueResponse
	if err := json.Unmarshal(respBody, &jiraIssue); err != nil {
		return fmt.Errorf("failed to parse Jira response: %w", err)
	}

	comments := jiraIssue.Fields.Comment.Comments
	sort.Slice(comments, func(i, j int) bool {
		t1, err1 := time.Parse("2006-01-02T15:04:05.000-0700", comments[i].Created)
		t2, err2 := time.Parse("2006-01-02T15:04:05.000-0700", comments[j].Created)
		if err1 == nil && err2 == nil {
			return t1.Before(t2)
		}
		return comments[i].Created < comments[j].Created
	})

	var commentOutputs []OutputComment
	var formattedComments strings.Builder

	for _, c := range comments {
		authorName := c.Author.DisplayName
		if authorName == "" {
			authorName = c.Author.Name
		}
		bodyText := adfToPlainText(c.Body)

		commentOutputs = append(commentOutputs, OutputComment{
			Created: c.Created,
			Author:  authorName,
			Body:    bodyText,
		})

		formattedComments.WriteString(fmt.Sprintf("%s - %s:\n%s\n", c.Created, authorName, bodyText))
	}

	descriptionText := adfToPlainText(jiraIssue.Fields.Description)
	acceptanceText := adfToPlainText(jiraIssue.Fields.AcceptanceCriteria)

	fmt.Printf("=== SUMMARY ===\n%s\n\n", jiraIssue.Fields.Summary)
	fmt.Printf("=== DESCRIPTION ===\n%s\n\n", descriptionText)
	fmt.Printf("=== ACCEPTANCE CRITERIA ===\n%s\n\n", acceptanceText)
	fmt.Printf("=== COMMENTS (old -> new) ===\n%s", formattedComments.String())

	outputObj := OutputTicket{
		TicketID:           jiraIssue.Key,
		Summary:            jiraIssue.Fields.Summary,
		Description:        descriptionText,
		AcceptanceCriteria: acceptanceText,
		Comments:           commentOutputs,
	}

	jsonBytes, err := json.MarshalIndent(outputObj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output JSON: %w", err)
	}

	if err := os.WriteFile(outputFileName, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", outputFileName, err)
	}

	fmt.Printf("\nSaved structured ticket JSON to: %s\n", outputFileName)
	return nil
}
