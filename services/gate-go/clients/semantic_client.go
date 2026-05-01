package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/retry"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
	"github.com/sony/gobreaker"
)

// SemanticClient connects to the semantic-python service to classify deployment intent.
type SemanticClient struct {
	semanticURL string
	httpClient  *http.Client
	cb          *gobreaker.CircuitBreaker
}

// NewSemanticClient creates a new semantic client with optional mTLS.
func NewSemanticClient(url string, tlsConfig *tls.Config) *SemanticClient {
	st := gobreaker.Settings{
		Name:        "SemanticClient",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &SemanticClient{
		semanticURL: url,
		httpClient:  &http.Client{Timeout: 5 * time.Second, Transport: transport},
		cb:          gobreaker.NewCircuitBreaker(st),
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

	respInterface, err := retry.DoWithExponentialBackoff(ctx, 3, 100*time.Millisecond, 2*time.Second, func() (interface{}, error) {
		return c.cb.Execute(func() (interface{}, error) {
			resp, reqErr := c.httpClient.Do(httpReq)
			if reqErr != nil {
				return nil, reqErr
			}
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				return nil, fmt.Errorf("server error: %d", resp.StatusCode)
			}
			return resp, nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("semantic service unavailable (cb): %w", err)
	}

	httpResp := respInterface.(*http.Response)
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
