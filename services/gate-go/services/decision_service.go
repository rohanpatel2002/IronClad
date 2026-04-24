package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
)

// DecisionService handles the core decision-making logic
type DecisionService struct {
	topologyClient     TopologyClient
	semanticClient     *clients.SemanticClient
	scoringClient      ScoringClient
	deploymentRepo     DeploymentRepository
	riskScoreRepo      RiskScoreRepository
}

// TopologyClient interface for dependency graph queries
type TopologyClient interface {
	GetBlastRadius(ctx context.Context, service string, changedFiles []string) (float64, []string, error)
}

// ScoringClient interface for risk scoring
type ScoringClient interface {
	ScoreDeployment(ctx context.Context, req *ScoringRequest) (*ScoringResponse, error)
}

// DeploymentRepository persists deployment records
type DeploymentRepository interface {
	Store(ctx context.Context, record *models.DeploymentRecord) error
	Get(ctx context.Context, id string) (*models.DeploymentRecord, error)
}

// RiskScoreRepository persists risk scores
type RiskScoreRepository interface {
	Store(ctx context.Context, record *models.RiskScoreRecord) error
}

// ScoringRequest for the scoring service
type ScoringRequest struct {
	Service        string
	CommitHash     string
	BlastRadius    float64
	ChangedFiles   []string
	Environment    string
	ServiceCrit    float64 // criticality 0-1
	Intent         string
}

// ScoringResponse from the scoring service
type ScoringResponse struct {
	BlastRadius   float64
	Reversibility float64
	TimingRisk    float64
	Confidence    float64
	Factors       []string
}

// NewDecisionService creates a new decision service
func NewDecisionService(
	topology TopologyClient,
	semantic *clients.SemanticClient,
	scoring ScoringClient,
	deployRepo DeploymentRepository,
	riskRepo RiskScoreRepository,
) *DecisionService {
	return &DecisionService{
		topologyClient:     topology,
		semanticClient:     semantic,
		scoringClient:      scoring,
		deploymentRepo:     deployRepo,
		riskScoreRepo:      riskRepo,
	}
}

