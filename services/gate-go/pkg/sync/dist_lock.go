package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// DistLock provides a distributed mutex using Redis.
type DistLock struct {
	client *redis.Client
	key    string
	value  string
}

// NewDistLock creates a lock instance for a specific key.
func NewDistLock(client *redis.Client, key string) *DistLock {
	return &DistLock{
		client: client,
		key:    "lock:" + key,
		value:  uuid.New().String(),
	}
}

// Lock attempts to acquire the lock with a TTL.
func (l *DistLock) Lock(ctx context.Context, ttl time.Duration) (bool, error) {
	return l.client.SetNX(ctx, l.key, l.value, ttl).Result()
}

// Unlock releases the lock if it's held by this instance.
func (l *DistLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}
	return nil
}

// GetKey returns the raw lock key name in Redis.
func (l *DistLock) GetKey() string {
	return l.key
}

