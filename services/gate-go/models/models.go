package models

import (
	"time"
)

// DeploymentRequest represents an incoming deployment decision request
type DeploymentRequest struct {
	CommitHash   string   `json:"commit_hash" binding:"required"`
	Service      string   `json:"service" binding:"required"`
	Branch       string   `json:"branch" binding:"required"`
	AuthorEmail  string   `json:"author_email"`
	DiffSize     int64    `json:"diff_size_bytes"`
	ChangedFiles []string `json:"changed_files"`
	Environment  string   `json:"environment" binding:"required"`
	ForceCheck   bool     `json:"force_check"`
}

// RiskScores represents the three-axis risk assessment
type RiskScores struct {
	BlastRadius    float64 `json:"blast_radius" binding:"min=0,max=1"`    // 0-1 scale
	Reversibility  float64 `json:"reversibility" binding:"min=0,max=1"`   // 0-1 scale
	TimingRisk     float64 `json:"timing_risk" binding:"min=0,max=1"`     // 0-1 scale
	ComputedAt     time.Time `json:"computed_at"`
}

// Decision represents the final gate decision
type Decision string

const (
	DecisionAllow Decision = "ALLOW"
	DecisionWarn  Decision = "WARN"
	DecisionBlock Decision = "BLOCK"
)

// Explanation provides human-readable reasoning for the decision
type Explanation struct {
	Summary             string   `json:"summary"`
	RiskFactors         []string `json:"risk_factors"`
	Mitigations         []string `json:"mitigations"`
	RelatedIncidents    []string `json:"related_incidents,omitempty"`
	HistoricalPrecedent string   `json:"historical_precedent,omitempty"`
}

// DeploymentDecision is the response from the gate
type DeploymentDecision struct {
	DecisionID             string        `json:"decision_id"`
	Decision               Decision      `json:"decision"`
	RiskScores             RiskScores    `json:"risk_scores"`
	Confidence             float64       `json:"confidence"`
	Explanation            Explanation   `json:"explanation"`
	SuggestedSafeWindows   []TimeWindow  `json:"suggested_safe_windows,omitempty"`
	DecisionTimestamp      time.Time     `json:"decision_timestamp"`
}

// TimeWindow represents a recommended deployment window
type TimeWindow struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Confidence float64   `json:"confidence"`
	Reason     string    `json:"reason,omitempty"`
}

// DeploymentRecord is the persistent record in the database
type DeploymentRecord struct {
	ID              string    `db:"id"`
	DeployTimestamp time.Time `db:"deploy_timestamp"`
	CommitHash      string    `db:"commit_hash"`
	Branch          string    `db:"branch"`
	ServiceName     string    `db:"service_name"`
	AuthorEmail     string    `db:"author_email"`
	DiffSummary     string    `db:"diff_summary"`
	DecisionStatus  string    `db:"decision_status"`
	DecisionTime    time.Time `db:"decision_timestamp"`
	Explanation     string    `db:"explanation"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// RiskScoreRecord is the persistent risk score in the database
type RiskScoreRecord struct {
	ID              string    `db:"id"`
	DeploymentID    string    `db:"deployment_id"`
	BlastRadius     float64   `db:"blast_radius"`
	Reversibility   float64   `db:"reversibility"`
	TimingRisk      float64   `db:"timing_risk"`
	FinalDecision   string    `db:"final_decision"`
	ComputedAt      time.Time `db:"computed_at"`
}
