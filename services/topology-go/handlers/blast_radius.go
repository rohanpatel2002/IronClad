package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/topology-go/graph"
)

// TopologyHandler provides HTTP handlers for the topology/blast-radius API.
type TopologyHandler struct {
	graph *graph.DependencyGraph
}

// NewTopologyHandler creates a handler backed by the given graph.
func NewTopologyHandler(g *graph.DependencyGraph) *TopologyHandler {
	return &TopologyHandler{graph: g}
}

// RegisterRoutes attaches all topology routes to the given router group.
func (h *TopologyHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/blast-radius", h.handleBlastRadius)
	rg.GET("/services", h.handleListServices)
	rg.GET("/services/:name", h.handleGetService)
	rg.POST("/services", h.handleAddService)
}

// blastRadiusRequest is the body for POST /api/v1/blast-radius
type blastRadiusRequest struct {
	Service      string   `json:"service" binding:"required"`
	ChangedFiles []string `json:"changed_files"`
}

// handleBlastRadius computes the blast radius for a service.
//
//	POST /api/v1/blast-radius
func (h *TopologyHandler) handleBlastRadius(c *gin.Context) {
	var req blastRadiusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	_, known := h.graph.GetService(req.Service)
	if !known {
		c.JSON(http.StatusOK, gin.H{
			"service":            req.Service,
			"blast_radius_score": 0.5,
			"impacted_services":  []string{},
			"warning":            "Service not found in topology graph — using default score",
		})
		return
	}

	result := h.graph.ComputeBlastRadius(req.Service)
	RecordBlastRadiusTraversal()
	c.JSON(http.StatusOK, result)
}

// handleListServices returns all known services in the graph.
//
//	GET /api/v1/services
func (h *TopologyHandler) handleListServices(c *gin.Context) {
	services := h.graph.ListServices()
	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

// handleGetService returns a single service node with its connections.
//
//	GET /api/v1/services/:name
func (h *TopologyHandler) handleGetService(c *gin.Context) {
	name := c.Param("name")
	node, ok := h.graph.GetService(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Service not in topology graph",
		})
		return
	}

	result := h.graph.ComputeBlastRadius(name)

	c.JSON(http.StatusOK, gin.H{
		"service":            node,
		"blast_radius_score": result.BlastRadiusScore,
		"impacted_services":  result.ImpactedServices,
	})
}

// addServiceRequest is the body for POST /api/v1/services
type addServiceRequest struct {
	Name        string   `json:"name" binding:"required"`
	Criticality float64  `json:"criticality"`
	DependsOn   []string `json:"depends_on"`
}

// handleAddService adds a service to the live graph.
//
//	POST /api/v1/services
func (h *TopologyHandler) handleAddService(c *gin.Context) {
	var req addServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Criticality == 0 {
		req.Criticality = 0.5
	}

	h.graph.AddService(graph.ServiceNode{
		Name:        req.Name,
		Criticality: req.Criticality,
		DependsOn:   req.DependsOn,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service added to topology graph",
		"name":    req.Name,
	})
}
