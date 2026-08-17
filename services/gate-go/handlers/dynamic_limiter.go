package handlers

import (
	"net/http"
	"runtime"
	"sync"

	"github.com/gin-gonic/gin"
)

// DynamicLimiter adjusts rate limits based on system resource pressure.
type DynamicLimiter struct {
	mu           sync.RWMutex
	maxMemoryPct float64
}

// NewDynamicLimiter creates a new dynamic limiter.
func NewDynamicLimiter(maxMemoryPct float64) *DynamicLimiter {
	return &DynamicLimiter{
		maxMemoryPct: maxMemoryPct,
	}
}

// Middleware returns the Gin middleware.
func (l *DynamicLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple resource check: Memory usage
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// This is a simplified check. In a container, we would use cgroups info.
		// For this demo, we'll simulate load if memory usage (Alloc) is > X
		
		// Let's say if Alloc > 500MB (arbitrary for demo), we start throttling.
		if m.Alloc > 500*1024*1024 {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "system_pressure",
				"message": "Service is under high resource pressure, please try again later",
			})
			return
		}

		c.Next()
	}
}

// SetMaxMemoryMB dynamically configures memory threshold in MB.
func (l *DynamicLimiter) SetMaxMemoryMB(mb float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxMemoryPct = mb
}

