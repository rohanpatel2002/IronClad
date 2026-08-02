package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/audit"
)

func TestSecurity_FullIntegrationFlow(t *testing.T) {
	jwtMgr := NewJWTManager()
	blacklist := NewTokenBlacklist(nil)
	auditLogger := audit.NewAuditLoggerWithSecret(nil, "integration-secret")

	// 1. Generate token
	tokenStr, err := jwtMgr.Generate("dave", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtMgr.Verify(tokenStr)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	// 2. Setup protected handler
	handler := SecurityHeadersMiddleware(
		BearerAuthMiddleware(jwtMgr, blacklist)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	// 3. Test initial request succeeds
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/secure", nil)
	req1.Header.Set("Authorization", "Bearer "+tokenStr)
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rec1.Code)
	}
	if rec1.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected Security Headers to be set")
	}

	// 4. Revoke token and log audit record
	ctx := context.Background()
	err = blacklist.Revoke(ctx, claims.ID, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	rec := audit.LogRecord{
		TenantID:  "tenant-1",
		Actor:     claims.Username,
		Action:    "TOKEN_REVOKE",
		Status:    "SUCCESS",
		IPAddress: "127.0.0.1",
	}
	rec.Signature = auditLogger.ComputeSignature(rec)
	auditLogger.Log(ctx, rec)

	if !auditLogger.VerifySignature(rec) {
		t.Errorf("Audit log signature verification failed")
	}

	// 5. Subsequent request with revoked token fails
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/secure", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for revoked token, got %d", rec2.Code)
	}
}
