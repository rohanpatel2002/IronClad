package slo

import "testing"

func TestSLOMonitor(t *testing.T) {
	m := NewSLOMonitor(0.99)
	if m.Availability() != 1.0 {
		t.Errorf("expected initial availability 1.0")
	}

	for i := 0; i < 99; i++ {
		m.Record(true)
	}
	m.Record(false)

	if m.IsViolating() {
		t.Errorf("expected 99%% availability to meet 0.99 objective")
	}

	budget := m.ErrorBudgetRemaining()
	if budget < 0 || budget > 1 {
		t.Errorf("unexpected error budget: %f", budget)
	}

	m.Record(false)
	if !m.IsViolating() {
		t.Errorf("expected availability < 0.99 to be violating")
	}
}
