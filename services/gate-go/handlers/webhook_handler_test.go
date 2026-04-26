package handlers_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

const testWebhookSecret = "ironclad-test-secret-key"

func signPayload(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func setupWebhookRouter(t *testing.T) *gin.Engine {
	t.Helper()
	t.Setenv("GITHUB_WEBHOOK_SECRET", testWebhookSecret)
	gin.SetMode(gin.TestMode)

	// We pass nil DecisionService — webhook should reject before reaching it
	wh := handlers.NewWebhookHandler((*services.DecisionService)(nil))
	router := gin.New()
	router.Use(handlers.PrometheusMiddleware())
	v1 := router.Group("/api/v1")
	wh.RegisterRoutes(v1)
	return router
}

func TestWebhook_RejectsRequestWithNoSignature(t *testing.T) {
	router := setupWebhookRouter(t)

	body := `{"action":"opened"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	// Deliberately omit X-Hub-Signature-256

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWebhook_RejectsInvalidSignature(t *testing.T) {
	router := setupWebhookRouter(t)

	body := `{"action":"opened"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid sig, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWebhook_IgnoresNonPREvents(t *testing.T) {
	router := setupWebhookRouter(t)

	body := `{"action":"push"}`
	sig := signPayload(testWebhookSecret, body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push") // Not pull_request
	req.Header.Set("X-Hub-Signature-256", sig)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for non-PR event, got %d", w.Code)
	}
}

func TestWebhook_IgnoresNonActionableActions(t *testing.T) {
	router := setupWebhookRouter(t)

	body := `{"action":"closed","pull_request":{"head":{"ref":"main","sha":"abc"},"user":{"login":"user"}},"repository":{"name":"repo","full_name":"org/repo"},"number":1}`
	sig := signPayload(testWebhookSecret, body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", sig)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for ignored action, got %d", w.Code)
	}
}
