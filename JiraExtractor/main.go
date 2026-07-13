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

// Structs matching the Jira API response
type JiraAuthor struct {
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
}

type JiraComment struct {
	ID      string     `json:"id"`
	Author  JiraAuthor `json:"author"`
	Body    string     `json:"body"`
	Created string     `json:"created"`
}

type JiraCommentList struct {
	Comments []JiraComment `json:"comments"`
}

type JiraFields struct {
	Summary            string          `json:"summary"`
	Description        string          `json:"description"`
	AcceptanceCriteria string          `json:"customfield_11813"`
	Comment            JiraCommentList `json:"comment"`
}

type JiraIssueResponse struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Fields JiraFields `json:"fields"`
}

// Struct for the output JSON file
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

// findAndLoadEnv walks up the directory tree to find and parse the first .env file it encounters.
func findAndLoadEnv() string {
	// 1. Try starting from the current working directory
	dir, err := os.Getwd()
	if err == nil {
		for {
			// Check direct .env
			envPath := filepath.Join(dir, ".env")
			if _, err := os.Stat(envPath); err == nil {
				if err := parseEnvFile(envPath); err == nil {
					return envPath
				}
			}
			// Check orchestrator/.env
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

	// 2. Try adjacent to the running executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := parseEnvFile(envPath); err == nil {
				return envPath
			}
		}
		// Check orchestrator/.env adjacent to exe dir
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

func main() {
	ticketFlag := flag.String("ticket", "", "Jira ticket ID (e.g. FWS-66214)")
	flag.Parse()

	if *ticketFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -ticket <ID> is required.")
		flag.Usage()
		os.Exit(1)
	}

	ticketID := strings.TrimSpace(*ticketFlag)

	// Load env
	envPath := findAndLoadEnv()
	if envPath == "" {
		fmt.Fprintln(os.Stderr, "Warning: No .env file loaded. Using existing environment variables.")
	} else {
		fmt.Printf("Loaded environment from: %s\n", envPath)
	}

	// Get Token
	token := os.Getenv("JIRA_PAT")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: JIRA_PAT not found in environment variables.")
		os.Exit(1)
	}

	// Get Base URL
	baseUrl := strings.TrimRight(os.Getenv("JIRA_BASE_URL"), "/")
	if baseUrl == "" {
		fmt.Fprintln(os.Stderr, "Error: JIRA_BASE_URL not found in environment variables.")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	outputFileName := fmt.Sprintf("%s.json", ticketID)

	err := extractTicket(client, baseUrl, ticketID, token, outputFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// extractTicket performs the API call, processes fields, sorts comments, prints to stdout, and writes to outputFileName.
func extractTicket(client *http.Client, baseUrl string, ticketID string, token string, outputFileName string) error {
	// Build Jira API URL exactly as specified to avoid parsing issues
	issuePath := fmt.Sprintf("%s/rest/api/2/issue/%s", baseUrl, ticketID)
	query := "?fields=summary,description,comment,customfield_11813"
	uri := issuePath + query

	fmt.Printf("=== REQUEST URL ===\n%s\n\n", uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

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
		return fmt.Errorf("Jira API returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	var jiraIssue JiraIssueResponse
	if err := json.Unmarshal(respBody, &jiraIssue); err != nil {
		return fmt.Errorf("failed to parse Jira response: %w", err)
	}

	// Sort comments by created date (old -> new)
	comments := jiraIssue.Fields.Comment.Comments
	sort.Slice(comments, func(i, j int) bool {
		t1, err1 := time.Parse("2006-01-02T15:04:05.000-0700", comments[i].Created)
		t2, err2 := time.Parse("2006-01-02T15:04:05.000-0700", comments[j].Created)
		if err1 == nil && err2 == nil {
			return t1.Before(t2)
		}
		// Fallback to lexicographical comparison
		return comments[i].Created < comments[j].Created
	})

	// Construct structured output for stdout
	var commentOutputs []OutputComment
	var formattedComments strings.Builder
	for _, c := range comments {
		authorName := c.Author.DisplayName
		if authorName == "" {
			authorName = c.Author.Name
		}
		commentOutputs = append(commentOutputs, OutputComment{
			Created: c.Created,
			Author:  authorName,
			Body:    c.Body,
		})

		formattedComments.WriteString(fmt.Sprintf("%s - %s:\n%s\n", c.Created, authorName, c.Body))
	}

	// Print structured ticket in requested format
	fmt.Printf("=== SUMMARY ===\n%s\n\n", jiraIssue.Fields.Summary)
	fmt.Printf("=== DESCRIPTION ===\n%s\n\n", jiraIssue.Fields.Description)
	fmt.Printf("=== ACCEPTANCE CRITERIA ===\n%s\n\n", jiraIssue.Fields.AcceptanceCriteria)
	fmt.Printf("=== COMMENTS (old → new) ===\n%s", formattedComments.String())

	// Build and write OutputTicket JSON file
	outputObj := OutputTicket{
		TicketID:           jiraIssue.Key,
		Summary:            jiraIssue.Fields.Summary,
		Description:        jiraIssue.Fields.Description,
		AcceptanceCriteria: jiraIssue.Fields.AcceptanceCriteria,
		Comments:           commentOutputs,
	}

	jsonBytes, err := json.MarshalIndent(outputObj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output JSON: %w", err)
	}

	err = os.WriteFile(outputFileName, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", outputFileName, err)
	}

	fmt.Printf("\nSaved structured ticket JSON to: %s\n", outputFileName)
	return nil
}
