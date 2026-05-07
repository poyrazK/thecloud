package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientListRouteTables(t *testing.T) {
	vpcID := "vpc-123"
	expectedRTs := []RouteTable{
		{ID: "rt-1", VPCID: vpcID, Name: "main-rt", IsMain: true},
		{ID: "rt-2", VPCID: vpcID, Name: "custom-rt", IsMain: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/route-tables", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, vpcID, r.URL.Query().Get("vpc_id"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Response[[]RouteTable]{Data: expectedRTs})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	rts, err := client.ListRouteTables(vpcID)

	require.NoError(t, err)
	assert.Len(t, rts, 2)
	assert.Equal(t, expectedRTs[0].ID, rts[0].ID)
	assert.True(t, rts[0].IsMain)
}

func TestClientCreateRouteTable(t *testing.T) {
	vpcID := "vpc-123"
	expectedRT := RouteTable{
		ID:     "rt-456",
		VPCID:  vpcID,
		Name:   "new-rt",
		IsMain: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/route-tables", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, vpcID, req["vpc_id"])
		assert.Equal(t, "new-rt", req["name"])
		assert.Equal(t, false, req["is_main"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Response[RouteTable]{Data: expectedRT})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	rt, err := client.CreateRouteTable(vpcID, "new-rt", false)

	require.NoError(t, err)
	assert.NotNil(t, rt)
	assert.Equal(t, expectedRT.ID, rt.ID)
}

func TestClientGetRouteTable(t *testing.T) {
	id := "rt-123"
	expectedRT := RouteTable{
		ID:    id,
		VPCID: "vpc-456",
		Name:  "test-rt",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/route-tables/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Response[RouteTable]{Data: expectedRT})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	rt, err := client.GetRouteTable(id)

	require.NoError(t, err)
	assert.NotNil(t, rt)
	assert.Equal(t, expectedRT.ID, rt.ID)
}

func TestClientDeleteRouteTable(t *testing.T) {
	id := "rt-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/route-tables/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteRouteTable(id)

	require.NoError(t, err)
}
