package auth

import (
	"testing"
	"time"
)

func TestAPIKeyManager_HashingAndScopes(t *testing.T) {
	mgr := NewAPIKeyManager()

	key, err := mgr.GenerateKeyWithScopes("tenant-1", 1*time.Hour, []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	tenantID, err := mgr.ValidateKey(key)
	if err != nil {
		t.Fatalf("Failed to validate key: %v", err)
	}
	if tenantID != "tenant-1" {
		t.Errorf("Expected tenant-1, got %s", tenantID)
	}

	if !mgr.ValidateScope(key, "read") {
		t.Errorf("Expected key to have 'read' scope")
	}
	if !mgr.ValidateScope(key, "write") {
		t.Errorf("Expected key to have 'write' scope")
	}
	if mgr.ValidateScope(key, "admin") {
		t.Errorf("Did not expect key to have 'admin' scope")
	}

	// Invalid key test
	if _, err := mgr.ValidateKey("invalid-raw-key"); err == nil {
		t.Errorf("Expected error for invalid key")
	}

	// Expired key test
	expKey, err := mgr.GenerateKey("tenant-2", -1*time.Second)
	if err != nil {
		t.Fatalf("Failed to generate expired key: %v", err)
	}
	if _, err := mgr.ValidateKey(expKey); err == nil {
		t.Errorf("Expected error for expired key")
	}
}
