// Package services implements core business logic.
package services

import (
	"context"
	"log/slog"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/crypto"
)

// ClusterService implements the managed Kubernetes service.
type ClusterService struct {
	repo        ports.ClusterRepository
	provisioner ports.ClusterProvisioner
	vpcSvc      ports.VpcService
	instanceSvc ports.InstanceService
	secretSvc   ports.SecretService
	taskQueue   ports.TaskQueue
	logger      *slog.Logger
}

// ClusterServiceParams holds dependencies for ClusterService.
type ClusterServiceParams struct {
	Repo        ports.ClusterRepository
	Provisioner ports.ClusterProvisioner
	VpcSvc      ports.VpcService
	InstanceSvc ports.InstanceService
	SecretSvc   ports.SecretService
	TaskQueue   ports.TaskQueue
	Logger      *slog.Logger
}

// NewClusterService creates a new ClusterService with the provided parameters.
func NewClusterService(params ClusterServiceParams) (*ClusterService, error) {
	if params.TaskQueue == nil {
		return nil, errors.New(errors.Internal, "taskQueue cannot be nil")
	}
	return &ClusterService{
		repo:        params.Repo,
		provisioner: params.Provisioner,
		vpcSvc:      params.VpcSvc,
		instanceSvc: params.InstanceSvc,
		secretSvc:   params.SecretSvc,
		taskQueue:   params.TaskQueue,
		logger:      params.Logger,
	}, nil
}

// Default cluster configuration values.
const (
	defaultClusterVersion = "v1.29.0"
	defaultWorkerCount    = 2
)

// CreateCluster initiates the provisioning of a new Kubernetes cluster.
func (s *ClusterService) CreateCluster(ctx context.Context, params ports.CreateClusterParams) (*domain.Cluster, error) {
	// 1. Verify VPC exists and belongs to user
	vpc, err := s.vpcSvc.GetVPC(ctx, params.VpcID.String())
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "vpc not found", err)
	}

	// Default version if not specified
	if params.Version == "" {
		params.Version = defaultClusterVersion
	}
	if params.Workers == 0 {
		params.Workers = defaultWorkerCount
	}

	// 2. Generate SSH Key Pair for the cluster
	privKey, _, err := crypto.GenerateSSHKeyPair()
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate cluster ssh key", err)
	}

	// Encrypt the SSH key before storing
	encryptedKey, err := s.secretSvc.Encrypt(ctx, params.UserID, privKey)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to encrypt cluster ssh key", err)
	}

	// 3. Create cluster record in database
	cluster := &domain.Cluster{
		ID:               uuid.New(),
		Name:             params.Name,
		UserID:           params.UserID,
		VpcID:            vpc.ID,
		Version:          params.Version,
		WorkerCount:      params.Workers,
		NetworkIsolation: params.NetworkIsolation,
		HAEnabled:        params.HAEnabled,
		Status:           domain.ClusterStatusPending,
		SSHKey:           encryptedKey,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, cluster); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cluster record", err)
	}

	// 3. Enqueue provisioning job
	job := domain.ClusterJob{
		ClusterID: cluster.ID,
		UserID:    cluster.UserID,
		Type:      domain.ClusterJobProvision,
	}

	if err := s.taskQueue.Enqueue(ctx, "k8s_jobs", job); err != nil {
		s.logger.Error("failed to enqueue cluster provision job", "cluster_id", cluster.ID, "error", err)
		cluster.Status = domain.ClusterStatusFailed
		if updateErr := s.repo.Update(ctx, cluster); updateErr != nil {
			s.logger.Error("failed to update cluster status after enqueue failure", "cluster_id", cluster.ID, "error", updateErr)
		}
		return nil, errors.Wrap(errors.Internal, "failed to enqueue provisioning task", err)
	}

	return cluster, nil
}

// GetCluster retrieves cluster details by ID.
func (s *ClusterService) GetCluster(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	cluster, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New(errors.NotFound, "cluster not found")
	}
	return cluster, nil
}

