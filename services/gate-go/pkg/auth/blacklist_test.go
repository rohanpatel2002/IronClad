package auth

import (
	"context"
	"testing"
	"time"
)

func TestTokenBlacklist_InMemoryFallback(t *testing.T) {
	bl := NewTokenBlacklist(nil) // nil redis client
	ctx := context.Background()

	revoked, err := bl.IsRevoked(ctx, "jti-123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if revoked {
		t.Errorf("Expected jti-123 to not be revoked initially")
	}

	err = bl.Revoke(ctx, "jti-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	revoked, err = bl.IsRevoked(ctx, "jti-123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !revoked {
		t.Errorf("Expected jti-123 to be revoked after calling Revoke")
	}
}
