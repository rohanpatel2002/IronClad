package governance

import (
	"context"
	"testing"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/audit"
)

type mockRepo struct{}

func (m *mockRepo) ListByTimeRange(ctx context.Context, start, end time.Time) ([]*models.DeploymentRecord, error) {
	return []*models.DeploymentRecord{
		{ID: "dep-1", ServiceName: "payment-api", DecisionStatus: "ALLOW", DecisionTime: time.Now()},
		{ID: "dep-2", ServiceName: "user-service", DecisionStatus: "BLOCK", DecisionTime: time.Now()},
	}, nil
}

func TestReportGenerator(t *testing.T) {
	logger := audit.NewAuditLogger("secret", true)
	gen := NewReportGenerator(&mockRepo{}, logger)

	rep, err := gen.GenerateSOC2Summary(context.Background(), time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rep.EvidenceCount != 2 {
		t.Errorf("expected 2 evidence records, got %d", rep.EvidenceCount)
	}

	isoRep, err := gen.GenerateISO27001Summary(context.Background(), time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isoRep.EvidenceCount != 2 {
		t.Errorf("expected 2 evidence records for ISO report, got %d", isoRep.EvidenceCount)
	}
}
