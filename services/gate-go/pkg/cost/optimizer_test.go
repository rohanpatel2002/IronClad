package cost

import (
	"testing"
	"time"
)

func TestOptimizer_ShouldDefer(t *testing.T) {
	opt := NewOptimizer()

	peakTime := time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC)
	if !opt.ShouldDefer(peakTime) {
		t.Errorf("Expected peak time 15:00 UTC to defer execution")
	}

	offPeakTime := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	if opt.ShouldDefer(offPeakTime) {
		t.Errorf("Expected off-peak time 10:00 UTC not to defer execution")
	}
}

func TestOptimizer_GetNextWindow(t *testing.T) {
	opt := NewOptimizer()

	peakTime := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	next := opt.GetNextWindow(peakTime)

	if next.Hour() != 21 {
		t.Errorf("Expected next window to start at 21:00 UTC, got %d", next.Hour())
	}
}

func TestOptimizer_EstimatedSavings(t *testing.T) {
	opt := NewOptimizer()
	peakTime := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	if opt.EstimatedSavings(peakTime) != 0.35 {
		t.Errorf("expected 0.35 estimated savings during peak time")
	}
}

