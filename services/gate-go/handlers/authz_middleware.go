package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type opaResponse struct {
	Result bool `json:"result"`
}

// AuthzMiddleware checks the OPA policy for authorization.
func AuthzMiddleware() gin.HandlerFunc {
	opaURL := os.Getenv("OPA_URL")
	if opaURL == "" {
		opaURL = "http://localhost:8181/v1/data/ironclad/authz/allow"
	}

	return func(c *gin.Context) {
		// Prepare input for OPA
		// In a real app, you'd extract user info from JWT
		input := map[string]interface{}{
			"input": map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"risk_score": c.GetFloat64("risk_score"), // Assuming set by previous middleware or handler
				"user": map[string]string{
					"role": "user", // Default
				},
			},
		}

		body, _ := json.Marshal(input)
		resp, err := http.Post(opaURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Authorization service unavailable"})
			return
		}
		defer resp.Body.Close()

		var opaResp opaResponse
		if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil || !opaResp.Result {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized by policy"})
			return
		}

		c.Next()
	}
}
