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
			Name: "ironclad_topology_http_requests_total",
			Help: "Total number of HTTP requests processed by topology-go",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ironclad_topology_http_request_duration_seconds",
			Help:    "Histogram of response latency (seconds) of topology-go HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	blastRadiusTraversals = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ironclad_topology_blast_radius_traversals_total",
			Help: "Total number of graph traversals for blast radius calculation",
		},
	)
)

// RecordBlastRadiusTraversal increments the traversal counter
func RecordBlastRadiusTraversal() {
	blastRadiusTraversals.Inc()
}

// PrometheusMiddleware collects basic HTTP metrics for Gin
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
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
