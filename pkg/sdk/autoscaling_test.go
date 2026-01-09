package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_AutoScaling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "POST" && r.URL.Path == "/autoscaling/groups" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Response[ScalingGroup]{
				Data: ScalingGroup{ID: "asg-1", Name: "test-asg", Status: "ACTIVE"},
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/autoscaling/groups" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Response[[]ScalingGroup]{
				Data: []ScalingGroup{{ID: "asg-1", Name: "test-asg"}},
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/autoscaling/groups/asg-1" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Response[ScalingGroup]{
				Data: ScalingGroup{ID: "asg-1", Name: "test-asg", Status: "ACTIVE"},
			})
			return
		}

		if r.Method == "DELETE" && r.URL.Path == "/autoscaling/groups/asg-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/autoscaling/groups/asg-1/policies" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "DELETE" && r.URL.Path == "/autoscaling/policies/pol-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	t.Run("CreateScalingGroup", func(t *testing.T) {
		req := CreateScalingGroupRequest{
			Name:         "test-asg",
			VpcID:        "vpc-1",
			Image:        "nginx",
			MinInstances: 1,
			MaxInstances: 5,
		}
		asg, err := client.CreateScalingGroup(req)
		assert.NoError(t, err)
		assert.Equal(t, "asg-1", asg.ID)
	})

	t.Run("ListScalingGroups", func(t *testing.T) {
		groups, err := client.ListScalingGroups()
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run("GetScalingGroup", func(t *testing.T) {
		group, err := client.GetScalingGroup("asg-1")
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "asg-1", group.ID)
	})

	t.Run("DeleteScalingGroup", func(t *testing.T) {
		err := client.DeleteScalingGroup("asg-1")
		assert.NoError(t, err)
	})

	t.Run("CreateScalingPolicy", func(t *testing.T) {
		req := CreatePolicyRequest{
			Name:        "scale-out",
			MetricType:  "cpu",
			TargetValue: 70,
		}
		err := client.CreateScalingPolicy("asg-1", req)
		assert.NoError(t, err)
	})

	t.Run("DeleteScalingPolicy", func(t *testing.T) {
		err := client.DeleteScalingPolicy("pol-1")
		assert.NoError(t, err)
	})
}
