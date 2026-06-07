package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// APIKey represents a machine-to-machine authentication credential.
type APIKey struct {
	Key       string
	TenantID  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// APIKeyManager handles issuance and validation of API keys.
type APIKeyManager struct {
	mu   sync.RWMutex
	keys map[string]APIKey
}

// NewAPIKeyManager creates a new manager.
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]APIKey),
	}
}

// GenerateKey creates a new key for a tenant.
func (m *APIKeyManager) GenerateKey(tenantID string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	key := hex.EncodeToString(b)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[key] = APIKey{
		Key:       key,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	return key, nil
}

// ValidateKey checks if a key is valid and not expired.
func (m *APIKeyManager) ValidateKey(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	k, ok := m.keys[key]
	if !ok {
		return "", errors.New("invalid api key")
	}

	if time.Now().After(k.ExpiresAt) {
		return "", errors.New("api key expired")
	}

	return k.TenantID, nil
}
