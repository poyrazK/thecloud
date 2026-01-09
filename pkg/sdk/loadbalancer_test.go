package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_LoadBalancer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "POST" && r.URL.Path == "/lb" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Response[LoadBalancer]{
				Data: LoadBalancer{ID: "lb-1", Name: "test-lb", Status: "ACTIVE"},
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/lb" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Response[[]LoadBalancer]{
				Data: []LoadBalancer{{ID: "lb-1", Name: "test-lb"}},
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/lb/lb-1" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Response[LoadBalancer]{
				Data: LoadBalancer{ID: "lb-1", Name: "test-lb", Status: "ACTIVE"},
			})
			return
		}

		if r.Method == "DELETE" && r.URL.Path == "/lb/lb-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/lb/lb-1/targets" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "DELETE" && r.URL.Path == "/lb/lb-1/targets/inst-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == "GET" && r.URL.Path == "/lb/lb-1/targets" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Response[[]LBTarget]{
				Data: []LBTarget{{InstanceID: "inst-1", Port: 80}},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	t.Run("CreateLB", func(t *testing.T) {
		lb, err := client.CreateLB("test-lb", "vpc-1", 80, "round-robin")
		assert.NoError(t, err)
		assert.Equal(t, "lb-1", lb.ID)
	})

	t.Run("ListLBs", func(t *testing.T) {
		lbs, err := client.ListLBs()
		assert.NoError(t, err)
		assert.Len(t, lbs, 1)
	})

	t.Run("GetLB", func(t *testing.T) {
		lb, err := client.GetLB("lb-1")
		assert.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, "lb-1", lb.ID)
	})

	t.Run("DeleteLB", func(t *testing.T) {
		err := client.DeleteLB("lb-1")
		assert.NoError(t, err)
	})

	t.Run("AddLBTarget", func(t *testing.T) {
		err := client.AddLBTarget("lb-1", "inst-1", 80, 1)
		assert.NoError(t, err)
	})

	t.Run("RemoveLBTarget", func(t *testing.T) {
		err := client.RemoveLBTarget("lb-1", "inst-1")
		assert.NoError(t, err)
	})

	t.Run("ListLBTargets", func(t *testing.T) {
		targets, err := client.ListLBTargets("lb-1")
		assert.NoError(t, err)
		assert.Len(t, targets, 1)
	})
}
