package config

import (
	"sync"
	"time"
)

// DynamicConfig holds configuration that can be updated at runtime.
type DynamicConfig struct {
	mu            sync.RWMutex
	RiskThreshold float64
	Maintenance   bool
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
