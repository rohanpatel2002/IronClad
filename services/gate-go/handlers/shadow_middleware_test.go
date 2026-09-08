package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/handlers"
)

func TestShadowMiddleware_DisabledWhenNoEnv(t *testing.T) {
	t.Setenv("SHADOW_URL", "")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(handlers.ShadowMiddleware())
	router.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(body))
	})

	bodyStr := "hello"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(bodyStr))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != bodyStr {
		t.Fatalf("expected body %q, got %q", bodyStr, w.Body.String())
	}
}

func TestShadowMiddleware_ShadowsRequest(t *testing.T) {
	shadowCalled := make(chan bool, 1)

	// Create a dummy shadow server
	shadowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) == "shadow-body" && r.Header.Get("X-Custom") == "test" {
			shadowCalled <- true
		} else {
			shadowCalled <- false
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowServer.Close()

	t.Setenv("SHADOW_URL", shadowServer.URL)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(handlers.ShadowMiddleware())
	router.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(body))
	})

	bodyStr := "shadow-body"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(bodyStr))
	req.Header.Set("X-Custom", "test")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != bodyStr {
		t.Fatalf("expected main request body %q, got %q", bodyStr, w.Body.String())
	}

	// Wait for shadow request
	select {
	case success := <-shadowCalled:
		if !success {
			t.Fatal("shadow server received incorrect request")
		}
	case <-time.After(time.Second):
		t.Fatal("shadow request timed out")
	}
}

func TestIsShadowModeEnabled(t *testing.T) {
	t.Setenv("SHADOW_URL", "http://shadow:8080")
	if !handlers.IsShadowModeEnabled() {
		t.Errorf("expected true when SHADOW_URL is set")
	}
}

