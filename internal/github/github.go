package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rdark/za/internal/util"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    string    `json:"author"`
	Repo      string    `json:"repository"`
	Reviews   int       `json:"reviews"`
}

// Client handles GitHub CLI interactions
type Client struct {
	org string
}

// NewClient creates a new GitHub client
func NewClient(org string) *Client {
	return &Client{
		org: org,
	}
}

// IsAvailable checks if GitHub CLI is available
func IsAvailable() bool {
	result := util.ExecuteShellCommand("gh --version", 5*time.Second)
	return result.Error == nil && result.ExitCode == 0
}

// GetPRsCreatedYesterday fetches PRs created yesterday in the organization
func (c *Client) GetPRsCreatedYesterday(date time.Time) ([]PullRequest, error) {
	yesterday := date.AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.searchPRs(startOfDay, endOfDay, "")
}

// GetPRsOpenAndUnreviewed fetches PRs opened in the last 7 days that are still open and unreviewed
func (c *Client) GetPRsOpenAndUnreviewed(date time.Time) ([]PullRequest, error) {
	sevenDaysAgo := date.AddDate(0, 0, -7)
	startOfDay := time.Date(sevenDaysAgo.Year(), sevenDaysAgo.Month(), sevenDaysAgo.Day(), 0, 0, 0, 0, sevenDaysAgo.Location())

	return c.searchPRs(startOfDay, time.Time{}, "--state=open review:none")
}

// searchPRs searches for PRs using GitHub CLI
func (c *Client) searchPRs(createdAfter time.Time, createdBefore time.Time, additionalFilters string) ([]PullRequest, error) {
	// Build args array - use gh CLI flags instead of query string for better compatibility
	args := []string{
		"search",
		"prs",
		"--owner", c.org,
		"--author", "@me",
	}

	// Add date filters
	if !createdAfter.IsZero() {
		args = append(args, "--created", ">="+createdAfter.Format("2006-01-02"))
	}

	// Add additional filters (can be flags or query string parts)
	if additionalFilters != "" {
		// Split by space to handle multiple filters
		filters := strings.Fields(additionalFilters)
		args = append(args, filters...)
	}

	// Add JSON output and limit
	args = append(args,
		"--json", "number,title,url,state,createdAt,updatedAt,author,repository",
		"--limit", "100",
	)

	result := util.ExecuteCommand(util.ExecConfig{
		Command: "gh",
		Args:    args,
		Timeout: 30 * time.Second,
	})

	if result.Error != nil {
		return nil, fmt.Errorf("gh search failed: %w (exit code: %d, stderr: %s)", result.Error, result.ExitCode, result.Stderr)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("gh search exited with code %d: %s", result.ExitCode, result.Stderr)
	}

	// Parse JSON response
	var prs []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		URL       string `json:"url"`
		State     string `json:"state"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
		} `json:"repository"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &prs); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	// Convert to our PR format
	results := make([]PullRequest, 0, len(prs))
	for _, pr := range prs {
		createdAt, err := time.Parse(time.RFC3339, pr.CreatedAt)
		if err != nil {
			continue
		}
		updatedAt, err := time.Parse(time.RFC3339, pr.UpdatedAt)
		if err != nil {
			continue
		}

		results = append(results, PullRequest{
			Number:    pr.Number,
			Title:     pr.Title,
			URL:       pr.URL,
			State:     pr.State,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Author:    pr.Author.Login,
			Repo:      pr.Repository.NameWithOwner,
		})
	}

	return results, nil
}

// FormatPRsAsBulletPoints formats PRs as markdown bullet points
func FormatPRsAsBulletPoints(prs []PullRequest, needsReviewPrefix bool) string {
	if len(prs) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, pr := range prs {
		repoShort := pr.Repo
		if parts := strings.Split(pr.Repo, "/"); len(parts) == 2 {
			repoShort = parts[1]
		}

		prefix := ""
		if needsReviewPrefix {
			prefix = "needs-review: "
		}

		sb.WriteString(fmt.Sprintf("* %s[%s#%d](%s): %s\n", prefix, repoShort, pr.Number, pr.URL, pr.Title))
	}
	return sb.String()
}
