package autoscaler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/autoscaler/protos"
	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAutoscalerServer_Refresh(t *testing.T) {
	clusterID := uuid.New().String()
	
	t.Run("Status_Running", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/clusters/%s", clusterID), r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data": {"id": "%s", "status": "RUNNING"}}`, clusterID)
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.Refresh(context.Background(), &protos.RefreshRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "RUNNING", server.lastStatus)
	})

	t.Run("Status_Repairing", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data": {"id": "%s", "status": "repairing"}}`, clusterID)
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.Refresh(context.Background(), &protos.RefreshRequest{})
		require.Error(t, err)
		assert.Nil(t, resp)
		
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unavailable, st.Code())
		assert.Contains(t, st.Message(), "restored/repaired")
	})
}

func TestAutoscalerServer_NodeGroups(t *testing.T) {
	clusterID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"data": {
					"id": "%s",
					"node_groups": [
						{"name": "pool-1", "min_size": 1, "max_size": 10, "current_size": 2}
					]
				}
			}`, clusterID)
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.NodeGroups(context.Background(), &protos.NodeGroupsRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.NodeGroups, 1)
		assert.Equal(t, "pool-1", resp.NodeGroups[0].Id)
	})
}

func TestAutoscalerServer_NodeGroupForNode(t *testing.T) {
	clusterID := uuid.New().String()
	instanceID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == fmt.Sprintf("/instances/%s", instanceID) {
				fmt.Fprintf(w, `{"data": {"id": "%s", "metadata": {"thecloud.io/node-group": "pool-1"}} }`, instanceID)
			} else {
				fmt.Fprintf(w, `{
					"data": {
						"id": "%s",
						"node_groups": [{"name": "pool-1", "min_size": 1, "max_size": 10}]
					}
				}`, clusterID)
			}
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.NodeGroupForNode(context.Background(), &protos.NodeGroupForNodeRequest{
			Node: &protos.ExternalGrpcNode{ProviderID: "thecloud://" + instanceID},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.NodeGroup)
		assert.Equal(t, "pool-1", resp.NodeGroup.Id)
	})
}

func TestAutoscalerServer_NodeGroupTargetSize(t *testing.T) {
	clusterID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"data": {
					"id": "%s",
					"node_groups": [{"name": "pool-1", "current_size": 5}]
				}
			}`, clusterID)
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.NodeGroupTargetSize(context.Background(), &protos.NodeGroupTargetSizeRequest{Id: "pool-1"})
		require.NoError(t, err)
		assert.Equal(t, int32(5), resp.TargetSize)
	})
}

func TestAutoscalerServer_NodeGroupIncreaseSize(t *testing.T) {
	clusterID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.Method {
			case "GET":
				fmt.Fprintf(w, `{"data": {"id": "%s", "status": "RUNNING", "node_groups": [{"name": "pool-1", "current_size": 2, "max_size": 10}]}}`, clusterID)
			case "PUT":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"data": {}}`)
			}
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		_, err := server.NodeGroupIncreaseSize(context.Background(), &protos.NodeGroupIncreaseSizeRequest{Id: "pool-1", Delta: 1})
		require.NoError(t, err)
	})
}

func TestAutoscalerServer_NodeGroupDeleteNodes(t *testing.T) {
	clusterID := uuid.New().String()
	instanceID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == "GET" {
				fmt.Fprintf(w, `{"data": {"id": "%s", "status": "RUNNING", "node_groups": [{"name": "pool-1", "current_size": 2}]}}`, clusterID)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"data": {}}`)
			} else if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"data": {}}`)
			}
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		_, err := server.NodeGroupDeleteNodes(context.Background(), &protos.NodeGroupDeleteNodesRequest{
			Id: "pool-1",
			Nodes: []*protos.ExternalGrpcNode{{ProviderID: "thecloud://" + instanceID}},
		})
		require.NoError(t, err)
	})
}

func TestAutoscalerServer_NodeGroupNodes(t *testing.T) {
	clusterID := uuid.New().String()
	instanceID := uuid.New().String()
	
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data": [{"id": "%s", "metadata": {"thecloud.io/cluster-id": "%s", "thecloud.io/node-group": "pool-1"}, "status": "RUNNING"}]}`, instanceID, clusterID)
		}))
		defer ts.Close()

		client := sdk.NewClient(ts.URL, "test-token")
		server := NewAutoscalerServer(client, clusterID)

		resp, err := server.NodeGroupNodes(context.Background(), &protos.NodeGroupNodesRequest{Id: "pool-1"})
		require.NoError(t, err)
		assert.Len(t, resp.Instances, 1)
		assert.Equal(t, "thecloud://"+instanceID, resp.Instances[0].Id)
	})
}

func TestAutoscalerServer_Misc(t *testing.T) {
	server := NewAutoscalerServer(nil, "id")
	
	t.Run("GPULabel", func(t *testing.T) {
		resp, err := server.GPULabel(context.Background(), &protos.GPULabelRequest{})
		require.NoError(t, err)
		assert.Equal(t, "thecloud.io/gpu", resp.Label)
	})

	t.Run("GetAvailableGPUTypes", func(t *testing.T) {
		resp, err := server.GetAvailableGPUTypes(context.Background(), &protos.GetAvailableGPUTypesRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Cleanup", func(t *testing.T) {
		resp, err := server.Cleanup(context.Background(), &protos.CleanupRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
