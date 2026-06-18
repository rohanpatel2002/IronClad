package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/retry"
	"github.com/sony/gobreaker"
)

// TopologyClient computes blast radius by delegating to the topology-go service.
type TopologyClient struct {
	topologyURL string
	httpClient  *http.Client
	cb          *gobreaker.CircuitBreaker
	redis       *redis.Client
}

// NewTopologyClient creates a new topology client with optional mTLS and Redis caching.
func NewTopologyClient(url string, tlsConfig *tls.Config, redisClient *redis.Client) *TopologyClient {
	st := gobreaker.Settings{
		Name:        "TopologyClient",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &TopologyClient{
		topologyURL: url,
		httpClient:  &http.Client{Timeout: 3 * time.Second, Transport: transport},
		cb:          gobreaker.NewCircuitBreaker(st),
		redis:       redisClient,
	}
}

type blastRadiusRequest struct {
	Service      string   `json:"service"`
	ChangedFiles []string `json:"changed_files"`
}

type blastRadiusResponse struct {
	Service          string   `json:"service"`
	BlastRadiusScore float64  `json:"blast_radius_score"`
	ImpactedServices []string `json:"impacted_services"`
}

// GetBlastRadius delegates the blast radius calculation to the topology service with Redis caching.
func (t *TopologyClient) GetBlastRadius(ctx context.Context, service string, changedFiles []string) (float64, []string, error) {
	// 1. Try Cache
	if t.redis != nil {
		cacheKey := fmt.Sprintf("blast_radius:%s:%v", service, changedFiles)
		val, err := t.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var res blastRadiusResponse
			if err := json.Unmarshal([]byte(val), &res); err == nil {
				return res.BlastRadiusScore, res.ImpactedServices, nil
			}
		}
	}

	reqBody := blastRadiusRequest{
		Service:      service,
		ChangedFiles: changedFiles,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return 0, nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		t.topologyURL+"/api/v1/blast-radius", bytes.NewReader(payload))
	if err != nil {
		return 0, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	respInterface, err := retry.DoWithExponentialBackoff(ctx, 3, 100*time.Millisecond, 2*time.Second, func() (interface{}, error) {
		return t.cb.Execute(func() (interface{}, error) {
			resp, reqErr := t.httpClient.Do(httpReq)
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
		return 0, nil, fmt.Errorf("topology service unavailable: %w", err)
	}

	httpResp := respInterface.(*http.Response)
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("topology service returned status %d", httpResp.StatusCode)
	}

	var res blastRadiusResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&res); err != nil {
		return 0, nil, err
	}

	// 2. Write Cache
	if t.redis != nil {
		cacheKey := fmt.Sprintf("blast_radius:%s:%v", service, changedFiles)
		data, _ := json.Marshal(res)
		t.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return res.BlastRadiusScore, res.ImpactedServices, nil
}

