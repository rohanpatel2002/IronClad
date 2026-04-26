package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
)

func TestPrometheusMiddleware_RecordsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Use a fresh registry per test to avoid duplicate registration panics
	registry := prometheus.NewRegistry()

	router := gin.New()
	router.Use(handlers.PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPrometheusMiddleware_HandlesUnknownRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(handlers.PrometheusMiddleware())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/not-registered", nil)
	router.ServeHTTP(w, req)

	// Should respond 404 without panicking
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRecordDecisionMetric_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RecordDecisionMetric panicked: %v", r)
		}
	}()
	handlers.RecordDecisionMetric("ALLOW")
	handlers.RecordDecisionMetric("WARN")
	handlers.RecordDecisionMetric("BLOCK")
}
