package audit

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

// AuditExporter provides utilities for exporting audit logs to standard formats.
type AuditExporter struct{}

// NewAuditExporter creates a new exporter.
func NewAuditExporter() *AuditExporter {
	return &AuditExporter{}
}

// ExportJSON writes audit records as formatted JSON.
func (e *AuditExporter) ExportJSON(w io.Writer, records []LogRecord) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

// ExportCSV writes audit records in CSV format.
func (e *AuditExporter) ExportCSV(w io.Writer, records []LogRecord) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"TenantID", "Actor", "Action", "ResourceType", "ResourceID", "Status", "IPAddress", "UserAgent", "CorrelationID", "Signature"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, r := range records {
		row := []string{
			r.TenantID,
			r.Actor,
			r.Action,
			r.ResourceType,
			r.ResourceID,
			r.Status,
			r.IPAddress,
			r.UserAgent,
			r.CorrelationID,
			r.Signature,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
