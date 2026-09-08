package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCircuitBreakerStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/circuit-breaker/status", CircuitBreakerStatusHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/circuit-breaker/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
