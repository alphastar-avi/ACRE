package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// RunCommand executes a command inside the target repo and returns output.
func RunCommand(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command %s %v failed: %w (stderr: %s)", name, args, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// GetBaseBranch returns the name of the current active branch.
func GetBaseBranch(repoPath string) (string, error) {
	return RunCommand(repoPath, "git", "rev-parse", "--abbrev-ref", "HEAD")
}

// CheckoutBranch checks out the specified branch.
func CheckoutBranch(repoPath, branchName string) error {
	_, err := RunCommand(repoPath, "git", "checkout", branchName)
	return err
}

// CreateBranch creates a new branch and checks it out.
// If the branch already exists, it deletes it first to ensure a clean state.
func CreateBranch(repoPath, branchName, baseBranch string) error {
	// Switch to base first
	if err := CheckoutBranch(repoPath, baseBranch); err != nil {
		return err
	}
	// Try deleting if it exists
	_, _ = RunCommand(repoPath, "git", "branch", "-D", branchName)
	// Create and checkout new branch
	_, err := RunCommand(repoPath, "git", "checkout", "-b", branchName)
	return err
}

// CommitAndPush stages changes, commits, and pushes to remote origin.
func CommitAndPush(repoPath, branchName, commitMsg string) error {
	if _, err := RunCommand(repoPath, "git", "add", "-A"); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	// Commit. If no changes, commit might fail but we check for "nothing to commit"
	_, err := RunCommand(repoPath, "git", "commit", "-m", commitMsg)
	if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
		return fmt.Errorf("failed to commit: %w", err)
	}
	// Push to remote origin
	if _, err := RunCommand(repoPath, "git", "push", "-f", "origin", branchName); err != nil {
		return fmt.Errorf("failed to push branch to origin: %w", err)
	}
	return nil
}

// OwnerRepo represents parsed github owner and repository name.
type OwnerRepo struct {
	Owner string
	Repo  string
}

// GetOwnerRepo parses git remote URL to get owner and repository.
func GetOwnerRepo(repoPath string) (*OwnerRepo, error) {
	remoteURL, err := RunCommand(repoPath, "git", "remote", "get-url", "origin")
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote URL: %w", err)
	}

	// Match SSH: git@github.com:owner/repo.git
	// Match HTTPS: https://github.com/owner/repo.git or http://...
	re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/.]+)(?:\.git)?`)
	matches := re.FindStringSubmatch(remoteURL)
	if len(matches) < 3 {
		return nil, fmt.Errorf("unrecognized GitHub remote URL format: %s", remoteURL)
	}

	return &OwnerRepo{
		Owner: matches[1],
		Repo:  matches[2],
	}, nil
}

// CreatePR creates a GitHub Pull Request. If a GITHUB_TOKEN or GH_TOKEN is present,
// it uses the GitHub REST API. Otherwise, it generates a prefilled manual creation URL.
func CreatePR(repoPath, baseBranch, headBranch, title, body string) (string, string, error) {
	info, err := GetOwnerRepo(repoPath)
	if err != nil {
		return "", "", err
	}

	manualURL := fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s?expand=1&title=%s&body=%s",
		info.Owner,
		info.Repo,
		url.PathEscape(baseBranch),
		url.PathEscape(headBranch),
		url.QueryEscape(title),
		url.QueryEscape(body),
	)

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	if token == "" {
		return "", manualURL, nil
	}

	// Attempt API PR creation
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", info.Owner, info.Repo)
	payload := map[string]string{
		"title": title,
		"head":  headBranch,
		"base":  baseBranch,
		"body":  body,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", manualURL, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", manualURL, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", manualURL, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", manualURL, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var respJSON map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respJSON); err != nil {
		return "", manualURL, err
	}

	htmlURL, ok := respJSON["html_url"].(string)
	if !ok {
		return "", manualURL, fmt.Errorf("missing html_url in GitHub API response")
	}

	return htmlURL, manualURL, nil
}
