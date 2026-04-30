package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// rateLimiter struct holds a map of IP addresses to their specific rate limiters.
type rateLimiter struct {
	ips   map[string]*rate.Limiter
	mu    *sync.RWMutex
	r     rate.Limit
	b     int
	redis *redis.Client
}

// NewRateLimiter creates a new IP-based rate limiter middleware.
func NewRateLimiter(r rate.Limit, b int, redisClient *redis.Client) *rateLimiter {
	return &rateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		r:     r,
		b:     b,
		redis: redisClient,
	}
}

// getLimiter returns the rate limiter for a given IP, creating it if it doesn't exist (In-Memory).
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

// isAllowed checks if the request is allowed using either Redis or In-Memory.
func (l *rateLimiter) isAllowed(ctx context.Context, ip string) bool {
	if l.redis == nil {
		return l.getLimiter(ip).Allow()
	}

	// Redis-based sliding window rate limiting
	key := "rate_limit:" + ip
	now := time.Now().UnixNano()
	window := time.Second
	
	pipe := l.redis.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", string(now-int64(window)))
	pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: string(now)})
	pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window)
	
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return true // Fail open if Redis is down
	}

	count := cmds[2].(*redis.IntCmd).Val()
	return count <= int64(l.b)
}

// Middleware returns the Gin middleware handler.
func (l *rateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !l.isAllowed(c.Request.Context(), ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}
