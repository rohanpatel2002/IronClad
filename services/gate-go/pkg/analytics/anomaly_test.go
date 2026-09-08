package analytics

import (
	"testing"
	"time"
)

func TestDecisionStats(t *testing.T) {
	stats := NewDecisionStats(1 * time.Hour)
	for i := 0; i < 25; i++ {
		stats.Record(0.5)
	}

	if stats.IsAnomalous(0.5) {
		t.Errorf("expected 0.5 not to be anomalous")
	}

	z := stats.ZScore(0.5)
	if z != 0 {
		t.Errorf("expected Z-score 0 for identical sample, got %f", z)
	}
}
