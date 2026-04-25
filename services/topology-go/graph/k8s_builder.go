package graph

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rohanpatel2002/ironclad/services/topology-go/clients"
)

// K8sGraphBuilder handles dynamic fetching and caching of the graph from Kubernetes
type K8sGraphBuilder struct {
	k8sClient *clients.K8sClient
	graph     *DependencyGraph
	ttl       time.Duration

	mu         sync.RWMutex
	lastUpdate time.Time
}

// NewK8sGraphBuilder creates a new builder with TTL caching.
func NewK8sGraphBuilder(k8sClient *clients.K8sClient, ttl time.Duration) *K8sGraphBuilder {
	return &K8sGraphBuilder{
		k8sClient: k8sClient,
		graph:     New(),
		ttl:       ttl,
	}
}

// GetGraph returns the dependency graph. If the cache is expired, it refreshes from K8s.
func (b *K8sGraphBuilder) GetGraph(ctx context.Context) (*DependencyGraph, error) {
	b.mu.RLock()
	isExpired := time.Since(b.lastUpdate) > b.ttl
	b.mu.RUnlock()

	if isExpired {
		err := b.refresh(ctx)
		if err != nil {
			// If refresh fails, return the stale graph and log an error
			log.Printf("[K8sGraphBuilder] Failed to refresh topology: %v. Using stale cache.", err)
			
			// If we have no graph at all, we must return the error
			b.mu.RLock()
			empty := len(b.graph.ListServices()) == 0
			b.mu.RUnlock()
			
			if empty {
				return nil, fmt.Errorf("topology refresh failed and no cached graph available: %w", err)
			}
		}
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.graph, nil
}

// refresh pulls the latest ServiceMetadata from K8s and replaces the current graph
func (b *K8sGraphBuilder) refresh(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Double-check expiration after acquiring write lock
	if time.Since(b.lastUpdate) <= b.ttl {
		return nil
	}

	metadataList, err := b.k8sClient.GetServiceTopology(ctx)
	if err != nil {
		return err
	}

	newGraph := New()
	for _, meta := range metadataList {
		newGraph.AddService(ServiceNode{
			Name:        meta.Name,
			Criticality: meta.Criticality,
			DependsOn:   meta.DependsOn,
		})
	}

	b.graph = newGraph
	b.lastUpdate = time.Now()
	
	log.Printf("[K8sGraphBuilder] Topology refreshed successfully from K8s. %d services loaded.", len(metadataList))
	return nil
}

// StartBackgroundRefresher kicks off a goroutine to keep the graph warm.
func (b *K8sGraphBuilder) StartBackgroundRefresher(ctx context.Context) {
	ticker := time.NewTicker(b.ttl / 2) // Refresh before it expires
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				_ = b.refresh(ctx)
			}
		}
	}()
}