// ListClusters retrieves all clusters for a user.
func (s *ClusterService) ListClusters(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// DeleteCluster removes a cluster and its associated resources.
func (s *ClusterService) DeleteCluster(ctx context.Context, id uuid.UUID) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	cluster.Status = domain.ClusterStatusDeleting
	cluster.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, cluster); err != nil {
		return errors.Wrap(errors.Internal, "failed to update cluster status", err)
	}

	// Enqueue deprovision job
	job := domain.ClusterJob{
		ClusterID: cluster.ID,
		UserID:    cluster.UserID,
		Type:      domain.ClusterJobDeprovision,
	}

	if err := s.taskQueue.Enqueue(ctx, "k8s_jobs", job); err != nil {
		return errors.Wrap(errors.Internal, "failed to enqueue cluster deprovision job", err)
	}

	return nil
}

// GetKubeconfig retrieves the encrypted kubeconfig for a cluster, optionally for a specific guest role.
func (s *ClusterService) GetKubeconfig(ctx context.Context, id uuid.UUID, role string) (string, error) {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return "", err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return "", errors.New(errors.InvalidInput, "kubeconfig is only available when cluster is running")
	}

	// For admin role, use the stored kubeconfig
	if role == "" || role == "admin" {
		if cluster.Kubeconfig == "" {
			return "", errors.New(errors.NotFound, "kubeconfig not found for cluster")
		}

		// Decrypt the kubeconfig
		decrypted, err := s.secretSvc.Decrypt(ctx, cluster.UserID, cluster.Kubeconfig)
		if err != nil {
			s.logger.Error("failed to decrypt kubeconfig", "cluster_id", cluster.ID, "error", err)
			return "", errors.Wrap(errors.Internal, "failed to decrypt kubeconfig", err)
		}
		return decrypted, nil
	}

	// For other roles, generate dynamically from provisioner
	return s.provisioner.GetKubeconfig(ctx, cluster, role)
}

// RepairCluster triggers a re-run of critical provisioning steps (CNI, kube-proxy patches).
func (s *ClusterService) RepairCluster(ctx context.Context, id uuid.UUID) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	go func() {
		bgCtx := context.Background()
		bgCtx = appcontext.WithUserID(bgCtx, cluster.UserID)
		s.logger.Info("starting cluster repair", "cluster_id", cluster.ID)

		if err := s.provisioner.Repair(bgCtx, cluster); err != nil {
			s.logger.Error("cluster repair failed", "cluster_id", cluster.ID, "error", err)
			return
		}

		s.logger.Info("cluster repair completed", "cluster_id", cluster.ID)
	}()

	return nil
}

// ScaleCluster adjusts the number of worker nodes in the cluster.
func (s *ClusterService) ScaleCluster(ctx context.Context, id uuid.UUID, workers int) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	if workers < 1 {
		return errors.New(errors.InvalidInput, "at least 1 worker required")
	}

	cluster.WorkerCount = workers
	cluster.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, cluster); err != nil {
		return err
	}

	go func() {
		bgCtx := context.Background()
		bgCtx = appcontext.WithUserID(bgCtx, cluster.UserID)
		s.logger.Info("starting cluster scale", "cluster_id", cluster.ID, "new_workers", workers)

		if err := s.provisioner.Scale(bgCtx, cluster); err != nil {
			s.logger.Error("cluster scaling failed", "cluster_id", cluster.ID, "error", err)
			return
		}

		s.logger.Info("cluster scaling completed", "cluster_id", cluster.ID)
	}()

	return nil
}

// GetClusterHealth retrieves the operational status of the cluster.
func (s *ClusterService) GetClusterHealth(ctx context.Context, id uuid.UUID) (*ports.ClusterHealth, error) {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.provisioner.GetHealth(ctx, cluster)
}

