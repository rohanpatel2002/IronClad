package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/models"
	"github.com/rohanpatel2002/ironclad/services/gate-go/services"
)

// DecisionHandler holds dependencies for the decision HTTP layer
type DecisionHandler struct {
	svc   *services.DecisionService
	store *decisionStore
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

// NewDecisionHandler creates a new handler with the given service
func NewDecisionHandler(svc *services.DecisionService) *DecisionHandler {
	return &DecisionHandler{svc: svc, store: newDecisionStore()}
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	decision, err := h.svc.EvaluateDeployment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "evaluation_failed",
			"message": err.Error(),
		})
		return
	}

	h.store.save(decision)

	c.JSON(http.StatusOK, decision)
}

// handleGetDecision retrieves a previously made decision by ID.
//
//	GET /api/v1/decision/:id
func (h *DecisionHandler) handleGetDecision(c *gin.Context) {
	id := c.Param("id")
	d, ok := h.store.get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Decision ID not found. Decisions are cached in-memory; restart will clear history.",
		})
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
