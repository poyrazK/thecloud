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
	"github.com/stretchr/testify/require"
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
			ID:                   uuid.New(),
			UserID:               userID,
			VpcID:                uuid.New(),
			Name:                 testClusterName,
			Version:              testClusterVersion,
			ControlPlaneIPs:      []string{"10.0.0.1"},
			WorkerCount:          3,
			Status:               domain.ClusterStatusRunning,
			PodCIDR:              "10.244.0.0/16",
			ServiceCIDR:          "10.96.0.0/12",
			NetworkIsolation:     true,
			HAEnabled:            false,
			APIServerLBAddress:   nil,
			JobID:                nil,
			BackupSchedule:       "@daily",
			BackupRetentionDays:  7,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		mock.ExpectExec("INSERT INTO clusters").
			WithArgs(cluster.ID, cluster.UserID, cluster.VpcID, cluster.Name, cluster.Version,
				string(cluster.Status), cluster.ControlPlaneIPs, cluster.WorkerCount, cluster.HAEnabled,
				cluster.NetworkIsolation, cluster.PodCIDR, cluster.ServiceCIDR,
				cluster.APIServerLBAddress, cluster.KubeconfigEncrypted,
				cluster.SSHPrivateKeyEncrypted, cluster.JoinToken, cluster.TokenExpiresAt,
				cluster.CACertHash, cluster.JobID, cluster.BackupSchedule, cluster.BackupRetentionDays,
				cluster.CreatedAt, cluster.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.Create(ctx, cluster)
		require.NoError(t, err)
	})

	t.Run("Read Operations", func(t *testing.T) {
		clusterID := uuid.New()
		cols := []string{"id", "user_id", "vpc_id", "name", "version", "status", "control_plane_ips", "worker_count", "ha_enabled", "network_isolation", "pod_cidr", "service_cidr", "api_server_lb_address", "kubeconfig_encrypted", "ssh_private_key_encrypted", "join_token", "token_expires_at", "ca_cert_hash", "job_id", "backup_schedule", "backup_retention_days", "created_at", "updated_at"}
		ngCols := []string{"id", "cluster_id", "name", "instance_type", "min_size", "max_size", "current_size", "created_at", "updated_at"}

		testCases := []struct {
			name          string
			setupMock     func(mock pgxmock.PgxPoolIface)
			callFn        func(repo *ClusterRepository) (any, error)
			validate      func(t *testing.T, res any)
		}{
			{
				name: "GetByID",
				setupMock: func(mock pgxmock.PgxPoolIface) {
					t.Helper()
					mock.ExpectQuery("SELECT .* FROM clusters").WithArgs(clusterID, userID).
						WillReturnRows(pgxmock.NewRows(cols).
							AddRow(clusterID, userID, uuid.New(), testClusterName, testClusterVersion, string(domain.ClusterStatusRunning), []string{"10.0.0.1"}, 3, false, false, "10.244.0.0/16", "10.96.0.0/12", nil, "", "", "", nil, "", nil, "@daily", 7, time.Now(), time.Now()))
					mock.ExpectQuery("SELECT .* FROM cluster_node_groups").WithArgs(clusterID).
						WillReturnRows(pgxmock.NewRows(ngCols).
							AddRow(uuid.New(), clusterID, "default-pool", "standard-1", 1, 10, 3, time.Now(), time.Now()))
				},
				callFn: func(repo *ClusterRepository) (any, error) {
					return repo.GetByID(ctx, clusterID)
				},
				validate: func(t *testing.T, res any) {
					t.Helper()
					cluster := res.(*domain.Cluster)
					assert.NotNil(t, cluster)
					assert.Equal(t, clusterID, cluster.ID)
					assert.Len(t, cluster.NodeGroups, 1)
				},
			},
			{
				name: "ListAll",
				setupMock: func(mock pgxmock.PgxPoolIface) {
					t.Helper()
					mock.ExpectQuery("SELECT .* FROM clusters").
						WillReturnRows(pgxmock.NewRows(cols).
							AddRow(clusterID, userID, uuid.New(), "c1", "v1", "RUNNING", []string{}, 3, false, false, "", "", nil, "", "", "", nil, "", nil, "", 0, time.Now(), time.Now()))
					mock.ExpectQuery("SELECT .* FROM cluster_node_groups").WithArgs(clusterID).
						WillReturnRows(pgxmock.NewRows(ngCols).
							AddRow(uuid.New(), clusterID, "default-pool", "standard-1", 1, 10, 3, time.Now(), time.Now()))
				},
				callFn: func(repo *ClusterRepository) (any, error) {
					return repo.ListAll(ctx)
				},
				validate: func(t *testing.T, res any) {
					t.Helper()
					clusters := res.([]*domain.Cluster)
					assert.Len(t, clusters, 1)
					assert.Len(t, clusters[0].NodeGroups, 1)
				},
			},
			{
				name: "ListByUserID",
				setupMock: func(mock pgxmock.PgxPoolIface) {
					t.Helper()
					mock.ExpectQuery("SELECT .* FROM clusters WHERE user_id = \\$1").WithArgs(userID).
						WillReturnRows(pgxmock.NewRows(cols).
							AddRow(clusterID, userID, uuid.New(), "c1", "v1", "RUNNING", []string{}, 3, false, false, "", "", nil, "", "", "", nil, "", nil, "", 0, time.Now(), time.Now()))
					mock.ExpectQuery("SELECT .* FROM cluster_node_groups").WithArgs(clusterID).
						WillReturnRows(pgxmock.NewRows(ngCols).
							AddRow(uuid.New(), clusterID, "default-pool", "standard-1", 1, 10, 3, time.Now(), time.Now()))
				},
				callFn: func(repo *ClusterRepository) (any, error) {
					return repo.ListByUserID(ctx, userID)
				},
				validate: func(t *testing.T, res any) {
					t.Helper()
					clusters := res.([]*domain.Cluster)
					assert.Len(t, clusters, 1)
					assert.Len(t, clusters[0].NodeGroups, 1)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mock, _ := pgxmock.NewPool()
				defer mock.Close()
				repo := NewClusterRepository(mock)
				tc.setupMock(mock)
				res, err := tc.callFn(repo)
				require.NoError(t, err)
				tc.validate(t, res)
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewClusterRepository(mock)
		id := uuid.New()
		cluster := &domain.Cluster{ID: id, UserID: userID, Status: domain.ClusterStatusFailed, ControlPlaneIPs: []string{}}

		mock.ExpectExec("UPDATE clusters").
			WithArgs(cluster.VpcID, cluster.Name, cluster.Version, string(cluster.Status),
				cluster.ControlPlaneIPs, cluster.WorkerCount, cluster.HAEnabled, cluster.NetworkIsolation,
				cluster.PodCIDR, cluster.ServiceCIDR, cluster.APIServerLBAddress,
				cluster.KubeconfigEncrypted, cluster.SSHPrivateKeyEncrypted,
				cluster.JoinToken, cluster.TokenExpiresAt, cluster.CACertHash,
				cluster.JobID, cluster.BackupSchedule, cluster.BackupRetentionDays,
				pgxmock.AnyArg(), cluster.ID, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(ctx, cluster)
		require.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewClusterRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM clusters").WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, id)
		require.NoError(t, err)
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
			WithArgs(nodeID, clusterID, pgxmock.AnyArg(), "worker", "active", pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		err := repo.AddNode(ctx, &domain.ClusterNode{ID: nodeID, ClusterID: clusterID, Role: domain.NodeRoleWorker, Status: "active", JoinedAt: now, InstanceID: uuid.New()})
		require.NoError(t, err)

		// GetNodes
		mock.ExpectQuery("SELECT .* FROM cluster_nodes").WithArgs(clusterID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "cluster_id", "instance_id", "role", "status", "joined_at", "last_heartbeat"}).
				AddRow(nodeID, clusterID, uuid.New(), "worker", "active", now, nil))
		nodes, err := repo.GetNodes(ctx, clusterID)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)

		// UpdateNode
		mock.ExpectExec("UPDATE cluster_nodes").
			WithArgs("error", pgxmock.AnyArg(), pgxmock.AnyArg(), nodeID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err = repo.UpdateNode(ctx, &domain.ClusterNode{ID: nodeID, Status: "error"})
		require.NoError(t, err)

		// DeleteNode
		mock.ExpectExec("DELETE FROM cluster_nodes").WithArgs(nodeID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		err = repo.DeleteNode(ctx, nodeID)
		require.NoError(t, err)
	})
}
