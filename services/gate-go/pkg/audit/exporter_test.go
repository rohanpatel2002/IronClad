package audit

import (
	"bytes"
	"strings"
	"testing"
)

func TestAuditExporter_ExportJSONAndCSV(t *testing.T) {
	exporter := NewAuditExporter()

	records := []LogRecord{
		{
			TenantID:      "t1",
			Actor:         "user1",
			Action:        "DEPLOY",
			ResourceType:  "service",
			ResourceID:    "svc-1",
			Status:        "ALLOW",
			IPAddress:     "10.0.0.1",
			UserAgent:     "cli/1.0",
			CorrelationID: "cid-100",
			Signature:     "sig123",
		},
	}

	// Test JSON export
	var jsonBuf bytes.Buffer
	if err := exporter.ExportJSON(&jsonBuf, records); err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), "user1") {
		t.Errorf("Expected JSON output to contain user1")
	}

	// Test CSV export
	var csvBuf bytes.Buffer
	if err := exporter.ExportCSV(&csvBuf, records); err != nil {
		t.Fatalf("Failed to export CSV: %v", err)
	}
	csvStr := csvBuf.String()
	if !strings.Contains(csvStr, "TenantID,Actor,Action") {
		t.Errorf("Expected CSV header in output")
	}
	if !strings.Contains(csvStr, "user1") {
		t.Errorf("Expected CSV body to contain user1")
	}
}
