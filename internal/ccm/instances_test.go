package ccm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInstancesV2(t *testing.T) {
	// Setup a real HTTP server to act as the Cloud API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances/test-node" {
			inst := sdk.Instance{
				ID:           "inst-123",
				Name:         "test-node",
				PrivateIP:    "10.0.0.5",
				Status:       "RUNNING",
				InstanceType: "standard-2",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sdk.Response[sdk.Instance]{Data: inst})
			return
		}
		if r.URL.Path == "/instances/stopped-node" {
			inst := sdk.Instance{
				ID:     "inst-456",
				Status: "STOPPED",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sdk.Response[sdk.Instance]{Data: inst})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": {"message": "not found"}}`)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL, "test-key")
	insts := newInstancesV2(client)

	t.Run("InstanceMetadata", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		meta, err := insts.InstanceMetadata(context.Background(), node)
		require.NoError(t, err)
		assert.Equal(t, "thecloud://inst-123", meta.ProviderID)
		assert.Equal(t, "standard-2", meta.InstanceType)
		assert.Contains(t, meta.NodeAddresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.5"})
	})

	t.Run("InstanceExists", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		exists, err := insts.InstanceExists(context.Background(), node)
		require.NoError(t, err)
		assert.True(t, exists)

		nodeMissing := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "missing-node"}}
		exists, err = insts.InstanceExists(context.Background(), nodeMissing)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("InstanceShutdown", func(t *testing.T) {
		nodeRunning := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		shutdown, err := insts.InstanceShutdown(context.Background(), nodeRunning)
		require.NoError(t, err)
		assert.False(t, shutdown)

		nodeStopped := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "stopped-node"}}
		shutdown, err = insts.InstanceShutdown(context.Background(), nodeStopped)
		require.NoError(t, err)
		assert.True(t, shutdown)
	})
}
