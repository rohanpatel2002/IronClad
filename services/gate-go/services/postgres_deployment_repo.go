package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

// PostgresDeploymentRepository implements DeploymentRepository using PostgreSQL with read splitting.
type PostgresDeploymentRepository struct {
	master  *sql.DB
	replica *sql.DB
}

// NewPostgresDeploymentRepository creates a new Postgres deployment repo with read splitting.
func NewPostgresDeploymentRepository(master, replica *sql.DB) *PostgresDeploymentRepository {
	if replica == nil {
		replica = master
	}
	return &PostgresDeploymentRepository{master: master, replica: replica}
}

// Store saves a new deployment record into the deployments table.
func (r *PostgresDeploymentRepository) Store(ctx context.Context, record *models.DeploymentRecord) error {
	query := `
		INSERT INTO deployments (
			id, deploy_timestamp, commit_hash, branch, service_name,
			author_email, diff_summary, semantic_intent, decision_status, decision_timestamp,
			explanation, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13
		)
	`
	_, err := r.master.ExecContext(ctx, query,
		record.ID, record.DeployTimestamp, record.CommitHash, record.Branch, record.ServiceName,
		record.AuthorEmail, record.DiffSummary, record.SemanticIntent, record.DecisionStatus, record.DecisionTime,
		record.Explanation, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store deployment: %w", err)
	}
	return nil
}

// Get retrieves a deployment record by ID.
func (r *PostgresDeploymentRepository) Get(ctx context.Context, id string) (*models.DeploymentRecord, error) {
	query := `
		SELECT 
			id, deploy_timestamp, commit_hash, branch, service_name,
			author_email, diff_summary, semantic_intent, decision_status, decision_timestamp,
			explanation, created_at, updated_at
		FROM deployments
		WHERE id = $1
	`
	row := r.replica.QueryRowContext(ctx, query, id)

	var rec models.DeploymentRecord
	// diff_summary, semantic_intent, and explanation can be null in the DB, so we use sql.NullString
	var diffSummary, semanticIntent, explanation, authorEmail sql.NullString
	// decision_timestamp can be null, but in Go models it's time.Time, so we use sql.NullTime
	var decisionTime sql.NullTime

	err := row.Scan(
		&rec.ID, &rec.DeployTimestamp, &rec.CommitHash, &rec.Branch, &rec.ServiceName,
		&authorEmail, &diffSummary, &semanticIntent, &rec.DecisionStatus, &decisionTime,
		&explanation, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	rec.AuthorEmail = authorEmail.String
	rec.DiffSummary = diffSummary.String
	rec.SemanticIntent = semanticIntent.String
	rec.Explanation = explanation.String
	if decisionTime.Valid {
		rec.DecisionTime = decisionTime.Time
	}

	return &rec, nil
}
