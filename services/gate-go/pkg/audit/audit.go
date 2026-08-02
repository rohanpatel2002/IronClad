package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// AuditLogger handles persistence of security-critical actions.
type AuditLogger struct {
	db      *sql.DB
	hmacKey []byte
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(db *sql.DB) *AuditLogger {
	secret := os.Getenv("AUDIT_HMAC_SECRET")
	if secret == "" {
		secret = "ironclad-audit-log-signature-key"
	}
	return NewAuditLoggerWithSecret(db, secret)
}

// NewAuditLoggerWithSecret creates a new audit logger with a specific HMAC secret.
func NewAuditLoggerWithSecret(db *sql.DB, hmacSecret string) *AuditLogger {
	return &AuditLogger{
		db:      db,
		hmacKey: []byte(hmacSecret),
	}
}

// LogRecord represents a single audit entry.
type LogRecord struct {
	TenantID      string                 `json:"tenant_id"`
	Actor         string                 `json:"actor"`
	Action        string                 `json:"action"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id"`
	Status        string                 `json:"status"`
	Details       map[string]interface{} `json:"details,omitempty"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	CorrelationID string                 `json:"correlation_id"`
	Signature     string                 `json:"signature,omitempty"`
}

// ComputeSignature calculates the HMAC-SHA256 checksum for tamper evidence.
func (l *AuditLogger) ComputeSignature(rec LogRecord) string {
	payload := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s",
		rec.TenantID, rec.Actor, rec.Action, rec.ResourceType,
		rec.ResourceID, rec.Status, rec.IPAddress, rec.UserAgent, rec.CorrelationID)

	h := hmac.New(sha256.New, l.hmacKey)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature checks if a record's signature matches its contents.
func (l *AuditLogger) VerifySignature(rec LogRecord) bool {
	if rec.Signature == "" {
		return false
	}
	expected := l.ComputeSignature(rec)
	return hmac.Equal([]byte(rec.Signature), []byte(expected))
}

// Log persists an audit record to the database.
func (l *AuditLogger) Log(ctx context.Context, rec LogRecord) {
	rec.Signature = l.ComputeSignature(rec)

	if l.db == nil {
		slog.Info("Audit log (dry-run)", "action", rec.Action, "actor", rec.Actor, "status", rec.Status, "sig", rec.Signature)
		return
	}

	detailsJSON, _ := json.Marshal(rec.Details)

	query := `
		INSERT INTO audit_logs (
			tenant_id, actor, action, resource_type, resource_id,
			status, details, ip_address, user_agent, correlation_id, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := l.db.ExecContext(ctx, query,
		rec.TenantID, rec.Actor, rec.Action, rec.ResourceType, rec.ResourceID,
		rec.Status, detailsJSON, rec.IPAddress, rec.UserAgent, rec.CorrelationID, rec.Signature,
	)
	if err != nil {
		slog.Error("Failed to persist audit log", "error", err, "action", rec.Action)
	}
}


