package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

type GitHubClient struct {
	token      string
	httpClient *http.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		token:      os.Getenv("GITHUB_TOKEN"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetPullRequestFiles fetches the list of files changed in a PR.
func (c *GitHubClient) GetPullRequestFiles(ctx context.Context, repoFullName string, prNumber int) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d/files", repoFullName, prNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch PR files: status %d", resp.StatusCode)
	}

	var files []models.GitHubPullRequestFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Filename)
	}
	return fileNames, nil
}

// PostPullRequestComment posts a comment to the PR.
func (c *GitHubClient) PostPullRequestComment(ctx context.Context, repoFullName string, prNumber int, comment string) error {
	if c.token == "" {
		fmt.Printf("[GitHub] Would post comment to %s#%d:\n%s\n", repoFullName, prNumber, comment)
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments", repoFullName, prNumber)
	
	payload := map[string]string{"body": comment}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to post comment: status %d", resp.StatusCode)
	}

	return nil
}
