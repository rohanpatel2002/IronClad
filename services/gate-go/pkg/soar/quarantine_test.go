package soar

import (
	"context"
	"testing"
)

func TestQuarantineManager_IsQuarantined(t *testing.T) {
	qm := NewQuarantineManager("mock://localhost:8181")
	quarantined, err := qm.IsQuarantined(context.Background(), "payment-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quarantined {
		t.Errorf("expected false for mock quarantine check")
	}
}
