package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// APIKey represents a machine-to-machine authentication credential metadata.
type APIKey struct {
	HashedKey string
	TenantID  string
	Scopes    []string
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

func hashKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// GenerateKey creates a new key for a tenant.
func (m *APIKeyManager) GenerateKey(tenantID string, ttl time.Duration) (string, error) {
	return m.GenerateKeyWithScopes(tenantID, ttl, nil)
}

// GenerateKeyWithScopes creates a new key for a tenant with specified scopes.
func (m *APIKeyManager) GenerateKeyWithScopes(tenantID string, ttl time.Duration, scopes []string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	rawKey := hex.EncodeToString(b)
	hashed := hashKey(rawKey)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[hashed] = APIKey{
		HashedKey: hashed,
		TenantID:  tenantID,
		Scopes:    scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	return rawKey, nil
}

// ValidateKey checks if a key is valid and not expired.
func (m *APIKeyManager) ValidateKey(rawKey string) (string, error) {
	entry, err := m.GetValidKey(rawKey)
	if err != nil {
		return "", err
	}
	return entry.TenantID, nil
}

// GetValidKey verifies rawKey and returns the APIKey entry.
func (m *APIKeyManager) GetValidKey(rawKey string) (*APIKey, error) {
	hashed := hashKey(rawKey)

	m.mu.RLock()
	defer m.mu.RUnlock()

	k, ok := m.keys[hashed]
	if !ok {
		return nil, errors.New("invalid api key")
	}

	if time.Now().After(k.ExpiresAt) {
		return nil, errors.New("api key expired")
	}

	return &k, nil
}

// ValidateScope checks if a raw API key has a specific scope.
func (m *APIKeyManager) ValidateScope(rawKey, requiredScope string) bool {
	entry, err := m.GetValidKey(rawKey)
	if err != nil {
		return false
	}
	for _, s := range entry.Scopes {
		if s == requiredScope || s == "*" {
			return true
		}
	}
	return false
}

