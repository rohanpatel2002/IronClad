package services

import (
	"context"
	"sync"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

// NoopDeploymentRepository is a thread-safe in-memory no-op that satisfies the
// DeploymentRepository interface until real Postgres persistence is wired.
type NoopDeploymentRepository struct {
	mu      sync.RWMutex
	records map[string]*models.DeploymentRecord
}

// NewNoopDeploymentRepository creates a new no-op deployment repository.
func NewNoopDeploymentRepository() *NoopDeploymentRepository {
	return &NoopDeploymentRepository{records: make(map[string]*models.DeploymentRecord)}
}

func (r *NoopDeploymentRepository) Store(_ context.Context, record *models.DeploymentRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.ID] = record
	return nil
}

func (r *NoopDeploymentRepository) Get(_ context.Context, id string) (*models.DeploymentRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rec, ok := r.records[id]; ok {
		return rec, nil
	}
	return nil, nil
}

func (r *NoopDeploymentRepository) ListByTimeRange(_ context.Context, start, end time.Time) ([]*models.DeploymentRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.DeploymentRecord
	for _, rec := range r.records {
		if (rec.DeployTimestamp.After(start) || rec.DeployTimestamp.Equal(start)) &&
			(rec.DeployTimestamp.Before(end) || rec.DeployTimestamp.Equal(end)) {
			result = append(result, rec)
		}
	}
	return result, nil
}

// NoopRiskScoreRepository is an in-memory no-op that satisfies the
// RiskScoreRepository interface until real Postgres persistence is wired.
type NoopRiskScoreRepository struct{}

// NewNoopRiskScoreRepository creates a new no-op risk score repository.
func NewNoopRiskScoreRepository() *NoopRiskScoreRepository {
	return &NoopRiskScoreRepository{}
}

func (r *NoopRiskScoreRepository) Store(_ context.Context, _ *models.RiskScoreRecord) error {
	return nil
}