// UpgradeCluster initiates an asynchronous version upgrade.
func (s *ClusterService) UpgradeCluster(ctx context.Context, id uuid.UUID, version string) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return errors.New(errors.Conflict, "cluster must be in running state to upgrade")
	}

	if cluster.Version == version {
		return errors.New(errors.InvalidInput, "cluster is already at this version")
	}

	if err := s.validateUpgrade(cluster.Version, version); err != nil {
		return err
	}

	cluster.Status = domain.ClusterStatusUpgrading
	cluster.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, cluster); err != nil {
		return err
	}

	// Enqueue upgrade job
	job := domain.ClusterJob{
		ClusterID: cluster.ID,
		UserID:    cluster.UserID,
		Type:      domain.ClusterJobUpgrade,
		Version:   version,
	}

	if s.taskQueue == nil {
		return errors.New(errors.Internal, "task queue not initialized")
	}

	if err := s.taskQueue.Enqueue(ctx, "k8s_jobs", job); err != nil {
		return errors.Wrap(errors.Internal, "failed to enqueue cluster upgrade job", err)
	}

	return nil
}

func (s *ClusterService) RotateSecrets(ctx context.Context, id uuid.UUID) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return errors.New(errors.Conflict, "cluster must be in running state to rotate secrets")
	}

	cluster.Status = domain.ClusterStatusUpdating
	if err := s.repo.Update(ctx, cluster); err != nil {
		return err
	}

	if err := s.provisioner.RotateSecrets(ctx, cluster); err != nil {
		cluster.Status = domain.ClusterStatusFailed
		if updateErr := s.repo.Update(ctx, cluster); updateErr != nil {
			s.logger.Error("failed to update cluster status after secret rotation failure", "cluster_id", cluster.ID, "error", updateErr)
		}
		return err
	}

	cluster.Status = domain.ClusterStatusRunning
	return s.repo.Update(ctx, cluster)
}

func (s *ClusterService) CreateBackup(ctx context.Context, id uuid.UUID) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return errors.New(errors.Conflict, "cluster must be in running state to create backup")
	}

	// Backup is a side operation, we don't necessarily change status to "backing up"
	// unless it's a very long operation that blocks other things.
	return s.provisioner.CreateBackup(ctx, cluster)
}

func (s *ClusterService) RestoreBackup(ctx context.Context, id uuid.UUID, backupPath string) error {
	cluster, err := s.GetCluster(ctx, id)
	if err != nil {
		return err
	}

	if cluster.Status != domain.ClusterStatusRunning {
		return errors.New(errors.Conflict, "cluster must be in running state to restore backup")
	}

	cluster.Status = domain.ClusterStatusRepairing
	if err := s.repo.Update(ctx, cluster); err != nil {
		return err
	}

	if err := s.provisioner.Restore(ctx, cluster, backupPath); err != nil {
		cluster.Status = domain.ClusterStatusFailed
		if updateErr := s.repo.Update(ctx, cluster); updateErr != nil {
			s.logger.Error("failed to update cluster status after restoration failure", "cluster_id", cluster.ID, "error", updateErr)
		}
		return err
	}

	cluster.Status = domain.ClusterStatusRunning
	return s.repo.Update(ctx, cluster)
}

func (s *ClusterService) validateUpgrade(current, target string) error {
	curVer, err := semver.NewVersion(current)
	if err != nil {
		s.logger.Warn("cluster has invalid current version string", "version", current, "error", err)
		return nil // Proceed anyway if current version is malformed, risk is on the user
	}

	tarVer, err := semver.NewVersion(target)
	if err != nil {
		return errors.New(errors.InvalidInput, "invalid target version format")
	}

	if tarVer.LessThan(curVer) || tarVer.Equal(curVer) {
		return errors.New(errors.InvalidInput, "target version must be higher than current")
	}

	// kubeadm only supports +1 minor version
	if tarVer.Major() != curVer.Major() || tarVer.Minor()-curVer.Minor() > 1 {
		return errors.New(errors.InvalidInput, "cannot skip minor versions or upgrade across major versions")
	}

	return nil
}
