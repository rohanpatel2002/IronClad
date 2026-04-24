package services_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

// To run this test, you need a running postgres instance. 
// Skipping by default unless a flag or env var is set is good practice, 
// but we'll mock it or let it fail fast if no DB is available.

func TestPostgresRepositories(t *testing.T) {
	// Simple skip if no db url
	t.Skip("Skipping postgres integration test in unit test suite")
	
	db, err := sql.Open("postgres", "postgres://ironclad:ironclad_dev@localhost:5432/ironclad?sslmode=disable")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not running on localhost:5432, skipping test")
	}

	deployRepo := services.NewPostgresDeploymentRepository(db)
	riskRepo := services.NewPostgresRiskScoreRepository(db)

	ctx := context.Background()

	// 1. Test Deployment Repo
	deployRecord := &models.DeploymentRecord{
		ID:              "test-deploy-123",
		DeployTimestamp: time.Now(),
		CommitHash:      "abcdef",
		Branch:          "main",
		ServiceName:     "test-service",
		AuthorEmail:     "test@example.com",
		DiffSummary:     "Test summary",
		SemanticIntent:  "feature",
		DecisionStatus:  "ALLOW",
		DecisionTime:    time.Now(),
		Explanation:     "All good",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = deployRepo.Store(ctx, deployRecord)
	if err != nil {
		t.Fatalf("failed to store deployment: %v", err)
	}

	retrievedDeploy, err := deployRepo.Get(ctx, deployRecord.ID)
	if err != nil {
		t.Fatalf("failed to get deployment: %v", err)
	}
	if retrievedDeploy == nil {
		t.Fatalf("deployment not found")
	}
	if retrievedDeploy.CommitHash != "abcdef" {
		t.Errorf("expected commit_hash abcdef, got %s", retrievedDeploy.CommitHash)
	}

	// 2. Test Risk Score Repo
	riskRecord := &models.RiskScoreRecord{
		ID:            "test-risk-123",
		DeploymentID:  "test-deploy-123",
		BlastRadius:   0.5,
		Reversibility: 0.6,
		TimingRisk:    0.1,
		FinalDecision: "ALLOW",
		ComputedAt:    time.Now(),
	}

	err = riskRepo.Store(ctx, riskRecord)
	if err != nil {
		t.Fatalf("failed to store risk score: %v", err)
	}
}
