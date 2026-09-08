package events

import (
	"context"
	"testing"
)

func TestInMemoryPublisher(t *testing.T) {
	pub := NewInMemoryPublisher()
	err := pub.PublishScoringEvent(context.Background(), map[string]string{"service": "payment-api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.Events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(pub.Events))
	}
}
