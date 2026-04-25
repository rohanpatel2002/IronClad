package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/clients"
	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

// WebhookHandler handles incoming webhooks from external systems like GitHub.
type WebhookHandler struct {
	svc           *services.DecisionService
	githubClient  *clients.GitHubClient
	webhookSecret []byte
}

// NewWebhookHandler creates a new handler.
func NewWebhookHandler(svc *services.DecisionService) *WebhookHandler {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		// Log warning, but proceed (useful for local dev without secrets)
	}
	return &WebhookHandler{
		svc:           svc,
		githubClient:  clients.NewGitHubClient(),
		webhookSecret: []byte(secret),
	}
}

// RegisterRoutes attaches webhook endpoints to the router group
func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/webhooks/github", h.handleGitHubWebhook)
}

// handleGitHubWebhook processes incoming GitHub PR/Push events.
func (h *WebhookHandler) handleGitHubWebhook(c *gin.Context) {
	// 1. Verify HMAC signature if a secret is configured
	if len(h.webhookSecret) > 0 {
		signature := c.GetHeader("X-Hub-Signature-256")
		if signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing X-Hub-Signature-256"})
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Validate signature
		if !h.verifySignature(signature, bodyBytes) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		// Restore the request body since we read it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 2. Parse the event type
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType != "pull_request" {
		// Ignore non-PR events for now
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "not a pull_request event"})
		return
	}

	// 3. Unmarshal the payload
	var payload models.GitHubWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// 4. We only care about opened or synchronize (new commits pushed) actions
	if payload.Action != "opened" && payload.Action != "synchronize" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "action not opened or synchronize"})
		return
	}

	// 5. Fetch changed files
	files, err := h.githubClient.GetPullRequestFiles(c.Request.Context(), payload.Repository.FullName, payload.Number)
	if err != nil {
		fmt.Printf("Failed to fetch PR files: %v\n", err)
		// Fallback to empty list
		files = []string{}
	}

	// 6. Evaluate deployment
	req := &models.DeploymentRequest{
		CommitHash:   payload.PullRequest.Head.SHA,
		Service:      payload.Repository.Name,
		Branch:       payload.PullRequest.Head.Ref,
		Environment:  "production", // Assume prod by default for PRs targeting main
		ChangedFiles: files,
		AuthorEmail:  payload.PullRequest.User.Login, // Github doesn't send email in this payload easily, use login
	}

	decision, err := h.svc.EvaluateDeployment(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "evaluation failed", "message": err.Error()})
		return
	}

	// 7. Post comment
	comment := formatDecisionComment(decision)
	_ = h.githubClient.PostPullRequestComment(c.Request.Context(), payload.Repository.FullName, payload.Number, comment)

	c.JSON(http.StatusOK, gin.H{
		"status":   "processed",
		"decision": decision.Decision,
	})
}

func formatDecisionComment(d *models.DeploymentDecision) string {
	icon := "✅"
	if d.Decision == models.DecisionWarn {
		icon = "⚠️"
	} else if d.Decision == models.DecisionBlock {
		icon = "❌"
	}

	return fmt.Sprintf("## %s IRONCLAD Deployment Decision: **%s**\n\n"+
		"**Summary:** %s\n\n"+
		"**Semantic Intent:** `%s` (Confidence: %.0f%%)\n\n"+
		"**Risk Factors:**\n%s\n",
		icon, d.Decision, d.Explanation.Summary, d.Intent, d.IntentConfidence*100, formatList(d.Explanation.RiskFactors))
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "- None"
	}
	res := ""
	for _, item := range items {
		res += fmt.Sprintf("- %s\n", item)
	}
	return res
}

// verifySignature checks the SHA-256 HMAC of the payload against the GitHub signature.
func (h *WebhookHandler) verifySignature(signature string, payload []byte) bool {
	const signaturePrefix = "sha256="
	
	if len(signature) < len(signaturePrefix) || signature[:len(signaturePrefix)] != signaturePrefix {
		return false
	}

	mac := hmac.New(sha256.New, h.webhookSecret)
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	actualSignature := signature[len(signaturePrefix):]
	
	return hmac.Equal([]byte(expectedSignature), []byte(actualSignature))
}
