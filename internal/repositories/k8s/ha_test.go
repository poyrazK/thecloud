package k8s

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockClusterRepoForHA struct{ mock.Mock }

type mockSecretService struct{ mock.Mock }

func (m *mockClusterRepoForHA) Create(ctx context.Context, c *domain.Cluster) error { return nil }
func (m *mockClusterRepoForHA) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepoForHA) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepoForHA) Update(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockClusterRepoForHA) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockClusterRepoForHA) AddNode(ctx context.Context, n *domain.ClusterNode) error {
	return nil
}
func (m *mockClusterRepoForHA) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	return nil, nil
}
func (m *mockClusterRepoForHA) DeleteNode(ctx context.Context, nodeID uuid.UUID) error { return nil }
func (m *mockClusterRepoForHA) UpdateNode(ctx context.Context, n *domain.ClusterNode) error {
	return nil
}

func (m *mockSecretService) CreateSecret(ctx context.Context, name, value, description string) (*domain.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) GetSecret(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) GetSecretByName(ctx context.Context, name string) (*domain.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) ListSecrets(ctx context.Context) ([]*domain.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) DeleteSecret(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSecretService) Encrypt(ctx context.Context, userID uuid.UUID, plainText string) (string, error) {
	args := m.Called(ctx, userID, plainText)
	return args.String(0), args.Error(1)
}
func (m *mockSecretService) Decrypt(ctx context.Context, userID uuid.UUID, cipherText string) (string, error) {
	return cipherText, nil
}

func TestParseJoinCommands(t *testing.T) {
	p := &KubeadmProvisioner{}
	output := `kubeadm join 10.0.0.10:6443 --token abc \
  --discovery-token-ca-cert-hash sha256:xxx

kubeadm join 10.0.0.10:6443 --token abc \
  --discovery-token-ca-cert-hash sha256:xxx \
  --control-plane --certificate-key 123`

	joinCmd, cpJoinCmd := p.parseJoinCommands(output)

	assert.Contains(t, joinCmd, "kubeadm join 10.0.0.10:6443")
	assert.Contains(t, joinCmd, "--discovery-token-ca-cert-hash sha256:xxx")
	assert.Contains(t, cpJoinCmd, "--control-plane")
	assert.Contains(t, cpJoinCmd, "--certificate-key 123")
}

func TestStoreKubeconfigSuccess(t *testing.T) {
	repo := new(mockClusterRepoForHA)
	secretSvc := new(mockSecretService)

	p := &KubeadmProvisioner{
		repo:      repo,
		secretSvc: secretSvc,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	cluster := &domain.Cluster{ID: uuid.New(), UserID: uuid.New()}

	secretSvc.On("Encrypt", mock.Anything, cluster.UserID, "raw").Return("enc", nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Kubeconfig == "enc"
	})).Return(nil)

	err := p.storeKubeconfig(context.Background(), cluster, "raw")

	assert.NoError(t, err)
	secretSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestStoreKubeconfigEncryptError(t *testing.T) {
	repo := new(mockClusterRepoForHA)
	secretSvc := new(mockSecretService)

	p := &KubeadmProvisioner{
		repo:      repo,
		secretSvc: secretSvc,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	cluster := &domain.Cluster{ID: uuid.New(), UserID: uuid.New()}

	secretSvc.On("Encrypt", mock.Anything, cluster.UserID, "raw").Return("", errors.New("boom"))

	err := p.storeKubeconfig(context.Background(), cluster, "raw")

	assert.Error(t, err)
	secretSvc.AssertExpectations(t)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestStoreKubeconfigUpdateError(t *testing.T) {
	repo := new(mockClusterRepoForHA)
	secretSvc := new(mockSecretService)

	p := &KubeadmProvisioner{
		repo:      repo,
		secretSvc: secretSvc,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	cluster := &domain.Cluster{ID: uuid.New(), UserID: uuid.New()}

	secretSvc.On("Encrypt", mock.Anything, cluster.UserID, "raw").Return("enc", nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Cluster")).Return(errors.New("db"))

	err := p.storeKubeconfig(context.Background(), cluster, "raw")

	assert.Error(t, err)
	secretSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}
