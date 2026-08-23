package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

func TestSemanticClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(services.IntentResponse{
			Intent:     "feature",
			Confidence: 0.95,
			Reasoning:  "Mock test intent",
		})
	}))
	defer ts.Close()

	client := NewSemanticClient(ts.URL, nil)
	resp, err := client.ClassifyIntent(context.Background(), &services.IntentRequest{
		ServiceName: "payment-api",
		CommitHash:  "abc123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Intent != "feature" {
		t.Errorf("expected intent 'feature', got %s", resp.Intent)
	}
}
