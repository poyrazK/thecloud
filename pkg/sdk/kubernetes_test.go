package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	kubeTestContentType = "Content-Type"
	kubeTestAppJSON     = "application/json"
	kubeTestClusters    = "/clusters"
	kubeTestClusterName = "c1"
	kubeTestKubeconfig  = "kubeconfig"
	kubeTestRoleAdmin   = "admin"
	kubeTestBackupPath  = "s3://bucket/backup"
)

func TestClientListClusters(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubeTestClusters, r.URL.Path)

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[[]*Cluster]{Data: []*Cluster{{ID: clusterID, Name: kubeTestClusterName}}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	clusters, err := client.ListClusters()

	assert.NoError(t, err)
	assert.Len(t, clusters, 1)
	assert.Equal(t, clusterID, clusters[0].ID)
}

func TestClientCreateCluster(t *testing.T) {
	clusterID := uuid.New()
	vpcID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, kubeTestClusters, r.URL.Path)

		var payload CreateClusterInput
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, kubeTestClusterName, payload.Name)
		assert.Equal(t, vpcID, payload.VpcID)
		assert.Equal(t, 3, payload.WorkerCount)
		assert.True(t, payload.HA)

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[*Cluster]{Data: &Cluster{ID: clusterID, Name: kubeTestClusterName}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	cluster, err := client.CreateCluster(&CreateClusterInput{
		Name:        kubeTestClusterName,
		VpcID:       vpcID,
		Version:     "1.29",
		WorkerCount: 3,
		HA:          true,
	})

	assert.NoError(t, err)
	assert.Equal(t, clusterID, cluster.ID)
}

func TestClientGetCluster(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubeTestClusters+"/"+clusterID.String(), r.URL.Path)

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[*Cluster]{Data: &Cluster{ID: clusterID, Name: kubeTestClusterName}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	cluster, err := client.GetCluster(clusterID.String())

	assert.NoError(t, err)
	assert.Equal(t, clusterID, cluster.ID)
}

func TestClientDeleteCluster(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, kubeTestClusters+"/"+clusterID.String(), r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteCluster(clusterID.String())

	assert.NoError(t, err)
}

func TestClientGetKubeconfig(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubeTestClusters+"/"+clusterID.String()+"/kubeconfig", r.URL.Path)
		assert.Equal(t, kubeTestRoleAdmin, r.URL.Query().Get("role"))

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[string]{Data: kubeTestKubeconfig})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	config, err := client.GetKubeconfig(clusterID.String(), kubeTestRoleAdmin)

	assert.NoError(t, err)
	assert.Equal(t, kubeTestKubeconfig, config)
}

func TestClientGetKubeconfigNoRole(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubeTestClusters+"/"+clusterID.String()+"/kubeconfig", r.URL.Path)
		assert.Equal(t, "", r.URL.RawQuery)

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[string]{Data: kubeTestKubeconfig})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	config, err := client.GetKubeconfig(clusterID.String(), "")

	assert.NoError(t, err)
	assert.Equal(t, kubeTestKubeconfig, config)
}

func TestClientGetKubeconfigErrorStatus(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	_, err := client.GetKubeconfig(clusterID.String(), kubeTestRoleAdmin)

	assert.Error(t, err)
}

func TestClientRepairScaleUpgradeRotateBackupRestoreCluster(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case kubeTestClusters + "/" + clusterID.String() + "/repair":
			assert.Equal(t, http.MethodPost, r.Method)
		case kubeTestClusters + "/" + clusterID.String() + "/scale":
			assert.Equal(t, http.MethodPost, r.Method)
			var payload ScaleClusterInput
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, 5, payload.Workers)
		case kubeTestClusters + "/" + clusterID.String() + "/upgrade":
			assert.Equal(t, http.MethodPost, r.Method)
			var payload UpgradeClusterInput
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, "1.30", payload.Version)
		case kubeTestClusters + "/" + clusterID.String() + "/rotate-secrets":
			assert.Equal(t, http.MethodPost, r.Method)
		case kubeTestClusters + "/" + clusterID.String() + "/backups":
			assert.Equal(t, http.MethodPost, r.Method)
		case kubeTestClusters + "/" + clusterID.String() + "/restore":
			assert.Equal(t, http.MethodPost, r.Method)
			var payload RestoreBackupInput
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, kubeTestBackupPath, payload.BackupPath)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)

	assert.NoError(t, client.RepairCluster(clusterID.String()))
	assert.NoError(t, client.ScaleCluster(clusterID.String(), 5))
	assert.NoError(t, client.UpgradeCluster(clusterID.String(), "1.30"))
	assert.NoError(t, client.RotateSecrets(clusterID.String()))
	assert.NoError(t, client.CreateBackup(clusterID.String()))
	assert.NoError(t, client.RestoreBackup(clusterID.String(), kubeTestBackupPath))
}

func TestClientGetClusterHealth(t *testing.T) {
	clusterID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, kubeTestClusters+"/"+clusterID.String()+"/health", r.URL.Path)

		w.Header().Set(kubeTestContentType, kubeTestAppJSON)
		json.NewEncoder(w).Encode(Response[*ClusterHealth]{Data: &ClusterHealth{Status: "ok", NodesReady: 3, NodesTotal: 3}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	status, err := client.GetClusterHealth(clusterID.String())

	assert.NoError(t, err)
	assert.Equal(t, "ok", status.Status)
	assert.Equal(t, 3, status.NodesReady)
}

func TestClientClusterErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	_, err := client.ListClusters()
	assert.Error(t, err)

	_, err = client.CreateCluster(&CreateClusterInput{Name: kubeTestClusterName})
	assert.Error(t, err)

	_, err = client.GetCluster("cluster-1")
	assert.Error(t, err)

	_, err = client.GetClusterHealth("cluster-1")
	assert.Error(t, err)
}
