package handlers

import (
	"log/slog"
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
	stop       chan struct{}
}

type bucket struct {
	count      int
	lastUpdate time.Time
}

// NewLeakyBucket creates a new leaky bucket limiter and starts a background cleanup routine.
func NewLeakyBucket(rate time.Duration, capacity int) *LeakyBucket {
	lb := &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*bucket),
		stop:     make(chan struct{}),
	}
	go lb.cleanupLoop()
	return lb
}

// Stop shuts down the background cleanup routine.
func (lb *LeakyBucket) Stop() {
	close(lb.stop)
}

func (lb *LeakyBucket) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lb.cleanup()
		case <-lb.stop:
			return
		}
	}
}

func (lb *LeakyBucket) cleanup() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	removed := 0
	for ip, b := range lb.buckets {
		// If the bucket is empty and hasn't been updated in 1 hour, remove it
		if b.count == 0 && now.Sub(b.lastUpdate) > 1*time.Hour {
			delete(lb.buckets, ip)
			removed++
		}
	}
	if removed > 0 {
		slog.Info("Cleaned up stale rate limit buckets", "count", removed)
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

// GetBucketState returns the current fill level and capacity for an IP.
func (lb *LeakyBucket) GetBucketState(ip string) (count int, capacity int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if b, ok := lb.buckets[ip]; ok {
		return b.count, lb.capacity
	}
	return 0, lb.capacity
}

