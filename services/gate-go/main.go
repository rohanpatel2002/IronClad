package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "gate-go",
		})
	})

	// Stub endpoint for deployment decisions
	router.POST("/api/v1/decision", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"decision": "PENDING",
			"message": "Decision engine not yet implemented",
		})
	})

	// Get deployment decision details
	router.GET("/api/v1/decision/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"id": id,
			"status": "pending",
		})
	})

	port := os.Getenv("GATE_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Gate service starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
