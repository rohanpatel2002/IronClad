package soar

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// QuarantineManager handles automated response to high-risk threats.
type QuarantineManager struct {
	opaURL string
	client *http.Client
}

// NewQuarantineManager creates a new SOAR manager.
func NewQuarantineManager(opaURL string) *QuarantineManager {
	return &QuarantineManager{
		opaURL: opaURL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// QuarantineService adds a service to the OPA blacklist dynamically.
func (m *QuarantineManager) QuarantineService(ctx context.Context, serviceID string, reason string) error {
	slog.Warn("AUTONOMOUS ACTION: Quarantining service", "service_id", serviceID, "reason", reason)

	// In a real world scenario, this would be a PUT to OPA's data API
	// or an update to a Redis blacklist that OPA consults.
	// We'll simulate the OPA data update here.
	url := fmt.Sprintf("%s/v1/data/ironclad/blacklist/%s", strings.TrimSuffix(m.opaURL, "/v1/data/ironclad/authz/allow"), serviceID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(`{"blocked": true}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update OPA blacklist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from OPA: %d", resp.StatusCode)
	}

	return nil
}

