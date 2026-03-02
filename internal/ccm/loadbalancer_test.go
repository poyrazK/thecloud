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

	type apiState struct {
		lbs     []sdk.LoadBalancer
		targets map[string][]sdk.LBTarget
	}

	setupFakeAPI := func(t *testing.T, state *apiState) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// List LBs
			if r.Method == "GET" && r.URL.Path == "/lb" {
				if err := json.NewEncoder(w).Encode(sdk.Response[[]sdk.LoadBalancer]{Data: state.lbs}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Create LB
			if r.Method == "POST" && r.URL.Path == "/lb" {
				var input struct {
					Name  string `json:"name"`
					VpcID string `json:"vpc_id"`
					Port  int    `json:"port"`
				}
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				newLB := sdk.LoadBalancer{
					ID:    uuid.New().String(),
					Name:  input.Name,
					VpcID: input.VpcID,
					Port:  input.Port,
				}
				state.lbs = append(state.lbs, newLB)
				if err := json.NewEncoder(w).Encode(sdk.Response[sdk.LoadBalancer]{Data: newLB}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Delete LB
			if r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/lb/") && !strings.Contains(r.URL.Path, "/targets") {
				id := strings.TrimPrefix(r.URL.Path, "/lb/")
				for i, lb := range state.lbs {
					if lb.ID == id {
						state.lbs = append(state.lbs[:i], state.lbs[i+1:]...)
						break
					}
				}
				_ = json.NewEncoder(w).Encode(sdk.Response[any]{})
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
				if err := json.NewEncoder(w).Encode(sdk.Response[sdk.Instance]{Data: inst}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// List Instances
			if r.Method == "GET" && r.URL.Path == "/instances" {
				insts := []sdk.Instance{
					{ID: "inst-1", Name: "node-1", VpcID: vpcID},
					{ID: "inst-2", Name: "node-2", VpcID: vpcID},
					{ID: "inst-3", Name: "node-3", VpcID: vpcID},
				}
				if err := json.NewEncoder(w).Encode(sdk.Response[[]sdk.Instance]{Data: insts}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// List Targets
			if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/targets") {
				parts := strings.Split(r.URL.Path, "/")
				lbID := parts[2]
				if err := json.NewEncoder(w).Encode(sdk.Response[[]sdk.LBTarget]{Data: state.targets[lbID]}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
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
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				state.targets[lbID] = append(state.targets[lbID], sdk.LBTarget{
					InstanceID: input.InstanceID,
					Port:       input.Port,
					Weight:     input.Weight,
				})
				_ = json.NewEncoder(w).Encode(sdk.Response[any]{})
				return
			}

			// Remove Target
			if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/targets/") {
				parts := strings.Split(r.URL.Path, "/")
				lbID := parts[2]
				instID := parts[4]

				filtered := []sdk.LBTarget{}
				for _, t := range state.targets[lbID] {
					if t.InstanceID != instID {
						filtered = append(filtered, t)
					}
				}
				state.targets[lbID] = filtered
				_ = json.NewEncoder(w).Encode(sdk.Response[any]{})
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{Port: 80, NodePort: 30001}},
		},
	}

	t.Run("EnsureLoadBalancer_Create", func(t *testing.T) {
		state := &apiState{targets: make(map[string][]sdk.LBTarget)}
		server := setupFakeAPI(t, state)
		defer server.Close()

		client := sdk.NewClient(server.URL, "test-key")
		lbProvider := newLoadBalancer(client)

		nodes := []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
		}

		status, err := lbProvider.EnsureLoadBalancer(context.Background(), "test-cluster", svc, nodes)
		require.NoError(t, err)
		assert.NotEmpty(t, status.Ingress)

		assert.Len(t, state.lbs, 1)
		lbID := state.lbs[0].ID
		assert.Len(t, state.targets[lbID], 2)
	})

	t.Run("GetLoadBalancer", func(t *testing.T) {
		state := &apiState{
			lbs: []sdk.LoadBalancer{
				{ID: "lb-1", Name: "k8s-test-cluster-default-my-svc", VpcID: vpcID, Port: 80},
			},
			targets: make(map[string][]sdk.LBTarget),
		}
		server := setupFakeAPI(t, state)
		defer server.Close()

		client := sdk.NewClient(server.URL, "test-key")
		lbProvider := newLoadBalancer(client)

		status, exists, err := lbProvider.GetLoadBalancer(context.Background(), "test-cluster", svc)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.NotEmpty(t, status.Ingress)
	})

	t.Run("UpdateLoadBalancer_ScaleUp", func(t *testing.T) {
		state := &apiState{
			lbs: []sdk.LoadBalancer{
				{ID: "lb-1", Name: "k8s-test-cluster-default-my-svc", VpcID: vpcID, Port: 80},
			},
			targets: map[string][]sdk.LBTarget{
				"lb-1": {
					{InstanceID: "inst-1", Port: 30001, Weight: 1},
					{InstanceID: "inst-2", Port: 30001, Weight: 1},
				},
			},
		}
		server := setupFakeAPI(t, state)
		defer server.Close()

		client := sdk.NewClient(server.URL, "test-key")
		lbProvider := newLoadBalancer(client)

		nodes := []*v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node-3"}},
		}

		err := lbProvider.UpdateLoadBalancer(context.Background(), "test-cluster", svc, nodes)
		require.NoError(t, err)
		assert.Len(t, state.targets["lb-1"], 3)
	})

	t.Run("EnsureLoadBalancerDeleted", func(t *testing.T) {
		state := &apiState{
			lbs: []sdk.LoadBalancer{
				{ID: "lb-1", Name: "k8s-test-cluster-default-my-svc", VpcID: vpcID, Port: 80},
			},
			targets: make(map[string][]sdk.LBTarget),
		}
		server := setupFakeAPI(t, state)
		defer server.Close()

		client := sdk.NewClient(server.URL, "test-key")
		lbProvider := newLoadBalancer(client)

		err := lbProvider.EnsureLoadBalancerDeleted(context.Background(), "test-cluster", svc)
		require.NoError(t, err)
		assert.Empty(t, state.lbs)
	})
}
