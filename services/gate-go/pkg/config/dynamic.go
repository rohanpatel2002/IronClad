package config

import (
	"sync"
	"time"
)

// DynamicConfig holds configuration that can be updated at runtime.
type DynamicConfig struct {
	mu            sync.RWMutex
	RiskThreshold      float64
	Maintenance        bool
	ServiceCriticality map[string]float64
}

func (c *DynamicConfig) GetServiceCriticality(service string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if crit, ok := c.ServiceCriticality[service]; ok {
		return crit
	}
	return 0.5 // Default criticality
}

var (
	instance *DynamicConfig
	once     sync.Once
)

// Get returns the singleton instance of DynamicConfig.
func Get() *DynamicConfig {
	once.Do(func() {
		instance = &DynamicConfig{
			RiskThreshold: 0.8,
			Maintenance:   false,
			ServiceCriticality: map[string]float64{
				"gate-go":        0.9,
				"topology-go":    0.8,
				"scoring-python": 0.7,
				"semantic-python": 0.6,
			},
		}
		go instance.watch()
	})
	return instance
}

func (c *DynamicConfig) watch() {
	// In a real implementation, this would watch a file, Etcd, or ConfigMap.
	// For now, we simulate a periodic check.
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		// Simulate update
		c.mu.Lock()
		// Logic to reload from source would go here
		c.mu.Unlock()
	}
}

func (c *DynamicConfig) GetRiskThreshold() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RiskThreshold
}

func (c *DynamicConfig) IsMaintenanceMode() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Maintenance
}

func (c *DynamicConfig) SetRiskThreshold(val float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RiskThreshold = val
}


