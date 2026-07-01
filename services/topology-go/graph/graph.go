// Package graph provides an in-memory directed service dependency graph
// with BFS blast radius traversal for the IRONCLAD topology service.
package graph

import "sync"

// ServiceNode represents a node in the service dependency graph.
type ServiceNode struct {
	Name         string   `json:"name"`
	Criticality  float64  `json:"criticality"` // 0-1
	DependsOn    []string `json:"depends_on"`  // outgoing edges (downstream)
	DependedOnBy []string `json:"depended_on_by"` // reverse edges (upstream)
}

// BlastRadiusResult is the output of a blast radius query.
type BlastRadiusResult struct {
	Service          string   `json:"service"`
	BlastRadiusScore float64  `json:"blast_radius_score"`
	ImpactedServices []string `json:"impacted_services"`
	TotalNodes       int      `json:"total_nodes"`
}

// DependencyGraph is a thread-safe directed graph of service dependencies.
type DependencyGraph struct {
	mu    sync.RWMutex
	nodes map[string]*ServiceNode
}

// New creates a new empty dependency graph.
func New() *DependencyGraph {
	return &DependencyGraph{nodes: make(map[string]*ServiceNode)}
}

// NewDefault creates a dependency graph pre-loaded with a realistic
// microservice topology for IRONCLAD development.
func NewDefault() *DependencyGraph {
	g := New()
	g.LoadDefault()
	return g
}

// LoadDefault populates the graph with a representative service topology.
func (g *DependencyGraph) LoadDefault() {
	services := []struct {
		name        string
		criticality float64
		deps        []string
	}{
		{"api-gateway", 0.95, []string{"auth-service", "order-service", "payment-api", "user-service"}},
		{"payment-api", 1.00, []string{"auth-service", "fraud-detection", "notification-service", "audit-logger"}},
		{"order-service", 0.90, []string{"payment-api", "inventory-service", "notification-service"}},
		{"auth-service", 0.95, []string{"user-service", "session-store", "audit-logger"}},
		{"user-service", 0.75, []string{"database-primary", "cache-redis"}},
		{"fraud-detection", 0.80, []string{"ml-inference", "audit-logger"}},
		{"inventory-service", 0.70, []string{"database-primary", "cache-redis"}},
		{"notification-service", 0.50, []string{"email-gateway", "sms-gateway"}},
		{"session-store", 0.60, []string{"cache-redis"}},
		{"audit-logger", 0.40, []string{"database-primary"}},
		{"ml-inference", 0.65, []string{"model-store"}},
		{"database-primary", 0.98, []string{}},
		{"cache-redis", 0.85, []string{}},
		{"email-gateway", 0.30, []string{}},
		{"sms-gateway", 0.30, []string{}},
		{"model-store", 0.55, []string{}},
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Add all nodes first
	for _, s := range services {
		g.nodes[s.name] = &ServiceNode{
			Name:        s.name,
			Criticality: s.criticality,
			DependsOn:   s.deps,
		}
	}

	// Build reverse edges
	for _, s := range services {
		for _, dep := range s.deps {
			if node, ok := g.nodes[dep]; ok {
				node.DependedOnBy = append(node.DependedOnBy, s.name)
			}
		}
	}
}

// AddService adds or replaces a service node in the graph.
func (g *DependencyGraph) AddService(node ServiceNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[node.Name] = &node

	// Update reverse edges for new node's dependencies
	for _, dep := range node.DependsOn {
		if target, ok := g.nodes[dep]; ok {
			target.DependedOnBy = append(target.DependedOnBy, node.Name)
		}
	}
}

// GetService returns a copy of a service node by name.
func (g *DependencyGraph) GetService(name string) (*ServiceNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[name]
	if !ok {
		return nil, false
	}
	copy := *n
	return &copy, true
}

// ListServices returns all service names in the graph.
func (g *DependencyGraph) ListServices() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	names := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		names = append(names, k)
	}
	return names
}

// ComputeBlastRadius performs BFS from `service` in both directions:
// upstream (who depends on this service) and downstream (what this service depends on).
// Returns a 0-1 blast radius score and the list of impacted services.
func (g *DependencyGraph) ComputeBlastRadius(service string) BlastRadiusResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	impacted := g.bfsImpact(service)
	score := g.computeScore(service, impacted)

	return BlastRadiusResult{
		Service:          service,
		BlastRadiusScore: score,
		ImpactedServices: impacted,
		TotalNodes:       len(g.nodes),
	}
}

// bfsImpact returns all services that could be affected by a change to `service`.
// Traverses upstream (reverse edges: services that call this service) and
// also includes direct downstream dependencies.
func (g *DependencyGraph) bfsImpact(start string) []string {
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true
	var result []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node, ok := g.nodes[current]
		if !ok {
			continue
		}

		// Upstream: services that call into `current` (reverse edges)
		for _, upstream := range node.DependedOnBy {
			if !visited[upstream] {
				visited[upstream] = true
				result = append(result, upstream)
				queue = append(queue, upstream)
			}
		}

		// Also include direct downstream for the starting node
		if current == start {
			for _, downstream := range node.DependsOn {
				if !visited[downstream] {
					visited[downstream] = true
					result = append(result, downstream)
					// Don't recurse downstream — only direct deps
				}
			}
		}
	}

	return result
}

// computeScore returns a 0-1 blast radius score based on weighted criticality.
func (g *DependencyGraph) computeScore(service string, impacted []string) float64 {
	if len(g.nodes) == 0 {
		return 0
	}

	totalCriticality := 0.0
	if node, ok := g.nodes[service]; ok {
		totalCriticality += node.Criticality
	}
	for _, svc := range impacted {
		if node, ok := g.nodes[svc]; ok {
			totalCriticality += node.Criticality
		}
	}

	score := totalCriticality / float64(len(g.nodes))
	if score > 1.0 {
		return 1.0
	}
	return score
}
