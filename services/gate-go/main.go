package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rohanpatel2002/ironclad/services/gate-go/clients"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	topologyURL := os.Getenv("TOPOLOGY_URL")
	if topologyURL == "" {
		topologyURL = "http://localhost:8081"
	}

	scoringURL := os.Getenv("SCORING_URL")
	if scoringURL == "" {
		scoringURL = "http://localhost:8083"
	}

	// Wire dependencies
	topologyClient := clients.NewTopologyClient(topologyURL)
	scoringClient := clients.NewScoringClient(scoringURL)
	
	var deployRepo services.DeploymentRepository
	var riskRepo services.RiskScoreRepository

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		deployRepo = services.NewPostgresDeploymentRepository(db)
		riskRepo = services.NewPostgresRiskScoreRepository(db)
		log.Println("Connected to PostgreSQL for persistence")
	} else {
		deployRepo = services.NewNoopDeploymentRepository()
		riskRepo = services.NewNoopRiskScoreRepository()
		log.Println("Using in-memory no-op repositories (DATABASE_URL not set)")
	}

	decisionSvc := services.NewDecisionService(topologyClient, scoringClient, deployRepo, riskRepo)
	decisionHandler := handlers.NewDecisionHandler(decisionSvc)

	// Configure Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())

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

	port := os.Getenv("GATE_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 IRONCLAD Gate service starting on port %s\n", port)
	fmt.Printf("   Topology: %s\n", topologyURL)
	fmt.Printf("   Scoring:  %s\n", scoringURL)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// requestLogger returns a Gin middleware that logs each request with timing
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("[GATE] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
		)
	}
}
