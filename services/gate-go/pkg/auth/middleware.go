package auth

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

type contextKey string

const (
	UserContextKey contextKey = "user_claims"
)

// SecurityHeadersMiddleware adds standard security hardening headers to HTTP responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
		next.ServeHTTP(w, r)
	})
}

// BearerAuthMiddleware validates Bearer JWT tokens in the Authorization header.
func BearerAuthMiddleware(jwtMgr *JWTManager, blacklist *TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtMgr.Verify(tokenStr)
			if err != nil {
				http.Error(w, "unauthorized token", http.StatusUnauthorized)
				return
			}

			if blacklist != nil && claims.ID != "" {
				revoked, err := blacklist.IsRevoked(r.Context(), claims.ID)
				if err == nil && revoked {
					http.Error(w, "token has been revoked", http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// IPRateLimiter manages per-IP rate limits using a simple token bucket.
type IPRateLimiter struct {
	mu       sync.Mutex
	requests map[string]int
	maxReqs  int
}

// NewIPRateLimiter creates a new rate limiter with a request cap.
func NewIPRateLimiter(maxRequests int) *IPRateLimiter {
	return &IPRateLimiter{
		requests: make(map[string]int),
		maxReqs:  maxRequests,
	}
}

// Allow checks if the client IP is allowed to proceed.
func (limiter *IPRateLimiter) Allow(ip string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	count := limiter.requests[ip]
	if count >= limiter.maxReqs {
		return false
	}
	limiter.requests[ip] = count + 1
	return true
}

// RateLimiterMiddleware returns an HTTP middleware that throttles IPs exceeding limits.
func RateLimiterMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}

			if !limiter.Allow(ip) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

