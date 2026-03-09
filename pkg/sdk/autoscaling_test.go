package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	autoScaleContentType = "Content-Type"
	autoScaleAppJSON     = "application/json"
	autoScaleAPIKey      = "test-key"
	autoScaleGroupID     = "group-1"
	autoScalePolicyID    = "policy-1"
	autoScaleGroupName   = "test-group"
	autoScalePath        = "/autoscaling/groups"
	autoScalePathPrefix  = "/autoscaling/groups/"
	policyPathSuffix     = "/policies"
)

func newAutoscalingTestServer(t *testing.T) *httptest.Server {
	t.Helper()
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
		// API returns simple Success response for policy creation (it returns error only in SDK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
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
		req := CreateScalingGroupRequest{
			Name:         autoScaleGroupName,
			VpcID:        "vpc-1",
			Image:        "nginx",
			Ports:        "80:80",
			MinInstances: 1,
			MaxInstances: 5,
			DesiredCount: 2,
		}
		g, err := client.CreateScalingGroup(req)
		require.NoError(t, err)
		if g != nil {
			assert.Equal(t, autoScaleGroupID, g.ID)
		}
	})

	t.Run("ListScalingGroups", func(t *testing.T) {
		gs, err := client.ListScalingGroups()
		require.NoError(t, err)
		assert.Len(t, gs, 1)
	})

	t.Run("GetScalingGroup", func(t *testing.T) {
		g, err := client.GetScalingGroup(autoScaleGroupID)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, autoScaleGroupID, g.ID)
	})

	t.Run("DeleteScalingGroup", func(t *testing.T) {
		err := client.DeleteScalingGroup(autoScaleGroupID)
		require.NoError(t, err)
	})

	t.Run("CreateScalingPolicy", func(t *testing.T) {
		req := CreatePolicyRequest{
			Name:        "cpu-target",
			MetricType:  "cpu",
			TargetValue: 70.0,
			ScaleOut:    1,
			ScaleIn:     1,
			CooldownSec: 300,
		}
		err := client.CreateScalingPolicy(autoScaleGroupID, req)
		require.NoError(t, err)
	})

	t.Run("DeleteScalingPolicy", func(t *testing.T) {
		err := client.DeleteScalingPolicy(autoScalePolicyID)
		require.NoError(t, err)
	})
}

func TestClientAutoscalingErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, autoScaleAPIKey)
	_, err := client.CreateScalingGroup(CreateScalingGroupRequest{Name: "g"})
	require.Error(t, err)

	_, err = client.ListScalingGroups()
	require.Error(t, err)

	_, err = client.GetScalingGroup(autoScaleGroupID)
	require.Error(t, err)

	err = client.DeleteScalingGroup(autoScaleGroupID)
	require.Error(t, err)

	err = client.CreateScalingPolicy(autoScaleGroupID, CreatePolicyRequest{Name: "p"})
	require.Error(t, err)

	err = client.DeleteScalingPolicy(autoScalePolicyID)
	require.Error(t, err)
}
