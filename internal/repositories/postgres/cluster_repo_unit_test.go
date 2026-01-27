package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

const (
	testClusterName    = "cluster-1"
	testClusterVersion = "v1.29.0"
)

func TestClusterRepository(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("Create", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewClusterRepository(mock)
		cluster := &domain.Cluster{
			ID:                 uuid.New(),
			UserID:             userID,
			VpcID:              uuid.New(),
			Name:               testClusterName,
			Version:            testClusterVersion,
			ControlPlaneIPs:    []string{},
			WorkerCount:        3,
			Status:             domain.ClusterStatusRunning,
			SSHKey:             "ssh",
			Kubeconfig:         "kube",
			NetworkIsolation:   true,
			HAEnabled:          false,
			APIServerLBAddress: nil,
			JobID:              nil,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		mock.ExpectExec("INSERT INTO clusters").
			WithArgs(cluster.ID, cluster.UserID, cluster.VpcID, cluster.Name, cluster.Version, cluster.ControlPlaneIPs, cluster.WorkerCount,
				string(cluster.Status), cluster.SSHKey, cluster.Kubeconfig, cluster.NetworkIsolation, cluster.HAEnabled, cluster.APIServerLBAddress, cluster.JobID, cluster.CreatedAt, cluster.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.Create(ctx, cluster)
		assert.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewClusterRepository(mock)
		id := uuid.New()

		mock.ExpectQuery("SELECT .* FROM clusters").WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "version", "control_plane_ips", "worker_count", "status", "ssh_key", "kubeconfig", "network_isolation", "ha_enabled", "api_server_lb_address", "job_id", "created_at", "updated_at"}).
				AddRow(id, userID, uuid.New(), testClusterName, testClusterVersion, []string{}, 3, string(domain.ClusterStatusRunning), "", "", false, false, nil, nil, time.Now(), time.Now()))

		cluster, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, cluster)
	})

	t.Run("NodeOps", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewClusterRepository(mock)
		clusterID := uuid.New()
		nodeID := uuid.New()
		now := time.Now()

		// AddNode
		mock.ExpectExec("INSERT INTO cluster_nodes").
			WithArgs(nodeID, clusterID, pgxmock.AnyArg(), "worker", "active", now).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		err := repo.AddNode(ctx, &domain.ClusterNode{ID: nodeID, ClusterID: clusterID, Role: domain.NodeRoleWorker, Status: "active", JoinedAt: now, InstanceID: uuid.New()})
		assert.NoError(t, err)

		// GetNodes
		mock.ExpectQuery("SELECT .* FROM cluster_nodes").WithArgs(clusterID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "cluster_id", "instance_id", "role", "status", "joined_at"}).
				AddRow(nodeID, clusterID, uuid.New(), "worker", "active", now))
		nodes, err := repo.GetNodes(ctx, clusterID)
		assert.NoError(t, err)
		assert.Len(t, nodes, 1)
	})
}
