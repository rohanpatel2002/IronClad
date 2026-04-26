package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimiter struct holds a map of IP addresses to their specific rate limiters.
type rateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewRateLimiter creates a new IP-based rate limiter middleware.
func NewRateLimiter(r rate.Limit, b int) *rateLimiter {
	return &rateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// getLimiter returns the rate limiter for a given IP, creating it if it doesn't exist.
func (l *rateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.ips[ip]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		limiter, exists = l.ips[ip]
		if !exists {
			limiter = rate.NewLimiter(l.r, l.b)
			l.ips[ip] = limiter
		}
		l.mu.Unlock()
	}

	return limiter
}

// Middleware returns the Gin middleware handler.
func (l *rateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := l.getLimiter(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}
