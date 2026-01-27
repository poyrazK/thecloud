package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	autoScaleContentType = "Content-Type"
	autoScaleAppJSON     = "application/json"
	autoScaleGroupID     = "asg-1"
	autoScaleGroupName   = "test-asg"
	autoScalePolicyID    = "pol-1"
	autoScaleAPIKey      = "test-key"
	autoScaleGroupPath   = "/autoscaling/groups"
	policyPathSuffix     = "/policies"
)

func newAutoscalingTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
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
	if r.Method == http.MethodPost && r.URL.Path == autoScaleGroupPath {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[ScalingGroup]{
			Data: ScalingGroup{ID: autoScaleGroupID, Name: autoScaleGroupName, Status: "ACTIVE"},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == autoScaleGroupPath {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]ScalingGroup]{
			Data: []ScalingGroup{{ID: autoScaleGroupID, Name: autoScaleGroupName}},
		})
		return true
	}
	if r.Method == http.MethodGet && r.URL.Path == autoScaleGroupPath+"/"+autoScaleGroupID {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[ScalingGroup]{
			Data: ScalingGroup{ID: autoScaleGroupID, Name: autoScaleGroupName, Status: "ACTIVE"},
		})
		return true
	}
	if r.Method == http.MethodDelete && r.URL.Path == autoScaleGroupPath+"/"+autoScaleGroupID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if r.Method == http.MethodPost && r.URL.Path == autoScaleGroupPath+"/"+autoScaleGroupID+policyPathSuffix {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func handleAutoscalingPolicy(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodDelete && r.URL.Path == "/autoscaling/policies/"+autoScalePolicyID {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func TestClientAutoScaling(t *testing.T) {
	server := newAutoscalingTestServer(t)
	defer server.Close()

	client := NewClient(server.URL, autoScaleAPIKey)

	t.Run("CreateScalingGroup", func(t *testing.T) {
		req := CreateScalingGroupRequest{
			Name:         autoScaleGroupName,
			VpcID:        "vpc-1",
			Image:        "nginx",
			MinInstances: 1,
			MaxInstances: 5,
		}
		asg, err := client.CreateScalingGroup(req)
		assert.NoError(t, err)
		assert.Equal(t, autoScaleGroupID, asg.ID)
	})

	t.Run("ListScalingGroups", func(t *testing.T) {
		groups, err := client.ListScalingGroups()
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run("GetScalingGroup", func(t *testing.T) {
		group, err := client.GetScalingGroup(autoScaleGroupID)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, autoScaleGroupID, group.ID)
	})

	t.Run("DeleteScalingGroup", func(t *testing.T) {
		err := client.DeleteScalingGroup(autoScaleGroupID)
		assert.NoError(t, err)
	})

	t.Run("CreateScalingPolicy", func(t *testing.T) {
		req := CreatePolicyRequest{
			Name:        "scale-out",
			MetricType:  "cpu",
			TargetValue: 70,
		}
		err := client.CreateScalingPolicy(autoScaleGroupID, req)
		assert.NoError(t, err)
	})

	t.Run("DeleteScalingPolicy", func(t *testing.T) {
		err := client.DeleteScalingPolicy(autoScalePolicyID)
		assert.NoError(t, err)
	})
}

func TestClientAutoScalingErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, autoScaleAPIKey)

	_, err := client.CreateScalingGroup(CreateScalingGroupRequest{Name: "test"})
	assert.Error(t, err)

	_, err = client.ListScalingGroups()
	assert.Error(t, err)

	_, err = client.GetScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	err = client.DeleteScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	err = client.CreateScalingPolicy(autoScaleGroupID, CreatePolicyRequest{Name: "scale"})
	assert.Error(t, err)

	err = client.DeleteScalingPolicy(autoScalePolicyID)
	assert.Error(t, err)
}

func TestClientAutoScalingRequestErrors(t *testing.T) {
	client := NewClient("http://127.0.0.1:0", autoScaleAPIKey)

	_, err := client.CreateScalingGroup(CreateScalingGroupRequest{Name: "asg"})
	assert.Error(t, err)

	_, err = client.ListScalingGroups()
	assert.Error(t, err)

	_, err = client.GetScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	err = client.DeleteScalingGroup(autoScaleGroupID)
	assert.Error(t, err)

	err = client.CreateScalingPolicy(autoScaleGroupID, CreatePolicyRequest{Name: "scale"})
	assert.Error(t, err)

	err = client.DeleteScalingPolicy(autoScalePolicyID)
	assert.Error(t, err)
}
