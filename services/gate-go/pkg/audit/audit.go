package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
)

// AuditLogger handles persistence of security-critical actions.
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// LogRecord represents a single audit entry.
type LogRecord struct {
	TenantID      string
	Actor         string
	Action        string
	ResourceType  string
	ResourceID    string
	Status        string
	Details       map[string]interface{}
	IPAddress     string
	UserAgent     string
	CorrelationID string
}

// Log persists an audit record to the database.
func (l *AuditLogger) Log(ctx context.Context, rec LogRecord) {
	if l.db == nil {
		slog.Info("Audit log (dry-run)", "action", rec.Action, "actor", rec.Actor, "status", rec.Status)
		return
	}

	detailsJSON, _ := json.Marshal(rec.Details)

	query := `
		INSERT INTO audit_logs (
			tenant_id, actor, action, resource_type, resource_id,
			status, details, ip_address, user_agent, correlation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := l.db.ExecContext(ctx, query,
		rec.TenantID, rec.Actor, rec.Action, rec.ResourceType, rec.ResourceID,
		rec.Status, detailsJSON, rec.IPAddress, rec.UserAgent, rec.CorrelationID,
	)
	if err != nil {
		slog.Error("Failed to persist audit log", "error", err, "action", rec.Action)
	}
}
