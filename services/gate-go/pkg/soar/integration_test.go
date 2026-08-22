package soar

import (
	"context"
	"testing"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/events"
)

func TestSOAR_EventIntegration(t *testing.T) {
	qm := NewQuarantineManager("mock://localhost")
	pub := events.NewInMemoryPublisher()

	ctx := context.Background()
	_ = qm.QuarantineService(ctx, "payment-api", "suspicious SQL payload")
	_ = pub.PublishScoringEvent(ctx, map[string]string{
		"action":  "QUARANTINE",
		"service": "payment-api",
	})

	if len(pub.Events) != 1 {
		t.Errorf("expected 1 quarantine event in publisher queue, got %d", len(pub.Events))
	}
}
