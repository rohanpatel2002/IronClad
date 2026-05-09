package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/governance"
)

// GovernanceHandler handles requests for security compliance reports.
type GovernanceHandler struct {
	generator *governance.ReportGenerator
}

// NewGovernanceHandler creates a new governance handler.
func NewGovernanceHandler(generator *governance.ReportGenerator) *GovernanceHandler {
	return &GovernanceHandler{generator: generator}
}

// GenerateSOC2Report handles the GET /api/v1/governance/report endpoint.
func (h *GovernanceHandler) GenerateSOC2Report(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query parameters are required (YYYY-MM-DD)"})
		return
	}

	start, err := time.Parse(time.DateOnly, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse(time.DateOnly, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use YYYY-MM-DD"})
		return
	}

	// Ensure end date includes the whole day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	report, err := h.generator.GenerateSOC2Summary(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
