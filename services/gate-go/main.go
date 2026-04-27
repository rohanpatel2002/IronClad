package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rohanpatel2002/ironclad/services/gate-go/clients"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/auth"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/logger"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	log := logger.New()
	slog.SetDefault(log)

	topologyURL := os.Getenv("TOPOLOGY_URL")
	if topologyURL == "" {
		topologyURL = "http://localhost:8081"
	}

	scoringURL := os.Getenv("SCORING_URL")
	if scoringURL == "" {
		scoringURL = "http://localhost:8083"
	}

	semanticURL := os.Getenv("SEMANTIC_URL")
	if semanticURL == "" {
		semanticURL = "http://localhost:8082"
	}

	// Wire dependencies
	topologyClient := clients.NewTopologyClient(topologyURL)
	semanticClient := clients.NewSemanticClient(semanticURL)
	scoringClient := clients.NewScoringClient(scoringURL)

	var deployRepo services.DeploymentRepository
	var riskRepo services.RiskScoreRepository

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Error("Failed to open database", "error", err)
			os.Exit(1)
		}
		if err := db.Ping(); err != nil {
			log.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}
		deployRepo = services.NewPostgresDeploymentRepository(db)
		riskRepo = services.NewPostgresRiskScoreRepository(db)
		log.Info("Connected to PostgreSQL for persistence")
	} else {
		deployRepo = services.NewNoopDeploymentRepository()
		riskRepo = services.NewNoopRiskScoreRepository()
		log.Info("Using in-memory no-op repositories", "reason", "DATABASE_URL not set")
	}

	decisionSvc := services.NewDecisionService(topologyClient, semanticClient, scoringClient, deployRepo, riskRepo)
	decisionHandler := handlers.NewDecisionHandler(decisionSvc)
	webhookHandler := handlers.NewWebhookHandler(decisionSvc)
	jwtManager := auth.NewJWTManager()

	// Configure Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handlers.SecurityHeadersMiddleware())
	router.Use(handlers.RequestIDMiddleware())
	router.Use(structuredRequestLogger(log))
	router.Use(handlers.PrometheusMiddleware())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "gate-go",
			"timestamp": time.Now().UTC(),
			"version":   "0.1.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	decisionHandler.RegisterRoutes(v1)
	webhookHandler.RegisterRoutes(v1)

	// Protected management routes
	mgmt := v1.Group("/mgmt", handlers.AuthMiddleware(jwtManager))
	mgmt.GET("/circuit-breaker/status", handlers.CircuitBreakerStatusHandler())

	port := os.Getenv("GATE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("IRONCLAD Gate service starting",
		"port", port,
		"topology_url", topologyURL,
		"semantic_url", semanticURL,
		"scoring_url", scoringURL,
	)

	if err := router.Run(":" + port); err != nil {
		log.Error("Failed to start server", "error", err)
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

// Ensure fmt is used
var _ = fmt.Sprintf
