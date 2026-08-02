package auth

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// TokenBlacklist handles revocation of compromised or logged-out JWTs.
type TokenBlacklist struct {
	client     *redis.Client
	prefix     string
	mu         sync.RWMutex
	memoryCache map[string]time.Time
}

// NewTokenBlacklist creates a new blacklist manager.
func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{
		client:      client,
		prefix:      "blacklist:jti:",
		memoryCache: make(map[string]time.Time),
	}
}

// Revoke adds a token ID (JTI) to the blacklist until it would have expired.
func (b *TokenBlacklist) Revoke(ctx context.Context, jti string, expiration time.Duration) error {
	if b.client != nil {
		err := b.client.Set(ctx, b.prefix+jti, "1", expiration).Err()
		if err == nil {
			return nil
		}
	}

	// Fallback to in-memory store
	b.mu.Lock()
	defer b.mu.Unlock()
	b.memoryCache[jti] = time.Now().Add(expiration)
	return nil
}

// IsRevoked checks if a token ID is in the blacklist.
func (b *TokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if b.client != nil {
		val, err := b.client.Get(ctx, b.prefix+jti).Result()
		if err == nil {
			return val == "1", nil
		}
		if err != redis.Nil {
			// Fallback check if Redis error
		} else {
			return false, nil
		}
	}

	// In-memory fallback check
	b.mu.RLock()
	expiry, exists := b.memoryCache[jti]
	b.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(expiry) {
		// Clean up expired entry
		b.mu.Lock()
		delete(b.memoryCache, jti)
		b.mu.Unlock()
		return false, nil
	}

	return true, nil
}

