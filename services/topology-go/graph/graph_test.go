package graph

import (
	"testing"
)

func TestDependencyGraph_DetectCycles(t *testing.T) {
	g := New()
	g.AddService(ServiceNode{Name: "svc-a", DependsOn: []string{"svc-b"}})
	g.AddService(ServiceNode{Name: "svc-b", DependsOn: []string{"svc-c"}})
	g.AddService(ServiceNode{Name: "svc-c", DependsOn: []string{"svc-a"}})

	cycles := g.DetectCycles()
	if len(cycles) == 0 {
		t.Fatalf("expected cycle, got none")
	}

	foundCycle := cycles[0]
	if len(foundCycle) < 3 {
		t.Errorf("expected cycle path length >= 3, got %d", len(foundCycle))
	}
}

func TestDependencyGraph_MaxDepth(t *testing.T) {
	g := NewDefault()
	depth := g.MaxDepth("api-gateway")
	if depth <= 0 {
		t.Errorf("expected api-gateway depth > 0, got %d", depth)
	}

	dbDepth := g.MaxDepth("database-primary")
	if dbDepth != 0 {
		t.Errorf("expected leaf node depth 0, got %d", dbDepth)
	}
}
