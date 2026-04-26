package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rohanpatel2002/ironclad/services/topology-go/clients"
	"github.com/rohanpatel2002/ironclad/services/topology-go/graph"
	"github.com/rohanpatel2002/ironclad/services/topology-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/topology-go/pkg/logger"
)

func main() {
	_ = godotenv.Load()

	log := logger.New()
	slog.SetDefault(log)

	var provider handlers.GraphProvider

	k8sClient, err := clients.NewK8sClient()
	if err == nil {
		log.Info("Successfully connected to Kubernetes. Using dynamic K8s graph builder.")
		builder := graph.NewK8sGraphBuilder(k8sClient, 5*time.Minute)
		builder.StartBackgroundRefresher(context.Background())
		provider = builder
	} else {
		log.Warn("Failed to connect to Kubernetes. Falling back to default static graph.", "error", err)
		provider = &defaultGraphProvider{g: graph.NewDefault()}
	}

	topologyHandler := handlers.NewTopologyHandler(provider)

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(structuredRequestLogger(log))
	router.Use(handlers.PrometheusMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "topology-go",
			"timestamp": time.Now().UTC(),
			"version":   "0.1.0",
		})
	})

	v1 := router.Group("/api/v1")
	topologyHandler.RegisterRoutes(v1)

	port := os.Getenv("TOPOLOGY_PORT")
	if port == "" {
		port = "8081"
	}

	log.Info("IRONCLAD Topology service starting", "port", port)

	if err := router.Run(":" + port); err != nil {
		log.Error("Failed to start topology server", "error", err)
		os.Exit(1)
	}
}

// structuredRequestLogger returns a Gin middleware that emits structured JSON logs.
func structuredRequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		log.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", c.GetHeader("X-Request-ID"),
		)
	}
}

type defaultGraphProvider struct {
	g *graph.DependencyGraph
}

func (p *defaultGraphProvider) GetGraph(ctx context.Context) (*graph.DependencyGraph, error) {
	return p.g, nil
}

// Ensure fmt is used
var _ = fmt.Sprintf
