package services_test

import (
	"context"
	"testing"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

// ---- Mock implementations ------------------------------------------------

type mockTopologyClient struct {
	blastRadius      float64
	impactedServices []string
	err              error
}

func (m *mockTopologyClient) GetBlastRadius(_ context.Context, _ string, _ []string) (float64, []string, error) {
	return m.blastRadius, m.impactedServices, m.err
}

type mockScoringClient struct {
	response *services.ScoringResponse
	err      error
}

func (m *mockScoringClient) ScoreDeployment(_ context.Context, _ *services.ScoringRequest) (*services.ScoringResponse, error) {
	return m.response, m.err
}

type mockSemanticClient struct {
	response *services.IntentResponse
	err      error
}

func (m *mockSemanticClient) ClassifyIntent(_ context.Context, _ *services.IntentRequest) (*services.IntentResponse, error) {
	return m.response, m.err
}

// makeService builds a DecisionService with mock dependencies
func makeService(blastRadius float64, impacted []string, scores *services.ScoringResponse) *services.DecisionService {
	topology := &mockTopologyClient{blastRadius: blastRadius, impactedServices: impacted}
	scoring := &mockScoringClient{response: scores}
	semantic := &mockSemanticClient{response: &services.IntentResponse{Intent: "feature", Confidence: 0.9, Reasoning: "test"}}
	return services.NewDecisionService(
		topology,
		semantic,
		scoring,
		services.NewNoopDeploymentRepository(),
		services.NewNoopRiskScoreRepository(),
	)
}

// ---- Decision threshold tests -------------------------------------------

func TestEvaluateDeployment_LowRisk_ReturnsAllow(t *testing.T) {
	svc := makeService(0.1, []string{"notification-service"}, &services.ScoringResponse{
		BlastRadius:   0.1,
		Reversibility: 0.2,
		TimingRisk:    0.15,
		Confidence:    0.92,
		Factors:       []string{},
	})

	req := &models.DeploymentRequest{
		CommitHash:   "abc123",
		Service:      "notification-service",
		Branch:       "main",
		Environment:  "staging",
		ChangedFiles: []string{"handler.go"},
	}

	decision, err := svc.EvaluateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != models.DecisionAllow {
		t.Errorf("expected ALLOW, got %s (scores: blast=%.2f, rev=%.2f, timing=%.2f)",
			decision.Decision, decision.RiskScores.BlastRadius, decision.RiskScores.Reversibility, decision.RiskScores.TimingRisk)
	}

	if decision.DecisionID == "" {
		t.Error("expected non-empty decision ID")
	}
}

func TestEvaluateDeployment_ModerateRisk_ReturnsWarn(t *testing.T) {
	svc := makeService(0.65, []string{"auth-service", "user-service"}, &services.ScoringResponse{
		BlastRadius:   0.65,
		Reversibility: 0.55,
		TimingRisk:    0.45,
		Confidence:    0.70,
		Factors:       []string{"Config changes detected"},
	})

	req := &models.DeploymentRequest{
		CommitHash:   "def456",
		Service:      "auth-service",
		Branch:       "feature/auth-refactor",
		Environment:  "staging",
		ChangedFiles: []string{"config.yaml", "auth.go"},
	}

	decision, err := svc.EvaluateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != models.DecisionWarn {
		t.Errorf("expected WARN, got %s", decision.Decision)
	}

	if len(decision.SuggestedSafeWindows) == 0 {
		t.Error("expected suggested safe windows for WARN decision")
	}
}

func TestEvaluateDeployment_HighRisk_ReturnsBlock(t *testing.T) {
	svc := makeService(0.95, []string{"payment-api", "order-service", "api-gateway"}, &services.ScoringResponse{
		BlastRadius:   0.95,
		Reversibility: 0.90,
		TimingRisk:    0.85,
		Confidence:    0.88,
		Factors:       []string{"Database migration detected", "Friday afternoon"},
	})

	req := &models.DeploymentRequest{
		CommitHash:   "ghi789",
		Service:      "database-primary",
		Branch:       "hotfix/schema-change",
		Environment:  "production",
		ChangedFiles: []string{"migrations/001_add_columns.sql"},
	}

	decision, err := svc.EvaluateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != models.DecisionBlock {
		t.Errorf("expected BLOCK, got %s", decision.Decision)
	}

	if len(decision.SuggestedSafeWindows) == 0 {
		t.Error("expected suggested safe windows for BLOCK decision")
	}
}

// ---- Table-driven threshold tests ----------------------------------------

func TestDetermineDecision_Thresholds(t *testing.T) {
	cases := []struct {
		name     string
		scores   *services.ScoringResponse
		expected models.Decision
	}{
		{
			name:     "all zero scores → ALLOW",
			scores:   &services.ScoringResponse{BlastRadius: 0, Reversibility: 0, TimingRisk: 0},
			expected: models.DecisionAllow,
		},
		{
			name:     "max 0.59 → ALLOW",
			scores:   &services.ScoringResponse{BlastRadius: 0.59, Reversibility: 0.3, TimingRisk: 0.2},
			expected: models.DecisionAllow,
		},
		{
			name:     "max 0.61 → WARN",
			scores:   &services.ScoringResponse{BlastRadius: 0.61, Reversibility: 0.3, TimingRisk: 0.2},
			expected: models.DecisionWarn,
		},
		{
			name:     "max 0.79 → WARN",
			scores:   &services.ScoringResponse{BlastRadius: 0.4, Reversibility: 0.79, TimingRisk: 0.3},
			expected: models.DecisionWarn,
		},
		{
			name:     "max 0.81 → BLOCK",
			scores:   &services.ScoringResponse{BlastRadius: 0.81, Reversibility: 0.5, TimingRisk: 0.4},
			expected: models.DecisionBlock,
		},
		{
			name:     "all max → BLOCK",
			scores:   &services.ScoringResponse{BlastRadius: 1.0, Reversibility: 1.0, TimingRisk: 1.0},
			expected: models.DecisionBlock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := makeService(tc.scores.BlastRadius, nil, tc.scores)
			req := &models.DeploymentRequest{
				CommitHash:  "test",
				Service:     "test-service",
				Branch:      "main",
				Environment: "staging",
			}
			decision, err := svc.EvaluateDeployment(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decision.Decision != tc.expected {
				t.Errorf("expected %s, got %s (blast=%.2f, rev=%.2f, timing=%.2f)",
					tc.expected, decision.Decision,
					tc.scores.BlastRadius, tc.scores.Reversibility, tc.scores.TimingRisk)
			}
		})
	}
}

// ---- Explanation tests ---------------------------------------------------

func TestDecisionExplanation_HasRequiredFields(t *testing.T) {
	svc := makeService(0.9, []string{"payment-api"}, &services.ScoringResponse{
		BlastRadius:   0.9,
		Reversibility: 0.85,
		TimingRisk:    0.8,
		Confidence:    0.88,
		Factors:       []string{"Critical service affected"},
	})

	req := &models.DeploymentRequest{
		CommitHash:  "xyz999",
		Service:     "payment-api",
		Branch:      "main",
		Environment: "production",
	}

	decision, err := svc.EvaluateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Explanation.Summary == "" {
		t.Error("explanation summary must not be empty")
	}
	if len(decision.Explanation.Mitigations) == 0 {
		t.Error("explanation must include at least one mitigation")
	}
	if decision.Confidence <= 0 || decision.Confidence > 1 {
		t.Errorf("confidence must be in (0,1], got %.2f", decision.Confidence)
	}
}

// ---- Decision ID uniqueness test ----------------------------------------

func TestDecisionIDs_AreUnique(t *testing.T) {
	svc := makeService(0.1, nil, &services.ScoringResponse{
		BlastRadius: 0.1, Reversibility: 0.1, TimingRisk: 0.1, Confidence: 0.9,
	})
	req := &models.DeploymentRequest{
		CommitHash: "abc", Service: "svc", Branch: "main", Environment: "staging",
	}

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		d, _ := svc.EvaluateDeployment(context.Background(), req)
		if ids[d.DecisionID] {
			t.Errorf("duplicate decision ID: %s", d.DecisionID)
		}
		ids[d.DecisionID] = true
	}
}
