package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	autoScaleContentType     = "Content-Type"
	autoScaleAppJSON         = "application/json"
	autoScaleAPIKey          = "test-key"
	autoScaleGroupID         = "group-1"
	autoScalePolicyID        = "policy-1"
	autoScaleGroupName       = "test-group"
	autoScalePath            = "/autoscaling/groups"
	autoScalePathPrefix      = "/autoscaling/groups/"
	policyPathSuffix         = "/policies"
)

func newAutoscalingTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(autoScaleContentType, autoScaleAppJSON)

		if handleAutoscalingGroup(w, r) {
			return
		}
		if handleAutoscalingPolicy(w, r) {
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func handleAutoscalingGroup(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost && r.URL.Path == autoScalePath {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Response[ScalingGroup]{
			Data: ScalingGroup{ID: autoScaleGroupID, Name: autoScaleGroupName},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == autoScalePath {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response[[]ScalingGroup]{
			Data: []ScalingGroup{{ID: autoScaleGroupID, Name: autoScaleGroupName}},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == autoScalePathPrefix+autoScaleGroupID {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response[ScalingGroup]{
			Data: ScalingGroup{ID: autoScaleGroupID, Name: autoScaleGroupName},
		})
		return true
	}
	if r.Method == http.MethodDelete && r.URL.Path == autoScalePathPrefix+autoScaleGroupID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func handleAutoscalingPolicy(w http.ResponseWriter, r *http.Request) bool {
	policyPath := autoScalePathPrefix + autoScaleGroupID + policyPathSuffix
	if r.Method == http.MethodPost && r.URL.Path == policyPath {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Response[ScalingPolicy]{
			Data: ScalingPolicy{ID: autoScalePolicyID, ScalingGroupID: autoScaleGroupID},
		})
		return true
	}
	if r.Method == http.MethodDelete && r.URL.Path == "/autoscaling/policies/"+autoScalePolicyID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func TestClientAutoscaling(t *testing.T) {
	server := newAutoscalingTestServer(t)
	defer server.Close()

	client := NewClient(server.URL, autoScaleAPIKey)

	t.Run("CreateScalingGroup", func(t *testing.T) {
		g, err := client.CreateScalingGroup(autoScaleGroupName, "vpc-1", "nginx", "80:80", 1, 5, 2)
		assert.NoError(t, err)
		assert.Equal(t, autoScaleGroupID, g.ID)
	})

	t.Run("ListScalingGroups", func(t *testing.T) {
		gs, err := client.ListScalingGroups()
		assert.NoError(t, err)
		assert.Len(t, gs, 1)
	})

	t.Run("GetScalingGroup", func(t *testing.T) {
		g, err := client.GetScalingGroup(autoScaleGroupID)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, autoScaleGroupID, g.ID)
	})

	t.Run("DeleteScalingGroup", func(t *testing.T) {
		err := client.DeleteScalingGroup(autoScaleGroupID)
		assert.NoError(t, err)
	})

	t.Run("CreateScalingPolicy", func(t *testing.T) {
		p, err := client.CreateScalingPolicy(autoScaleGroupID, "cpu-target", "cpu", 70.0, 1, 1, 300)
		assert.NoError(t, err)
		assert.Equal(t, autoScalePolicyID, p.ID)
	})

	t.Run("DeleteScalingPolicy", func(t *testing.T) {
		err := client.DeleteScalingPolicy(autoScalePolicyID)
		assert.NoError(t, err)
	})
}

func TestClientAutoscalingErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, autoScaleAPIKey)
	_, err := client.CreateScalingGroup("g", "v", "i", "p", 1, 2, 1)
	assert.Error(t, err)

	_, err = client.ListScalingGroups()
	assert.Error(t, err)

	_, err = client.GetScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	err = client.DeleteScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	_, err = client.CreateScalingPolicy(autoScaleGroupID, "p", "m", 50, 1, 1, 60)
	assert.Error(t, err)

	err = client.DeleteScalingPolicy(autoScalePolicyID)
	assert.Error(t, err)
}
