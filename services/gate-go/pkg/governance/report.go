package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/audit"
)

// DeploymentRepository interface for querying historical data
type DeploymentRepository interface {
	ListByTimeRange(ctx context.Context, start, end time.Time) ([]*models.DeploymentRecord, error)
}

// ReportGenerator compiles audit data into regulatory compliance reports.
type ReportGenerator struct {
	repo   DeploymentRepository
	logger *audit.AuditLogger
}

// NewReportGenerator creates a new governance reporting tool.
func NewReportGenerator(repo DeploymentRepository, logger *audit.AuditLogger) *ReportGenerator {
	return &ReportGenerator{repo: repo, logger: logger}
}

// ComplianceReport represents a structured governance document.
type ComplianceReport struct {
	ReportID         string              `json:"report_id"`
	GeneratedAt      time.Time           `json:"generated_at"`
	Summary          string              `json:"summary"`
	EvidenceCount    int                 `json:"evidence_count"`
	AllowedDeploys   int                 `json:"allowed_deploys"`
	BlockedDeploys   int                 `json:"blocked_deploys"`
	ComplianceStatus string              `json:"compliance_status"`
	Records          []*models.DeploymentRecord `json:"records,omitempty"`
}

// GenerateSOC2Summary creates a summary of security actions for a given period.
func (g *ReportGenerator) GenerateSOC2Summary(ctx context.Context, start, end time.Time) (*ComplianceReport, error) {
	records, err := g.repo.ListByTimeRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records for report: %w", err)
	}

	var allowed, blocked int
	for _, r := range records {
		if r.DecisionStatus == "ALLOW" {
			allowed++
		} else if r.DecisionStatus == "BLOCK" {
			blocked++
		}
	}

	status := "PASSING"
	if blocked > 0 && float64(blocked)/float64(len(records)) > 0.2 {
		status = "REVIEW_REQUIRED"
	}

	report := &ComplianceReport{
		ReportID:         fmt.Sprintf("SOC2-%d", time.Now().Unix()),
		GeneratedAt:      time.Now().UTC(),
		Summary:          fmt.Sprintf("Security audit summary from %s to %s", start.Format(time.DateOnly), end.Format(time.DateOnly)),
		EvidenceCount:    len(records),
		AllowedDeploys:   allowed,
		BlockedDeploys:   blocked,
		ComplianceStatus: status,
		Records:          records,
	}

	return report, nil
}

// ExportAsJSON serializes the report for automated compliance ingestion.
func (r *ComplianceReport) ExportAsJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
