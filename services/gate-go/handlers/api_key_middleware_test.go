package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/auth"
)

func TestAPIKeyMiddleware_MissingKeyAllowsFallthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := auth.NewAPIKeyManager() // Need a 32 byte secret usually, but this is a stub
	router := gin.New()
	router.Use(handlers.APIKeyMiddleware(manager))
	
	called := false
	router.GET("/test", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := auth.NewAPIKeyManager()
	router := gin.New()
	router.Use(handlers.APIKeyMiddleware(manager))
	
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
