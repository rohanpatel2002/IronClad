package services

import (
	"context"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

// NoopDeploymentRepository is an in-memory no-op that satisfies the
// DeploymentRepository interface until real Postgres persistence is wired.
type NoopDeploymentRepository struct {
	records map[string]*models.DeploymentRecord
}

// NewNoopDeploymentRepository creates a new no-op deployment repository.
func NewNoopDeploymentRepository() *NoopDeploymentRepository {
	return &NoopDeploymentRepository{records: make(map[string]*models.DeploymentRecord)}
}

func (r *NoopDeploymentRepository) Store(_ context.Context, record *models.DeploymentRecord) error {
	r.records[record.ID] = record
	return nil
}

func (r *NoopDeploymentRepository) Get(_ context.Context, id string) (*models.DeploymentRecord, error) {
	if rec, ok := r.records[id]; ok {
		return rec, nil
	}
	return nil, nil
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
