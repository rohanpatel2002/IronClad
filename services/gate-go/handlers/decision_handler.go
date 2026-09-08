package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "github.com/rohanpatel2002/ironclad/services/gate-go/pkg/errors"
	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/analytics"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/audit"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/cost"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/soar"
	distsync "github.com/rohanpatel2002/ironclad/services/gate-go/pkg/sync"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
	"github.com/go-redis/redis/v8"
)

// DecisionHandler holds dependencies for the decision HTTP layer
type DecisionHandler struct {
	svc         *services.DecisionService
	store       *decisionStore
	audit       *audit.AuditLogger
	anomaly     *analytics.DecisionStats
	soar        *soar.QuarantineManager
	cost        *cost.Optimizer
	redisClient *redis.Client
}

// decisionStore is a thread-safe in-memory cache for recent decisions
type decisionStore struct {
	mu      sync.RWMutex
	records map[string]*models.DeploymentDecision
}

func newDecisionStore() *decisionStore {
	return &decisionStore{records: make(map[string]*models.DeploymentDecision)}
}

func (s *decisionStore) save(d *models.DeploymentDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[d.DecisionID] = d
}

func (s *decisionStore) get(id string) (*models.DeploymentDecision, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.records[id]
	return d, ok
}

func (s *decisionStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}


// NewDecisionHandler creates a new handler with the given service and audit logger.
func NewDecisionHandler(
	svc *services.DecisionService,
	auditLogger *audit.AuditLogger,
	anomaly *analytics.DecisionStats,
	soar *soar.QuarantineManager,
	cost *cost.Optimizer,
	redisClient *redis.Client,
) *DecisionHandler {
	return &DecisionHandler{
		svc:         svc,
		store:       newDecisionStore(),
		audit:       auditLogger,
		anomaly:     anomaly,
		soar:        soar,
		cost:        cost,
		redisClient: redisClient,
	}
}

// RegisterRoutes attaches decision endpoints to the router group
func (h *DecisionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/decision", h.handleDecision)
	rg.GET("/decision/:id", h.handleGetDecision)
	rg.GET("/decisions", h.handleListDecisions)
}

// handleDecision evaluates a deployment request and returns a gate decision.
//
//	POST /api/v1/decision
func (h *DecisionHandler) handleDecision(c *gin.Context) {
	var req models.DeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.New(http.StatusBadRequest, apperrors.ErrInvalidRequest, err.Error()))
		return
	}

	// Cost Optimization Check
	if h.cost.ShouldDefer(time.Now()) {
		nextWindow := h.cost.GetNextWindow(time.Now())
		c.Header("X-IronClad-Cost-Advice", "peak_period_detected")
		c.Header("X-IronClad-Suggested-Window", nextWindow.Format(time.RFC3339))
	}

	// Concurrency Control via Distributed Lock
	lock := distsync.NewDistLock(h.redisClient, "deploy:"+req.Service)
	acquired, err := lock.Lock(c.Request.Context(), 30*time.Second)
	if err != nil {
		apperrors.Respond(c, apperrors.New(http.StatusInternalServerError, apperrors.ErrInternal, "lock acquisition error"))
		return
	}
	if !acquired {
		apperrors.Respond(c, apperrors.New(http.StatusConflict, apperrors.ErrInternal, "concurrent deployment in progress"))
		return
	}
	defer lock.Unlock(c.Request.Context())

	decision, err := h.svc.EvaluateDeployment(c.Request.Context(), &req)
	if err != nil {
		apperrors.Respond(c, apperrors.New(http.StatusInternalServerError, apperrors.ErrInternal, err.Error()))
		return
	}

	h.store.save(decision)
	RecordDecisionMetric(string(decision.Decision))

	// Autonomous Security Anomaly Detection
	val := 0.0
	if decision.Decision == models.DecisionAllow {
		val = 1.0
	}
	h.anomaly.Record(val)

	if h.anomaly.IsAnomalous(val) {
		h.soar.QuarantineService(c.Request.Context(), req.Service, "detected anomalous deployment decision pattern")
	}

	// Record Audit Log
	h.audit.Log(c.Request.Context(), audit.LogRecord{
		TenantID:      c.GetString("tenant_id"),
		Actor:         c.GetString("user"),
		Action:        "deploy_request",
		ResourceType:  "deployment",
		ResourceID:    decision.DecisionID,
		Status:        string(decision.Decision),
		Details: map[string]interface{}{
			"service": req.Service,
			"branch":  req.Branch,
			"commit":  req.CommitHash,
		},
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		CorrelationID: c.GetHeader("X-Correlation-ID"),
	})

	c.JSON(http.StatusOK, decision)
}

// handleGetDecision retrieves a previously made decision by ID.
//
//	GET /api/v1/decision/:id
func (h *DecisionHandler) handleGetDecision(c *gin.Context) {
	id := c.Param("id")
	d, ok := h.store.get(id)
	if !ok {
		apperrors.Respond(c, apperrors.New(http.StatusNotFound, apperrors.ErrNotFound, "Decision ID not found"))
		return
	}
	c.JSON(http.StatusOK, d)
}

// handleListDecisions returns all cached decisions (newest first).
//
//	GET /api/v1/decisions
func (h *DecisionHandler) handleListDecisions(c *gin.Context) {
	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	list := make([]*models.DeploymentDecision, 0, len(h.store.records))
	for _, v := range h.store.records {
		list = append(list, v)
	}

	// Sort newest first
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].DecisionTimestamp.After(list[i].DecisionTimestamp) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"decisions": list,
		"count":     len(list),
		"timestamp": time.Now().UTC(),
	})
}

