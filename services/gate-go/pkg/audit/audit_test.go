package audit

import (
	"context"
	"testing"
)

func TestAuditLogger_DryRun(t *testing.T) {
	logger := NewAuditLogger(nil)
	rec := LogRecord{
		TenantID:      "tenant-1",
		Actor:         "user-admin",
		Action:        "POLICY_UPDATE",
		ResourceType:  "policy",
		ResourceID:    "pol-123",
		Status:        "SUCCESS",
		Details:       map[string]interface{}{"key": "val"},
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		CorrelationID: "corr-abc",
	}

	// Verify logging in dry-run mode does not panic
	logger.Log(context.Background(), rec)
}
