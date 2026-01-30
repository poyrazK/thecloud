package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateGatewayRoute(t *testing.T) {
	expectedRoute := GatewayRoute{
		ID:          "route-1",
		Name:        "test-route",
		PathPrefix:  "/api/v1",
		TargetURL:   "http://backend:8080",
		StripPrefix: true,
		RateLimit:   100,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/gateway/routes", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req struct {
			Name        string `json:"name"`
			PathPrefix  string `json:"path_prefix"`
			TargetURL   string `json:"target_url"`
			StripPrefix bool   `json:"strip_prefix"`
			RateLimit   int    `json:"rate_limit"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedRoute.Name, req.Name)
		assert.Equal(t, expectedRoute.PathPrefix, req.PathPrefix)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedRoute)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	route, err := client.CreateGatewayRoute("test-route", "/api/v1", "http://backend:8080", true, 100)

	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, expectedRoute.ID, route.ID)
}

func TestClient_ListGatewayRoutes(t *testing.T) {
	expectedRoutes := []GatewayRoute{
		{ID: "route-1", Name: "route-1"},
		{ID: "route-2", Name: "route-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/gateway/routes", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedRoutes)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	routes, err := client.ListGatewayRoutes()

	assert.NoError(t, err)
	assert.Len(t, routes, 2)
}

func TestClient_DeleteGatewayRoute(t *testing.T) {
	id := "route-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/gateway/routes/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteGatewayRoute(id)

	assert.NoError(t, err)
}
