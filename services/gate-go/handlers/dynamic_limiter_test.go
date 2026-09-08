package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDynamicLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewDynamicLimiter(500)
	limiter.SetMaxMemoryMB(1024)

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 under normal memory conditions, got %d", w.Code)
	}
}
