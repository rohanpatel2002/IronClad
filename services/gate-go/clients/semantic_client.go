package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
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

// ClassifyIntent calls the semantic service to classify the deployment intent.
func (c *SemanticClient) ClassifyIntent(ctx context.Context, req *services.IntentRequest) (*services.IntentResponse, error) {
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

	var res services.IntentResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
