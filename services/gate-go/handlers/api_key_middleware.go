package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/auth"
)

// APIKeyMiddleware validates requests using X-API-Key header.
func APIKeyMiddleware(manager *auth.APIKeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.Next() // Allow fallthrough to JWT middleware if needed
			return
		}

		tenantID, err := manager.ValidateKey(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			return
		}

		c.Set("tenant_id", tenantID)
		c.Set("user", "api-key-actor")
		c.Set("role", "admin") // API keys usually have elevated permissions for automation
		c.Next()
	}
}

