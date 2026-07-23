package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	manager := NewJWTManager()

	token, err := manager.Generate("testuser", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected role admin, got %s", claims.Role)
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
