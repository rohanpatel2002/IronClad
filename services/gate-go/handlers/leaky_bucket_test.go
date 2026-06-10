package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLeakyBucket_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 1 request per 100ms, capacity 2
	lb := NewLeakyBucket(100*time.Millisecond, 2)
	defer lb.Stop()

	router := gin.New()
	router.Use(lb.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First request - OK
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "1.1.1.1:1234"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - OK (capacity is 2)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "1.1.1.1:1234"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request - Blocked (bucket full)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "1.1.1.1:1234"
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Wait for a "leak" (100ms)
	time.Sleep(150 * time.Millisecond)

	// Fourth request - OK (one slot leaked)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "1.1.1.1:1234"
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}

func TestLeakyBucket_Cleanup(t *testing.T) {
	lb := &LeakyBucket{
		rate:     100 * time.Millisecond,
		capacity: 2,
		buckets:  make(map[string]*bucket),
	}

	// Add a fresh bucket
	lb.buckets["1.1.1.1"] = &bucket{count: 1, lastUpdate: time.Now()}
	
	// Add a stale empty bucket
	lb.buckets["2.2.2.2"] = &bucket{count: 0, lastUpdate: time.Now().Add(-2 * time.Hour)}

	lb.cleanup()

	assert.Contains(t, lb.buckets, "1.1.1.1")
	assert.NotContains(t, lb.buckets, "2.2.2.2")
}

