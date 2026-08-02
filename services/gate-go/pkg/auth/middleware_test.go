package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeadersMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options to be nosniff")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options to be DENY")
	}
}

func TestBearerAuthMiddleware(t *testing.T) {
	jwtMgr := NewJWTManager()
	bl := NewTokenBlacklist(nil)

	protectedHandler := BearerAuthMiddleware(jwtMgr, bl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Unauthorized request (no header)
	req1 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec1 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for request without header, got %d", rec1.Code)
	}

	// Authorized request
	tokenStr, err := jwtMgr.Generate("charlie", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	rec2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid bearer token, got %d", rec2.Code)
	}
}
