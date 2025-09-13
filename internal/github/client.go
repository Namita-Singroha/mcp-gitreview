package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TokenProvider func() string

type Client struct {
	TokenProvider TokenProvider
}

// NewClient creates a new GitHub client with optional token
func NewClientFunc(tp TokenProvider) *Client {
	return &Client{TokenProvider: tp}
}

// createAuthenticatedRequest builds an HTTP request with Authorization header if token exists
func (c *Client) createAuthenticatedRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if c.TokenProvider != nil {
		token := c.TokenProvider()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	return req, nil
}

// doRequest performs an HTTP request with authentication and JSON decoding
func (c *Client) doRequest(method, url string, body io.Reader, out interface{}) error {
	req, err := c.createAuthenticatedRequest(method, url, body)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: status %d, body: %s", resp.StatusCode, string(raw))
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// FetchCommits fetches last 10 commits
func (c *Client) FetchCommits(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits", repo)
	var commits []Commit
	if err := c.doRequest("GET", url, nil, &commits); err != nil {
		return "", err
	}

	if len(commits) > 10 {
		commits = commits[:10]
	}

	result := "Last 10 commits:\n"
	for _, cm := range commits {
		result += fmt.Sprintf("%s - %s (%s)\n", cm.SHA[:7], cm.Commit.Message, cm.Commit.Author.Name)
	}
	return result, nil
}

// FetchBranches fetches all branches
func (c *Client) FetchBranches(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/branches", repo)
	var branches []Branch
	if err := c.doRequest("GET", url, nil, &branches); err != nil {
		return "", err
	}

	result := "Branches:\n"
	for _, b := range branches {
		result += b.Name + "\n"
	}
	return result, nil
}

// FetchLatestCommitMessage fetches the latest commit message
func (c *Client) FetchLatestCommitMessage(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits", repo)
	var commits []Commit
	if err := c.doRequest("GET", url, nil, &commits); err != nil {
		return "", err
	}

	if len(commits) == 0 {
		return "No commits found", nil
	}

	return fmt.Sprintf("Latest commit: %s - %s", commits[0].SHA[:7], commits[0].Commit.Message), nil
}

// FetchRepoInfo fetches repository metadata (description, stars, forks, issues)
func (c *Client) FetchRepoInfo(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	var data struct {
		Description string `json:"description"`
		Stars       int    `json:"stargazers_count"`
		Forks       int    `json:"forks_count"`
		Issues      int    `json:"open_issues_count"`
	}
	if err := c.doRequest("GET", url, nil, &data); err != nil {
		return "", err
	}

	result := fmt.Sprintf(
		"Description: %s\nStars: %d ⭐\nForks: %d 🍴\nOpen Issues: %d 🐞",
		data.Description, data.Stars, data.Forks, data.Issues,
	)
	return result, nil
}

// FetchPullRequests fetches open pull requests
func (c *Client) FetchPullRequests(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=open", repo)
	var prs []struct {
		Title string `json:"title"`
		User  struct {
			Login string `json:"login"`
		} `json:"user"`
		HTMLURL string `json:"html_url"`
	}
	if err := c.doRequest("GET", url, nil, &prs); err != nil {
		return "", err
	}

	if len(prs) == 0 {
		return "No open pull requests", nil
	}

	result := "Open Pull Requests:\n"
	for _, pr := range prs {
		result += fmt.Sprintf("- %s (by %s) → %s\n", pr.Title, pr.User.Login, pr.HTMLURL)
	}
	return result, nil
}
