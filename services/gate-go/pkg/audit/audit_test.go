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

	logger.Log(context.Background(), rec)
}

func TestAuditLogger_HMACSignatureVerification(t *testing.T) {
	logger := NewAuditLoggerWithSecret(nil, "my-secret-key")

	rec := LogRecord{
		TenantID:      "tenant-alpha",
		Actor:         "user@example.com",
		Action:        "LOGIN",
		ResourceType:  "user_session",
		ResourceID:    "sess-999",
		Status:        "SUCCESS",
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		CorrelationID: "corr-12345",
	}

	sig := logger.ComputeSignature(rec)
	if sig == "" {
		t.Fatalf("Expected non-empty signature")
	}

	rec.Signature = sig
	if !logger.VerifySignature(rec) {
		t.Errorf("Expected signature verification to succeed")
	}

	// Tampered record test
	tamperedRec := rec
	tamperedRec.Actor = "hacker@example.com"
	if logger.VerifySignature(tamperedRec) {
		t.Errorf("Expected signature verification to fail for tampered record")
	}

	// Dry run Log invocation should not panic
	logger.Log(context.Background(), rec)
}

