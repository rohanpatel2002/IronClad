package slo

import (
	"sync"
	"time"
)

// SLOMonitor tracks error rates against a defined objective.
type SLOMonitor struct {
	mu           sync.RWMutex
	TotalReqs    int64
	FailedReqs   int64
	Objective    float64 // e.g., 0.999 for 99.9% success rate
	WindowStart  time.Time
}

// NewSLOMonitor creates a monitor with a target objective.
func NewSLOMonitor(objective float64) *SLOMonitor {
	return &SLOMonitor{
		Objective:   objective,
		WindowStart: time.Now(),
	}
}

// Record observes a request outcome.
func (m *SLOMonitor) Record(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalReqs++
	if !success {
		m.FailedReqs++
	}
}

// Availability returns the current success rate.
func (m *SLOMonitor) Availability() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.TotalReqs == 0 {
		return 1.0
	}
	return float64(m.TotalReqs-m.FailedReqs) / float64(m.TotalReqs)
}

// IsViolating returns true if the current availability is below the objective.
func (m *SLOMonitor) IsViolating() bool {
	return m.Availability() < m.Objective
}

// Reset clears the counters for a new window.
func (m *SLOMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalReqs = 0
	m.FailedReqs = 0
	m.WindowStart = time.Now()
}

