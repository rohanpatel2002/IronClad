package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/retry"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
	"github.com/sony/gobreaker"
)

// ScoringClient computes 3-axis risk scores either via the scoring-python
// microservice (when available) or via an embedded in-process scorer.
type ScoringClient struct {
	scoringURL string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

// NewScoringClient creates a new scoring client with sensible timeouts.
func NewScoringClient(url string) *ScoringClient {
	st := gobreaker.Settings{
		Name:        "ScoringClient",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
	}
	return &ScoringClient{
		scoringURL: url,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		cb:         gobreaker.NewCircuitBreaker(st),
	}
}

// ScoreDeployment computes risk scores. It first tries the scoring-python
// service; on failure it falls back to the embedded scorer.
func (s *ScoringClient) ScoreDeployment(ctx context.Context, req *services.ScoringRequest) (*services.ScoringResponse, error) {
	// Try remote scoring service first
	resp, err := s.callRemoteScorer(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Fallback: embedded in-process scorer
	return s.scoreInProcess(req), nil
}

// ---- Remote scorer --------------------------------------------------------

type remoteScoringRequest struct {
	Service      string   `json:"service"`
	CommitHash   string   `json:"commit_hash"`
	BlastRadius  float64  `json:"blast_radius"`
	ChangedFiles []string `json:"changed_files"`
	Environment  string   `json:"environment"`
	ServiceCrit  float64  `json:"service_criticality"`
	Intent       string   `json:"intent"`
}

type remoteScoringResponse struct {
	BlastRadius   float64  `json:"blast_radius_score"`
	Reversibility float64  `json:"reversibility_score"`
	TimingRisk    float64  `json:"timing_risk_score"`
	Confidence    float64  `json:"confidence"`
	Factors       []string `json:"factors"`
}

func (s *ScoringClient) callRemoteScorer(ctx context.Context, req *services.ScoringRequest) (*services.ScoringResponse, error) {
	payload, err := json.Marshal(remoteScoringRequest{
		Service:      req.Service,
		CommitHash:   req.CommitHash,
		BlastRadius:  req.BlastRadius,
		ChangedFiles: req.ChangedFiles,
		Environment:  req.Environment,
		ServiceCrit:  req.ServiceCrit,
		Intent:       req.Intent,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.scoringURL+"/api/v1/score", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	respInterface, err := retry.DoWithExponentialBackoff(ctx, 3, 100*time.Millisecond, 2*time.Second, func() (interface{}, error) {
		return s.cb.Execute(func() (interface{}, error) {
			resp, reqErr := s.httpClient.Do(httpReq)
			if reqErr != nil {
				return nil, reqErr
			}
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				return nil, fmt.Errorf("server error: %d", resp.StatusCode)
			}
			return resp, nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("scoring service unavailable (cb): %w", err)
	}

	httpResp := respInterface.(*http.Response)
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scoring service returned %d", httpResp.StatusCode)
	}

	var remote remoteScoringResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&remote); err != nil {
		return nil, err
	}

	return &services.ScoringResponse{
		BlastRadius:   remote.BlastRadius,
		Reversibility: remote.Reversibility,
		TimingRisk:    remote.TimingRisk,
		Confidence:    remote.Confidence,
		Factors:       remote.Factors,
	}, nil
}

// ---- Embedded in-process scorer ------------------------------------------

// scoreInProcess applies all three risk axes without any external calls.
func (s *ScoringClient) scoreInProcess(req *services.ScoringRequest) *services.ScoringResponse {
	var factors []string

	// Axis 1: Blast Radius (passed in from topology, adjusted by criticality)
	blastScore := req.BlastRadius * req.ServiceCrit
	if blastScore > 0.6 {
		factors = append(factors, fmt.Sprintf("High blast radius (%.0f%% of system affected)", blastScore*100))
	}

	// Axis 2: Reversibility — inferred from changed files and semantic intent
	reversibility, revFactors := scoreReversibility(req.ChangedFiles, req.Intent)
	factors = append(factors, revFactors...)

	// Axis 3: Timing Risk — based on current UTC time
	timingRisk, timeFactors := scoreTimingRisk(req.Environment)
	factors = append(factors, timeFactors...)

	// Environment modifier
	envMultiplier := environmentMultiplier(req.Environment)

	// Final weighted combination
	// Weights: blast_radius=0.40, reversibility=0.35, timing=0.25
	combined := (blastScore*0.40 + reversibility*0.35 + timingRisk*0.25) * envMultiplier
	combined = clamp01(combined)

	// Confidence: higher when scores are far from thresholds
	confidence := computeConfidence(blastScore, reversibility, timingRisk)

	return &services.ScoringResponse{
		BlastRadius:   clamp01(blastScore),
		Reversibility: clamp01(reversibility),
		TimingRisk:    clamp01(timingRisk),
		Confidence:    confidence,
		Factors:       factors,
	}
}

// scoreReversibility returns a 0-1 score based on file types changed and intent.
// High reversibility score = HARD to reverse (risky).
func scoreReversibility(changedFiles []string, intent string) (float64, []string) {
	if len(changedFiles) == 0 {
		return 0.3, nil // unknown → moderate
	}

	var score float64
	var factors []string

	migrationCount, configCount, codeCount, testCount := 0, 0, 0, 0

	for _, f := range changedFiles {
		lower := strings.ToLower(f)
		switch {
		case strings.Contains(lower, "migration") || strings.HasSuffix(lower, ".sql"):
			migrationCount++
		case strings.Contains(lower, "config") || strings.HasSuffix(lower, ".yaml") ||
			strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".env"):
			configCount++
		case strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, "_test.py") ||
			strings.HasSuffix(lower, ".test.ts"):
			testCount++
		default:
			codeCount++
		}
	}

	if intent == "migration" || migrationCount > 0 {
		score = 0.85 // DB migrations are very hard to reverse
		factors = append(factors, "Database migration intent/files detected — high irreversibility")
	} else if intent == "config_update" || (configCount > 0 && codeCount == 0) {
		score = 0.55 // config-only changes are moderate
		factors = append(factors, "Configuration update intent — moderate reversibility risk")
	} else if intent == "hotfix" {
		score = 0.80 // Hotfixes are usually rushed and risky to reverse
		factors = append(factors, "Hotfix intent — high risk, monitor closely")
	} else if testCount > 0 && codeCount == 0 && configCount == 0 {
		score = 0.1 // test-only is easily reversible
	} else {
		// Mixed code changes — score by file count
		score = math.Min(0.7, 0.2+float64(codeCount)*0.05)
		if codeCount > 5 {
			factors = append(factors, fmt.Sprintf("%d code files changed — broad surface area", codeCount))
		}
	}

	return score, factors
}

// scoreTimingRisk returns a 0-1 score based on current UTC time and day.
func scoreTimingRisk(environment string) (float64, []string) {
	now := time.Now().UTC()
	hour := now.Hour()
	weekday := now.Weekday()
	var factors []string
	var score float64

	// Production is always higher base risk
	baseProd := environment == "production"

	switch {
	case weekday == time.Friday && hour >= 14:
		score = 0.95
		factors = append(factors, "Friday afternoon deployment — extremely high risk window")
	case weekday == time.Saturday || weekday == time.Sunday:
		score = 0.70
		factors = append(factors, "Weekend deployment — reduced on-call coverage")
	case hour >= 22 || hour < 6:
		score = 0.75
		factors = append(factors, "Off-hours deployment (night) — reduced incident response capacity")
	case hour >= 17 && hour < 22:
		score = 0.45
		factors = append(factors, "Evening deployment — team partially available")
	case hour >= 10 && hour < 16:
		score = 0.15 // Business hours, peak team availability
	default:
		score = 0.30
	}

	if baseProd {
		score = clamp01(score * 1.2)
		factors = append(factors, "Production environment — additional scrutiny applied")
	}

	return score, factors
}

// environmentMultiplier returns a risk multiplier for the target environment.
func environmentMultiplier(env string) float64 {
	switch strings.ToLower(env) {
	case "production", "prod":
		return 1.3
	case "staging":
		return 1.0
	case "dev", "development":
		return 0.6
	default:
		return 1.0
	}
}

// computeConfidence returns how confident the scorer is in its output.
// Confidence is lower when scores cluster near threshold boundaries.
func computeConfidence(blast, rev, timing float64) float64 {
	// Distance from decision boundaries (0.6 and 0.8)
	avgScore := (blast + rev + timing) / 3.0
	distFromWarn := math.Abs(avgScore - 0.6)
	distFromBlock := math.Abs(avgScore - 0.8)
	minDist := math.Min(distFromWarn, distFromBlock)

	// Scale: near boundary → low confidence, far → high confidence
	confidence := 0.50 + minDist*1.5
	return clamp01(confidence)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
