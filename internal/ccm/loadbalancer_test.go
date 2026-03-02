package ccm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoadBalancer(t *testing.T) {
	vpcID := uuid.New().String()
	
	// State for the fake API
	lbs := []sdk.LoadBalancer{}
	targets := make(map[string][]sdk.LBTarget)
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// t.Logf("FAKE API: %s %s", r.Method, r.URL.Path)
		
		// List LBs
		if r.Method == "GET" && r.URL.Path == "/lb" {
			json.NewEncoder(w).Encode(sdk.Response[[]sdk.LoadBalancer]{Data: lbs})
			return
		}
		
		// Create LB
		if r.Method == "POST" && r.URL.Path == "/lb" {
			var input struct {
				Name  string `json:"name"`
				VpcID string `json:"vpc_id"`
				Port  int    `json:"port"`
			}
			json.NewDecoder(r.Body).Decode(&input)
			newLB := sdk.LoadBalancer{
				ID:    uuid.New().String(),
				Name:  input.Name,
				VpcID: input.VpcID,
				Port:  input.Port,
			}
			lbs = append(lbs, newLB)
			json.NewEncoder(w).Encode(sdk.Response[sdk.LoadBalancer]{Data: newLB})
			return
		}

		// Delete LB
		if r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/lb/") && !strings.Contains(r.URL.Path, "/targets") {
			id := strings.TrimPrefix(r.URL.Path, "/lb/")
			for i, lb := range lbs {
				if lb.ID == id {
					lbs = append(lbs[:i], lbs[i+1:]...)
					break
				}
			}
			json.NewEncoder(w).Encode(sdk.Response[any]{})
			return
		}
		
		// Get Instance
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/instances/") {
			id := strings.TrimPrefix(r.URL.Path, "/instances/")
			inst := sdk.Instance{
				ID:    id,
				Name:  id,
				VpcID: vpcID,
			}
			json.NewEncoder(w).Encode(sdk.Response[sdk.Instance]{Data: inst})
			return
		}
		
		// List Instances
		if r.Method == "GET" && r.URL.Path == "/instances" {
			insts := []sdk.Instance{
				{ID: "inst-1", Name: "node-1", VpcID: vpcID},
				{ID: "inst-2", Name: "node-2", VpcID: vpcID},
				{ID: "inst-3", Name: "node-3", VpcID: vpcID},
			}
			json.NewEncoder(w).Encode(sdk.Response[[]sdk.Instance]{Data: insts})
			return
		}
		
		// List Targets
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/targets") {
			parts := strings.Split(r.URL.Path, "/")
			lbID := parts[2]
			json.NewEncoder(w).Encode(sdk.Response[[]sdk.LBTarget]{Data: targets[lbID]})
			return
		}
		
		// Add Target
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/targets") {
			parts := strings.Split(r.URL.Path, "/")
			lbID := parts[2]
			var input struct {
				InstanceID string `json:"instance_id"`
				Port       int    `json:"port"`
				Weight     int    `json:"weight"`
			}
			json.NewDecoder(r.Body).Decode(&input)
			targets[lbID] = append(targets[lbID], sdk.LBTarget{
				InstanceID: input.InstanceID,
				Port:       input.Port,
				Weight:     input.Weight,
			})
			json.NewEncoder(w).Encode(sdk.Response[any]{})
			return
		}

		// Remove Target
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/targets/") {
			parts := strings.Split(r.URL.Path, "/")
			lbID := parts[2]
			instID := parts[4]
			
			filtered := []sdk.LBTarget{}
			for _, t := range targets[lbID] {
				if t.InstanceID != instID {
					filtered = append(filtered, t)
				}
			}
			targets[lbID] = filtered
			json.NewEncoder(w).Encode(sdk.Response[any]{})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL, "test-key")
	lbProvider := newLoadBalancer(client)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{Port: 80, NodePort: 30001}},
		},
	}

	t.Run("EnsureLoadBalancer_Create", func(t *testing.T) {
		nodes := []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
		}

		status, err := lbProvider.EnsureLoadBalancer(context.Background(), "test-cluster", svc, nodes)
		require.NoError(t, err)
		assert.NotEmpty(t, status.Ingress)
		
		assert.Len(t, lbs, 1)
		lbID := lbs[0].ID
		assert.Len(t, targets[lbID], 2)
	})

	t.Run("GetLoadBalancer", func(t *testing.T) {
		status, exists, err := lbProvider.GetLoadBalancer(context.Background(), "test-cluster", svc)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.NotEmpty(t, status.Ingress)
	})

	t.Run("UpdateLoadBalancer_ScaleUp", func(t *testing.T) {
		nodes := []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-3"}},
		}

		err := lbProvider.UpdateLoadBalancer(context.Background(), "test-cluster", svc, nodes)
		require.NoError(t, err)
		
		lbID := lbs[0].ID
		assert.Len(t, targets[lbID], 3)
	})

	t.Run("UpdateLoadBalancer_ScaleDown", func(t *testing.T) {
		nodes := []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
		}

		err := lbProvider.UpdateLoadBalancer(context.Background(), "test-cluster", svc, nodes)
		require.NoError(t, err)
		
		lbID := lbs[0].ID
		assert.Len(t, targets[lbID], 1)
		assert.Equal(t, "inst-1", targets[lbID][0].InstanceID)
	})

	t.Run("EnsureLoadBalancerDeleted", func(t *testing.T) {
		err := lbProvider.EnsureLoadBalancerDeleted(context.Background(), "test-cluster", svc)
		require.NoError(t, err)
		assert.Len(t, lbs, 0)
	})
}
