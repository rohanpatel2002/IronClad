package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ironclad_gate_http_requests_total",
			Help: "Total number of HTTP requests processed by gate-go",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ironclad_gate_http_request_duration_seconds",
			Help:    "Histogram of response latency (seconds) of gate-go HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	decisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ironclad_gate_decisions_total",
			Help: "Total number of deployment decisions made",
		},
		[]string{"decision"},
	)
)

// RecordDecisionMetric increments the counter for a specific decision outcome
func RecordDecisionMetric(decision string) {
	decisionsTotal.WithLabelValues(decision).Inc()
}

// PrometheusMiddleware collects basic HTTP metrics for Gin
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Wait for the next handlers to process the request
		c.Next()
		
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
