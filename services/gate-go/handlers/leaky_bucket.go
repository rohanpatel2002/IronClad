package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LeakyBucket implements a simple leaky bucket rate limiter.
type LeakyBucket struct {
	rate       time.Duration // how often a "drop" leaks
	capacity   int           // max capacity of the bucket
	buckets    map[string]*bucket
	mu         sync.Mutex
}

type bucket struct {
	count      int
	lastUpdate time.Time
}

// NewLeakyBucket creates a new leaky bucket limiter.
func NewLeakyBucket(rate time.Duration, capacity int) *LeakyBucket {
	return &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*bucket),
	}
}

// Middleware returns the Gin middleware.
func (lb *LeakyBucket) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lb.mu.Lock()
		defer lb.mu.Unlock()

		b, ok := lb.buckets[ip]
		if !ok {
			b = &bucket{count: 0, lastUpdate: time.Now()}
			lb.buckets[ip] = b
		}

		// Leak drops based on time passed
		now := time.Now()
		elapsed := now.Sub(b.lastUpdate)
		leaked := int(elapsed / lb.rate)
		if leaked > 0 {
			b.count -= leaked
			if b.count < 0 {
				b.count = 0
			}
			b.lastUpdate = now
		}

		// Check if bucket can accept a new request
		if b.count < lb.capacity {
			b.count++
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Leaky bucket overflow, please slow down",
			})
		}
	}
}
