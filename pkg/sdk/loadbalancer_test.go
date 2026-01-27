package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	lbContentType    = "Content-Type"
	lbApplicationJSON = "application/json"
	lbID             = "lb-1"
	lbName           = "test-lb"
	lbInstanceID     = "inst-1"
	lbAPIKey         = "test-key"
	lbPath           = "/lb"
	lbPathPrefix     = "/lb/"
)

func newLoadBalancerTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		w.Header().Set(lbContentType, lbApplicationJSON)

		if handleLBBase(w, r) {
			return
		}
		if handleLBTargets(w, r) {
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func handleLBBase(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost && r.URL.Path == lbPath {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[LoadBalancer]{
			Data: LoadBalancer{ID: lbID, Name: lbName, Status: "ACTIVE"},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == lbPath {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]LoadBalancer]{
			Data: []LoadBalancer{{ID: lbID, Name: lbName}},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == lbPathPrefix+lbID {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[LoadBalancer]{
			Data: LoadBalancer{ID: lbID, Name: lbName, Status: "ACTIVE"},
		})
		return true
	}
	if r.Method == http.MethodDelete && r.URL.Path == lbPathPrefix+lbID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func handleLBTargets(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost && r.URL.Path == lbPathPrefix+lbID+"/targets" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	if r.Method == http.MethodDelete && r.URL.Path == lbPathPrefix+lbID+"/targets/"+lbInstanceID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == lbPathPrefix+lbID+"/targets" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]LBTarget]{
			Data: []LBTarget{{InstanceID: lbInstanceID, Port: 80}},
		})
		return true
	}
	return false
}

func TestClientLoadBalancer(t *testing.T) {
	server := newLoadBalancerTestServer(t)
	defer server.Close()

	client := NewClient(server.URL, lbAPIKey)

	t.Run("CreateLB", func(t *testing.T) {
		lb, err := client.CreateLB(lbName, "vpc-1", 80, "round-robin")
		assert.NoError(t, err)
		assert.Equal(t, lbID, lb.ID)
	})

	t.Run("ListLBs", func(t *testing.T) {
		lbs, err := client.ListLBs()
		assert.NoError(t, err)
		assert.Len(t, lbs, 1)
	})

	t.Run("GetLB", func(t *testing.T) {
		lb, err := client.GetLB(lbID)
		assert.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, lbID, lb.ID)
	})

	t.Run("DeleteLB", func(t *testing.T) {
		err := client.DeleteLB(lbID)
		assert.NoError(t, err)
	})

	t.Run("AddLBTarget", func(t *testing.T) {
		err := client.AddLBTarget(lbID, lbInstanceID, 80, 1)
		assert.NoError(t, err)
	})

	t.Run("RemoveLBTarget", func(t *testing.T) {
		err := client.RemoveLBTarget(lbID, lbInstanceID)
		assert.NoError(t, err)
	})

	t.Run("ListLBTargets", func(t *testing.T) {
		targets, err := client.ListLBTargets(lbID)
		assert.NoError(t, err)
		assert.Len(t, targets, 1)
	})
}

func TestClientLoadBalancerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, lbAPIKey)
	_, err := client.CreateLB("lb", "vpc-1", 80, "round-robin")
	assert.Error(t, err)

	_, err = client.ListLBs()
	assert.Error(t, err)

	_, err = client.GetLB(lbID)
	assert.Error(t, err)

	err = client.DeleteLB(lbID)
	assert.Error(t, err)

	err = client.AddLBTarget(lbID, lbInstanceID, 80, 1)
	assert.Error(t, err)

	err = client.RemoveLBTarget(lbID, lbInstanceID)
	assert.Error(t, err)

	_, err = client.ListLBTargets(lbID)
	assert.Error(t, err)
}
