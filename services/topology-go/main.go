package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rohanpatel2002/ironclad/services/topology-go/clients"
	"github.com/rohanpatel2002/ironclad/services/topology-go/graph"
	"github.com/rohanpatel2002/ironclad/services/topology-go/handlers"
)

func main() {
	_ = godotenv.Load()

	var provider handlers.GraphProvider

	k8sClient, err := clients.NewK8sClient()
	if err == nil {
		log.Println("Successfully connected to Kubernetes. Using dynamic K8s graph builder.")
		builder := graph.NewK8sGraphBuilder(k8sClient, 5*time.Minute)
		builder.StartBackgroundRefresher(context.Background())
		provider = builder
	} else {
		log.Printf("Failed to connect to Kubernetes: %v. Falling back to default static graph.", err)
		provider = &defaultGraphProvider{g: graph.NewDefault()}
	}

	topologyHandler := handlers.NewTopologyHandler(provider)

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
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

	fmt.Printf("🗺  IRONCLAD Topology service starting on port %s\n", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start topology server: %v", err)
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[TOPOLOGY] %s %s %d %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

type defaultGraphProvider struct {
	g *graph.DependencyGraph
}

func (p *defaultGraphProvider) GetGraph(ctx context.Context) (*graph.DependencyGraph, error) {
	return p.g, nil
}
