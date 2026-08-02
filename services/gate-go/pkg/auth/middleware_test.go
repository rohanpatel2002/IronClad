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

func TestRateLimiterMiddleware(t *testing.T) {
	limiter := NewIPRateLimiter(2)
	handler := RateLimiterMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "192.168.1.100:12345"

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("Expected request 1 to succeed, got %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req1)
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected request 2 to succeed, got %d", rec2.Code)
	}

	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req1)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected request 3 to be rate limited (429), got %d", rec3.Code)
	}
}

