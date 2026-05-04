package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// TokenBlacklist handles revocation of compromised or logged-out JWTs.
type TokenBlacklist struct {
	client *redis.Client
	prefix string
}

// NewTokenBlacklist creates a new blacklist manager.
func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{
		client: client,
		prefix: "blacklist:jti:",
	}
}

// Revoke adds a token ID (JTI) to the blacklist until it would have expired.
func (b *TokenBlacklist) Revoke(ctx context.Context, jti string, expiration time.Duration) error {
	return b.client.Set(ctx, b.prefix+jti, "1", expiration).Err()
}

// IsRevoked checks if a token ID is in the blacklist.
func (b *TokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	val, err := b.client.Get(ctx, b.prefix+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis error: %w", err)
	}
	return val == "1", nil
}
