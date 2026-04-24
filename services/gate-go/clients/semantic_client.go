package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SemanticClient connects to the semantic-python service to classify deployment intent.
type SemanticClient struct {
	semanticURL string
	httpClient  *http.Client
}

// NewSemanticClient creates a new semantic client.
func NewSemanticClient(url string) *SemanticClient {
	return &SemanticClient{
		semanticURL: url,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// IntentRequest represents the data sent to the semantic service.
type IntentRequest struct {
	Service      string   `json:"service"`
	CommitHash   string   `json:"commit_hash"`
	Branch       string   `json:"branch"`
	ChangedFiles []string `json:"changed_files"`
	DiffSummary  string   `json:"diff_summary,omitempty"`
}

// IntentResponse represents the semantic classification result.
type IntentResponse struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// ClassifyIntent calls the semantic service to classify the deployment intent.
func (c *SemanticClient) ClassifyIntent(ctx context.Context, req *IntentRequest) (*IntentResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.semanticURL+"/api/v1/classify", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("semantic service unavailable: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("semantic service returned status: %d", httpResp.StatusCode)
	}

	var res IntentResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
