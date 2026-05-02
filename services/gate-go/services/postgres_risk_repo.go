package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

// PostgresRiskScoreRepository implements RiskScoreRepository using PostgreSQL with read splitting.
type PostgresRiskScoreRepository struct {
	master  *sql.DB
	replica *sql.DB
}

// NewPostgresRiskScoreRepository creates a new Postgres risk score repo with read splitting.
func NewPostgresRiskScoreRepository(master, replica *sql.DB) *PostgresRiskScoreRepository {
	if replica == nil {
		replica = master
	}
	return &PostgresRiskScoreRepository{master: master, replica: replica}
}

// Store saves a new risk score record into the risk_scores table.
func (r *PostgresRiskScoreRepository) Store(ctx context.Context, record *models.RiskScoreRecord) error {
	query := `
		INSERT INTO risk_scores (
			id, deployment_id, blast_radius, reversibility, timing_risk,
			final_decision, computed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`
	_, err := r.master.ExecContext(ctx, query,
		record.ID, record.DeploymentID, record.BlastRadius, record.Reversibility,
		record.TimingRisk, record.FinalDecision, record.ComputedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store risk score: %w", err)
	}
	return nil
}
