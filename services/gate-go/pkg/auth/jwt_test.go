package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	mgr := NewJWTManager()
	tokenStr, err := mgr.Generate("alice", "admin")
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := mgr.Verify(tokenStr)
	if err != nil {
		t.Fatalf("Failed to verify JWT: %v", err)
	}

	if claims.Username != "alice" || claims.Role != "admin" {
		t.Errorf("Unexpected claims: username=%s role=%s", claims.Username, claims.Role)
	}
}

func TestJWTManager_InvalidIssuerOrAudience(t *testing.T) {
	mgr1 := NewJWTManager()
	mgr1.SetIssuerAudience("issuer-1", "aud-1")

	tokenStr, err := mgr1.Generate("bob", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	mgr2 := NewJWTManager()
	mgr2.SetIssuerAudience("issuer-2", "aud-1")

	_, err = mgr2.Verify(tokenStr)
	if err == nil {
		t.Errorf("Expected verification error for mismatched issuer")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := NewJWTManager()

	_, err := manager.Verify("invalid.jwt.token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := &JWTManager{
		secretKey:     []byte("test-secret"),
		tokenDuration: -1 * time.Hour,
	}

	token, err := manager.Generate("expireduser", "user")
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	_, err = manager.Verify(token)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

