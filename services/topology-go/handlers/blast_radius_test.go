package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/topology-go/graph"
)

type mockGraphProvider struct {
	g *graph.DependencyGraph
}

func (m *mockGraphProvider) GetGraph(ctx context.Context) (*graph.DependencyGraph, error) {
	return m.g, nil
}

func setupTestRouter() (*gin.Engine, *TopologyHandler) {
	gin.SetMode(gin.TestMode)
	g := graph.New()
	g.AddService(graph.ServiceNode{Name: "auth-service", Criticality: 0.9, DependsOn: []string{"db"}})
	g.AddService(graph.ServiceNode{Name: "db", Criticality: 1.0, DependsOn: []string{}})
	
	provider := &mockGraphProvider{g: g}
	handler := NewTopologyHandler(provider)
	
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)
	
	return router, handler
}

func TestHandleBlastRadius(t *testing.T) {
	router, _ := setupTestRouter()

	t.Run("Valid Request", func(t *testing.T) {
		reqBody := blastRadiusRequest{Service: "db"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/blast-radius", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp["blast_radius_score"] == nil {
			t.Errorf("Expected blast_radius_score in response")
		}
	})

	t.Run("Unknown Service", func(t *testing.T) {
		reqBody := blastRadiusRequest{Service: "unknown-service"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/blast-radius", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp["blast_radius_score"].(float64) != 0.5 {
			t.Errorf("Expected default score 0.5, got %v", resp["blast_radius_score"])
		}
		if resp["warning"] == nil {
			t.Errorf("Expected warning for unknown service")
		}
	})

	t.Run("Invalid Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/blast-radius", bytes.NewBuffer([]byte(`{invalid json`)))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", w.Code)
		}
	})
}
