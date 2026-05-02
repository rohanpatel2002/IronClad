package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/rohanpatel2002/ironclad/services/gate-go/clients"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/auth"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/logger"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/mtls"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/tracing"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	log := logger.New()
	slog.SetDefault(log)

	// Init Tracing
	tp, err := tracing.InitTracer("gate-go")
	if err != nil {
		log.Error("Failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	// Init mTLS Config
	tlsCfg, err := mtls.LoadTLSConfig(mtls.Config{
		CACertFile: os.Getenv("MTLS_CA_CERT"),
		CertFile:   os.Getenv("MTLS_CERT"),
		KeyFile:    os.Getenv("MTLS_KEY"),
		ServerName: os.Getenv("MTLS_SERVER_NAME"),
	})
	if err != nil {
		log.Warn("Failed to load mTLS config, proceeding with insecure communication", "error", err)
	} else if tlsCfg != nil {
		log.Info("mTLS enabled for microservice communication")
	}

	// Init Redis
	var redisClient *redis.Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Error("Failed to parse REDIS_URL", "error", err)
		} else {
			redisClient = redis.NewClient(opt)
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				log.Error("Failed to connect to Redis", "error", err)
				redisClient = nil
			} else {
				log.Info("Connected to Redis for distributed rate limiting")
			}
		}
	}

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
	topologyClient := clients.NewTopologyClient(topologyURL, tlsCfg, redisClient)
	semanticClient := clients.NewSemanticClient(semanticURL, tlsCfg)
	scoringClient := clients.NewScoringClient(scoringURL, tlsCfg)

	var deployRepo services.DeploymentRepository
	var riskRepo services.RiskScoreRepository

	dbURL := os.Getenv("DATABASE_URL")
	replicaURL := os.Getenv("DATABASE_REPLICA_URL")

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

		var dbReplica *sql.DB
		if replicaURL != "" {
			dbReplica, err = sql.Open("postgres", replicaURL)
			if err != nil {
				log.Warn("Failed to open replica database, falling back to master", "error", err)
			} else if err := dbReplica.Ping(); err != nil {
				log.Warn("Failed to connect to replica database, falling back to master", "error", err)
				dbReplica = nil
			}
		}

		deployRepo = services.NewPostgresDeploymentRepository(db, dbReplica)
		riskRepo = services.NewPostgresRiskScoreRepository(db, dbReplica)
		log.Info("Connected to PostgreSQL for persistence (read-splitting enabled)")
	} else {
		deployRepo = services.NewNoopDeploymentRepository()
		riskRepo = services.NewNoopRiskScoreRepository()
		log.Info("Using in-memory no-op repositories", "reason", "DATABASE_URL not set")
	}

	decisionSvc := services.NewDecisionService(topologyClient, semanticClient, scoringClient, deployRepo, riskRepo)
	decisionHandler := handlers.NewDecisionHandler(decisionSvc)
	webhookHandler := handlers.NewWebhookHandler(decisionSvc, redisClient)
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
	if tp != nil {
		router.Use(otelgin.Middleware("gate-go"))
	}

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
	
	// Expose pprof on a separate group or within mgmt
	debug := mgmt.Group("/debug/pprof")
	debug.GET("/", gin.WrapH(http.DefaultServeMux))
	debug.GET("/cmdline", gin.WrapH(http.DefaultServeMux))
	debug.GET("/profile", gin.WrapH(http.DefaultServeMux))
	debug.GET("/symbol", gin.WrapH(http.DefaultServeMux))
	debug.GET("/trace", gin.WrapH(http.DefaultServeMux))
	debug.GET("/allocs", gin.WrapH(http.DefaultServeMux))
	debug.GET("/block", gin.WrapH(http.DefaultServeMux))
	debug.GET("/goroutine", gin.WrapH(http.DefaultServeMux))
	debug.GET("/heap", gin.WrapH(http.DefaultServeMux))
	debug.GET("/mutex", gin.WrapH(http.DefaultServeMux))
	debug.GET("/threadcreate", gin.WrapH(http.DefaultServeMux))

	port := os.Getenv("GATE_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Info("IRONCLAD Gate service starting",
			"port", port,
			"topology_url", topologyURL,
			"semantic_url", semanticURL,
			"scoring_url", scoringURL,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exiting")
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
