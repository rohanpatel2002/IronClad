package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/rohanpatel2002/ironclad/services/topology-go/clients"
	"github.com/rohanpatel2002/ironclad/services/topology-go/graph"
	"github.com/rohanpatel2002/ironclad/services/topology-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/topology-go/pkg/logger"
	"github.com/rohanpatel2002/ironclad/services/topology-go/pkg/tracing"
)

func main() {
	_ = godotenv.Load()

	log := logger.New()
	slog.SetDefault(log)

	// Init Tracing
	tp, err := tracing.InitTracer("topology-go")
	if err != nil {
		log.Error("Failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

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
	if tp != nil {
		router.Use(otelgin.Middleware("topology-go"))
	}

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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Info("IRONCLAD Topology service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start topology server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down topology server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Topology server forced to shutdown", "error", err)
	}

	log.Info("Topology server exiting")
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
