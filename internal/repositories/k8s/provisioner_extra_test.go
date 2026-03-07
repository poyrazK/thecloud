package k8s

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRotateSecrets(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	clusterID := uuid.New()
	userID := uuid.New()
	cluster := &domain.Cluster{
		ID:     clusterID,
		UserID: userID,
		ControlPlaneIPs: []string{"10.0.0.1"},
	}

	t.Run("Success", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		repo := new(mockClusterRepo)
		secretSvc := new(mockSecretService)

		p := &KubeadmProvisioner{
			repo:      repo,
			secretSvc: secretSvc,
			logger:    logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		executor.On("Run", mock.Anything, "kubeadm certs renew all").Return("", nil).Once()
		executor.On("Run", mock.Anything, "cat "+adminKubeconfig).Return("new-kubeconfig", nil).Once()
		secretSvc.On("Encrypt", mock.Anything, userID, "new-kubeconfig").Return("encrypted-kubeconfig", nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.KubeconfigEncrypted == "encrypted-kubeconfig"
		})).Return(nil).Once()

		err := p.RotateSecrets(ctx, cluster)
		require.NoError(t, err)
		
		executor.AssertExpectations(t)
		repo.AssertExpectations(t)
		secretSvc.AssertExpectations(t)
	})

	t.Run("NoControlPlaneIPs", func(t *testing.T) {
		p := &KubeadmProvisioner{
			logger: logger,
		}
		err := p.RotateSecrets(ctx, &domain.Cluster{ControlPlaneIPs: []string{}})
		require.Error(t, err)
	})
}

func TestCreateBackup(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{"10.0.0.1"},
	}

	t.Run("Success", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		storage := new(mockStorageService)

		p := &KubeadmProvisioner{
			storageSvc: storage,
			logger:     logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		executor.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
			return strings.Contains(cmd, "snapshot save /tmp/snapshot.db")
		})).Return("", nil).Once()
		
		executor.On("Run", mock.Anything, "base64 /tmp/snapshot.db").Return("YmFja3VwLWRhdGE=", nil).Once()
		
		storage.On("Upload", mock.Anything, "k8s-backups", mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, clusterID.String())
		}), mock.Anything).Return(&domain.Object{}, nil).Once()

		err := p.CreateBackup(ctx, cluster)
		require.NoError(t, err)
		
		executor.AssertExpectations(t)
		storage.AssertExpectations(t)
	})
}

func TestFailCluster(t *testing.T) {
	ctx := context.Background()
	repo := new(mockClusterRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cluster := &domain.Cluster{ID: uuid.New()}

	p := &KubeadmProvisioner{
		repo:   repo,
		logger: logger,
	}

	repo.On("Update", ctx, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusFailed
	})).Return(nil).Once()

	err := p.failCluster(ctx, cluster, "test error", fmt.Errorf("underlying"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}
