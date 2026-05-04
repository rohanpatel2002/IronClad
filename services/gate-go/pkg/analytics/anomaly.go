package analytics

import (
	"math"
	"sync"
	"time"
)

// DecisionStats tracks metrics over a sliding window.
type DecisionStats struct {
	mu           sync.RWMutex
	Count        float64
	Mean         float64
	M2           float64 // Running sum of squares of differences from the mean
	WindowSize   time.Duration
	LastDecision time.Time
}

// NewDecisionStats initializes a stats tracker.
func NewDecisionStats(window time.Duration) *DecisionStats {
	return &DecisionStats{
		WindowSize: window,
	}
}

// Record observes a new decision (1 for ALLOW, 0 for BLOCK).
func (s *DecisionStats) Record(value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Count++
	delta := value - s.Mean
	s.Mean += delta / s.Count
	delta2 := value - s.Mean
	s.M2 += delta * delta2
	s.LastDecision = time.Now()
}

// Variance calculates the variance of the observed values.
func (s *DecisionStats) Variance() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Count < 2 {
		return 0
	}
	return s.M2 / s.Count
}

// IsAnomalous returns true if the value is more than 3 standard deviations from the mean.
func (s *DecisionStats) IsAnomalous(value float64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Count < 20 { // Need enough samples to establish a baseline
		return false
	}
	stdDev := math.Sqrt(s.Variance())
	if stdDev == 0 {
		return value != s.Mean
	}
	zScore := math.Abs(value-s.Mean) / stdDev
	return zScore > 3.0
}
