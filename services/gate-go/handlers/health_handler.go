package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// HealthHandler manages service health diagnostics.
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *sql.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: rdb}
}

// LivenessCheck returns 200 if the service is running.
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

// ReadinessCheck returns 200 if all critical dependencies are available.
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	issues := make(map[string]string)

	// Check DB
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			slog.Warn("Health check failed: database unavailable", "error", err)
			issues["database"] = err.Error()
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			slog.Warn("Health check failed: redis unavailable", "error", err)
			issues["redis"] = err.Error()
		}
	}

	if len(issues) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "DOWN",
			"issues": issues,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

