package cost

import (
	"time"
)

// Optimizer provides recommendations for cost-effective deployment windows.
type Optimizer struct {
	// Simple mock of peak/off-peak pricing hours (UTC)
	PeakStart int
	PeakEnd   int
}

// NewOptimizer creates a new cost optimizer.
func NewOptimizer() *Optimizer {
	return &Optimizer{
		PeakStart: 14, // 2 PM UTC
		PeakEnd:   20, // 8 PM UTC
	}
}

// ShouldDefer returns true if the current time is a high-cost peak period.
func (o *Optimizer) ShouldDefer(now time.Time) bool {
	hour := now.Hour()
	return hour >= o.PeakStart && hour <= o.PeakEnd
}

// GetNextWindow returns the next off-peak time for deployment.
func (o *Optimizer) GetNextWindow(now time.Time) time.Time {
	if !o.ShouldDefer(now) {
		return now
	}
	// Move to peak end + 1 hour
	return time.Date(now.Year(), now.Month(), now.Day(), o.PeakEnd+1, 0, 0, 0, time.UTC)
}