// EvaluateDeployment evaluates a deployment request and returns a decision
func (ds *DecisionService) EvaluateDeployment(ctx context.Context, req *models.DeploymentRequest) (*models.DeploymentDecision, error) {
	decisionID := fmt.Sprintf("dec-%d-%d", time.Now().UnixNano(), rand.Int63())

	// Step 1: Get blast radius from topology
	blastRadius, impactedServices, err := ds.topologyClient.GetBlastRadius(ctx, req.Service, req.ChangedFiles)
	if err != nil {
		return nil, fmt.Errorf("topology check failed: %w", err)
	}

	// Step 2: Classify semantic intent
	intentReq := &clients.IntentRequest{
		Service:      req.Service,
		CommitHash:   req.CommitHash,
		Branch:       req.Branch,
		ChangedFiles: req.ChangedFiles,
	}
	intentRes, err := ds.semanticClient.ClassifyIntent(ctx, intentReq)
	if err != nil {
		// Log error and fallback to default intent
		fmt.Printf("Warning: semantic classification failed: %v\n", err)
		intentRes = &clients.IntentResponse{
			Intent:     "unknown",
			Confidence: 0.0,
			Reasoning:  "classification failed",
		}
	}

	// Step 3: Score the deployment
	scoringReq := &ScoringRequest{
		Service:      req.Service,
		CommitHash:   req.CommitHash,
		BlastRadius:  blastRadius,
		ChangedFiles: req.ChangedFiles,
		Environment:  req.Environment,
		ServiceCrit:  0.7, // TODO: fetch from config
		Intent:       intentRes.Intent,
	}

	scoreResp, err := ds.scoringClient.ScoreDeployment(ctx, scoringReq)
	if err != nil {
		return nil, fmt.Errorf("scoring failed: %w", err)
	}

	// Step 3: Determine decision based on scores
	decision := determineDecision(scoreResp)

	// Step 4: Generate explanation
	explanation := generateExplanation(decision, scoreResp, impactedServices)

	// Step 5: Suggest safe windows (if WARN or BLOCK)
	var safeWindows []models.TimeWindow
	if decision != models.DecisionAllow {
		safeWindows = suggestDeploymentWindows()
	}

	decisionTime := time.Now()
	result := &models.DeploymentDecision{
		DecisionID:           decisionID,
		Decision:             decision,
		RiskScores:           models.RiskScores{BlastRadius: scoreResp.BlastRadius, Reversibility: scoreResp.Reversibility, TimingRisk: scoreResp.TimingRisk, ComputedAt: decisionTime},
		Confidence:           scoreResp.Confidence,
		Explanation:          explanation,
		SuggestedSafeWindows: safeWindows,
		DecisionTimestamp:    decisionTime,
		Intent:               intentRes.Intent,
		IntentConfidence:     intentRes.Confidence,
	}

	// Step 6: Persist decision (non-blocking)
	go func() {
		deployRecord := &models.DeploymentRecord{
			ID:              decisionID,
			DeployTimestamp: time.Now(),
			CommitHash:      req.CommitHash,
			Branch:          req.Branch,
			ServiceName:     req.Service,
			AuthorEmail:     req.AuthorEmail,
			SemanticIntent:  intentRes.Intent,
			DecisionStatus:  string(decision),
			DecisionTime:    decisionTime,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		_ = ds.deploymentRepo.Store(context.Background(), deployRecord)

		riskRecord := &models.RiskScoreRecord{
			ID:            fmt.Sprintf("rs-%s", decisionID),
			DeploymentID:  decisionID,
			BlastRadius:   scoreResp.BlastRadius,
			Reversibility: scoreResp.Reversibility,
			TimingRisk:    scoreResp.TimingRisk,
			FinalDecision: string(decision),
			ComputedAt:    decisionTime,
		}
		_ = ds.riskScoreRepo.Store(context.Background(), riskRecord)
	}()

	return result, nil
}

// determineDecision maps risk scores to ALLOW/WARN/BLOCK
func determineDecision(scores *ScoringResponse) models.Decision {
	// Simple heuristic: if any axis is > 0.8, BLOCK; > 0.6, WARN; else ALLOW
	maxScore := max(scores.BlastRadius, scores.Reversibility, scores.TimingRisk)

	if maxScore > 0.8 {
		return models.DecisionBlock
	}
	if maxScore > 0.6 {
		return models.DecisionWarn
	}
	return models.DecisionAllow
}

// generateExplanation creates human-readable reasoning
func generateExplanation(decision models.Decision, scores *ScoringResponse, impactedServices []string) models.Explanation {
	var summary, mitigation string

	switch decision {
	case models.DecisionAllow:
		summary = "Safe to deploy: all risk axes within acceptable bounds"
		mitigation = "Monitor post-deploy for 1 hour"

	case models.DecisionWarn:
		summary = fmt.Sprintf("Moderate risk: review before deploying (max risk: %.0f%%)", max(scores.BlastRadius, scores.Reversibility, scores.TimingRisk)*100)
		mitigation = "Consider deploying during business hours with full incident response team available"

	case models.DecisionBlock:
		summary = fmt.Sprintf("High risk: do not deploy now (max risk: %.0f%%)", max(scores.BlastRadius, scores.Reversibility, scores.TimingRisk)*100)
		mitigation = "Wait for safer deployment window or reduce blast radius through staged rollout"
	}

	return models.Explanation{
		Summary:      summary,
		RiskFactors:  scores.Factors,
		Mitigations:  []string{mitigation},
		RelatedIncidents: impactedServices,
	}
}

// suggestDeploymentWindows recommends safe times to deploy
func suggestDeploymentWindows() []models.TimeWindow {
	now := time.Now()
	windows := []models.TimeWindow{
		{
			Start:      now.AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(10 * time.Hour),
			End:        now.AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(12 * time.Hour),
			Confidence: 0.92,
			Reason:     "Off-peak hours, full team available",
		},
		{
			Start:      now.AddDate(0, 0, 2).Truncate(24 * time.Hour).Add(14 * time.Hour),
			End:        now.AddDate(0, 0, 2).Truncate(24 * time.Hour).Add(16 * time.Hour),
			Confidence: 0.88,
			Reason:     "Mid-week, mid-day deployment",
		},
	}
	return windows
}

func max(values ...float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
