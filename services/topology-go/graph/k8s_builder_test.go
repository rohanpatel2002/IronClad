package graph

import (
	"testing"
	"time"
)

func TestK8sGraphBuilder_CacheTTL(t *testing.T) {
	builder := NewK8sGraphBuilder(nil, 1*time.Hour)
	if !builder.LastUpdated().IsZero() {
		t.Errorf("expected zero lastUpdated initially")
	}
}
