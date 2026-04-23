package clients

import (
	"context"
	"strings"
)

// serviceGraph is a hardcoded representation of a realistic microservice
// dependency graph. Each key depends on the services in its value slice.
// This is the mock used until the live topology-go service is running.
var serviceGraph = map[string][]string{
	"payment-api":      {"auth-service", "fraud-detection", "notification-service", "audit-logger"},
	"auth-service":     {"user-service", "session-store", "audit-logger"},
	"user-service":     {"database-primary", "cache-redis"},
	"fraud-detection":  {"ml-inference", "audit-logger"},
	"ml-inference":     {"model-store"},
	"order-service":    {"payment-api", "inventory-service", "notification-service"},
	"inventory-service": {"database-primary", "cache-redis"},
	"notification-service": {"email-gateway", "sms-gateway"},
	"session-store":    {"cache-redis"},
	"audit-logger":     {"database-primary"},
	"api-gateway":      {"auth-service", "order-service", "payment-api", "user-service"},
	"database-primary": {},
	"cache-redis":      {},
	"email-gateway":    {},
	"sms-gateway":      {},
	"model-store":      {},
}

// serviceCriticality maps services to a 0-1 criticality score
var serviceCriticality = map[string]float64{
	"payment-api":      1.0,
	"auth-service":     0.95,
	"api-gateway":      0.95,
	"database-primary": 0.98,
	"cache-redis":      0.85,
	"order-service":    0.90,
	"fraud-detection":  0.80,
	"user-service":     0.75,
	"inventory-service": 0.70,
	"notification-service": 0.50,
	"session-store":    0.60,
	"audit-logger":     0.40,
	"ml-inference":     0.65,
	"model-store":      0.55,
	"email-gateway":    0.30,
	"sms-gateway":      0.30,
}

// TopologyClient computes blast radius from the dependency graph.
// When the topology-go service is live, this will delegate over HTTP.
type TopologyClient struct {
	topologyURL string
}

// NewTopologyClient creates a new topology client.
func NewTopologyClient(url string) *TopologyClient {
	return &TopologyClient{topologyURL: url}
}

// GetBlastRadius performs a BFS over the service dependency graph to find all
// downstream services affected by a deployment, then computes a 0-1 blast
// radius score weighted by service criticality.
func (t *TopologyClient) GetBlastRadius(ctx context.Context, service string, changedFiles []string) (float64, []string, error) {
	// Build reverse graph: who depends on `service`?
	reverseGraph := buildReverseGraph(serviceGraph)

	// BFS to find all upstream dependents (services that call into `service`)
	impacted := bfsImpact(service, reverseGraph)

	// Also account for services that `service` itself calls (downstream risk)
	direct := serviceGraph[service]
	impacted = mergeUnique(impacted, direct)

	// File-pattern adjustment: migrations & infra files increase blast radius
	fileMultiplier := computeFileMultiplier(changedFiles)

	// Compute weighted blast radius score
	blastScore := computeBlastScore(service, impacted, fileMultiplier)

	return blastScore, impacted, nil
}

// GetServiceCriticality returns the criticality score for a service (0-1).
func GetServiceCriticality(service string) float64 {
	if c, ok := serviceCriticality[service]; ok {
		return c
	}
	return 0.5 // default for unknown services
}

// buildReverseGraph inverts the dependency graph to find dependents.
func buildReverseGraph(graph map[string][]string) map[string][]string {
	reverse := make(map[string][]string)
	for svc, deps := range graph {
		for _, dep := range deps {
			reverse[dep] = append(reverse[dep], svc)
		}
	}
	return reverse
}

// bfsImpact performs BFS from a starting node in the given graph.
func bfsImpact(start string, graph map[string][]string) []string {
	visited := make(map[string]bool)
	queue := []string{start}
	var result []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] || current == start {
			visited[current] = true
			// Still enqueue its neighbors
			for _, neighbor := range graph[current] {
				if !visited[neighbor] {
					queue = append(queue, neighbor)
				}
			}
			continue
		}

		visited[current] = true
		result = append(result, current)

		for _, neighbor := range graph[current] {
			if !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}
	}

	return result
}

// computeFileMultiplier returns a risk multiplier based on changed file patterns.
func computeFileMultiplier(changedFiles []string) float64 {
	multiplier := 1.0
	for _, f := range changedFiles {
		lower := strings.ToLower(f)
		switch {
		case strings.Contains(lower, "migration") || strings.HasSuffix(lower, ".sql"):
			multiplier = max64(multiplier, 1.5)
		case strings.Contains(lower, "config") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
			multiplier = max64(multiplier, 1.3)
		case strings.Contains(lower, "dockerfile") || strings.Contains(lower, "docker-compose"):
			multiplier = max64(multiplier, 1.4)
		case strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, "_test.py"):
			multiplier = min64(multiplier, 0.8) // test-only changes are lower risk
		}
	}
	return multiplier
}

// computeBlastScore combines impacted services and criticality into a 0-1 score.
func computeBlastScore(service string, impacted []string, multiplier float64) float64 {
	totalCriticality := GetServiceCriticality(service)
	for _, svc := range impacted {
		totalCriticality += GetServiceCriticality(svc)
	}

	totalServices := float64(len(serviceGraph))
	rawScore := totalCriticality / totalServices * multiplier

	// Clamp to [0, 1]
	if rawScore > 1.0 {
		return 1.0
	}
	return rawScore
}

// mergeUnique merges two string slices, deduplicating entries.
func mergeUnique(a, b []string) []string {
	seen := make(map[string]bool, len(a))
	for _, v := range a {
		seen[v] = true
	}
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			a = append(a, v)
		}
	}
	return a
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
