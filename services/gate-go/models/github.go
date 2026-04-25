package models

import "time"

// GitHubWebhookPayload represents the JSON payload from a GitHub pull_request event.
type GitHubWebhookPayload struct {
	Action      string     `json:"action"`
	Number      int        `json:"number"`
	PullRequest struct {
		URL    string `json:"url"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// GitHubPullRequestFile represents a file changed in a pull request.
type GitHubPullRequestFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
	Patch     string `json:"patch"`
}
