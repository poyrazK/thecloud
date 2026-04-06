// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/pkg/util"
)

const (
	// Default ports for database engines
	DefaultPostgresPort = "5432"
	DefaultMySQLPort    = "3306"

	// Connection Pooling (PgBouncer) defaults
	PoolerImage          = "edoburu/pgbouncer:latest"
	PoolerInternalPort   = "5432"
	DefaultPoolMode      = "transaction"
	DefaultMaxClientConn = "1000"
	DefaultPoolSize      = "20"

	// Exporter defaults
	PostgresExporterImage = "prometheuscommunity/postgres-exporter"
	PostgresExporterPort  = "9187"
	MySQLExporterImage    = "prom/mysqld-exporter"
	MySQLExporterPort     = "9104"
)

// DatabaseService manages database instances and lifecycle.
type DatabaseService struct {
	repo         ports.DatabaseRepository
	rbacSvc      ports.RBACService
	compute      ports.ComputeBackend
	vpcRepo      ports.VpcRepository
	volumeSvc    ports.VolumeService
	snapshotSvc  ports.SnapshotService
	snapshotRepo ports.SnapshotRepository
	eventSvc     ports.EventService
	auditSvc     ports.AuditService
	logger       *slog.Logger
}

// DatabaseServiceParams holds dependencies for DatabaseService creation.
type DatabaseServiceParams struct {
	Repo         ports.DatabaseRepository
	RBAC         ports.RBACService
	Compute      ports.ComputeBackend
	VpcRepo      ports.VpcRepository
	VolumeSvc    ports.VolumeService
	SnapshotSvc  ports.SnapshotService
	SnapshotRepo ports.SnapshotRepository
	EventSvc     ports.EventService
	AuditSvc     ports.AuditService
	Logger       *slog.Logger
}

// NewDatabaseService constructs a DatabaseService with its dependencies.
func NewDatabaseService(params DatabaseServiceParams) *DatabaseService {
	return &DatabaseService{
		repo:         params.Repo,
		rbacSvc:      params.RBAC,
		compute:      params.Compute,
		vpcRepo:      params.VpcRepo,
		volumeSvc:    params.VolumeSvc,
		snapshotSvc:  params.SnapshotSvc,
		snapshotRepo: params.SnapshotRepo,
		eventSvc:     params.EventSvc,
		auditSvc:     params.AuditSvc,
		logger:       params.Logger,
	}
}

func (s *DatabaseService) CreateDatabase(ctx context.Context, req ports.CreateDatabaseRequest) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBCreate, "*"); err != nil {
		return nil, err
	}

	dbEngine := domain.DatabaseEngine(req.Engine)

	if err := s.validateCreationRequest(req, dbEngine); err != nil {
		return nil, err
	}

	password, err := util.GenerateRandomPassword(16)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate password", err)
	}

	username := s.getDefaultUsername(dbEngine)
	db := s.initialDatabaseRecord(userID, req.Name, dbEngine, req.Version, username, password, req.VpcID)
	db.TenantID = tenantID
	db.Role = domain.RolePrimary
	db.AllocatedStorage = req.AllocatedStorage
	db.Parameters = req.Parameters
	db.MetricsEnabled = req.MetricsEnabled
	db.PoolingEnabled = req.PoolingEnabled

	imageName, env, defaultPort := s.getEngineConfig(dbEngine, req.Version, username, password, req.Name, db.Role, "")

	networkID, err := s.resolveVpcNetwork(ctx, req.VpcID)
	if err != nil {
		return nil, err
	}

	vol, err := s.volumeSvc.CreateVolume(ctx, fmt.Sprintf("db-vol-%s", db.ID.String()[:8]), req.AllocatedStorage)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create persistent volume", err)
	}

	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        fmt.Sprintf("cloud-db-%s-%s", req.Name, db.ID.String()[:8]),
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: []string{fmt.Sprintf("%s:%s", s.getBackendVolName(vol), s.getMountPath(dbEngine))},
		Env:         env,
		Cmd:         s.buildEngineCmd(dbEngine, req.Parameters),
	})

	if err != nil {
		s.cleanupVolumeQuietly(ctx, vol.ID.String())
		return nil, errors.Wrap(errors.Internal, "failed to launch database container", err)
	}

	db.ContainerID = containerID
	if err := s.resolveDatabasePort(ctx, db, allocatedPorts, defaultPort); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to resolve database port", err))
	}
	db.Status = domain.DatabaseStatusRunning

	if db.MetricsEnabled || db.PoolingEnabled {
		dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
		if err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to get database IP", err))
		}

		if err := s.provisionSidecars(ctx, db, dbEngine, dbIP, username, password, networkID); err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
		}
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
	}

	s.recordDatabaseCreation(ctx, userID, db, req.Engine)
	return db, nil
}

func (s *DatabaseService) validateCreationRequest(req ports.CreateDatabaseRequest, engine domain.DatabaseEngine) error {
	if !s.isValidEngine(engine) {
		return errors.New(errors.InvalidInput, "unsupported database engine")
	}
	if req.AllocatedStorage < 10 {
		return errors.New(errors.InvalidInput, "allocated storage must be at least 10GB")
	}
	if req.PoolingEnabled && engine != domain.EnginePostgres {
		return errors.New(errors.InvalidInput, "connection pooling is currently only supported for PostgreSQL")
	}
	return nil
}

func (s *DatabaseService) CreateReplica(ctx context.Context, primaryID uuid.UUID, name string) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBCreate, "*"); err != nil {
		return nil, err
	}

	primary, err := s.repo.GetByID(ctx, primaryID)
	if err != nil {
		return nil, err
	}

	primaryIP, err := s.compute.GetInstanceIP(ctx, primary.ContainerID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get primary IP", err)
	}

	db := s.initialDatabaseRecord(userID, name, primary.Engine, primary.Version, primary.Username, primary.Password, primary.VpcID)
	db.TenantID = tenantID
	db.Role = domain.RoleReplica
	db.PrimaryID = &primaryID
	db.AllocatedStorage = primary.AllocatedStorage
	db.MetricsEnabled = primary.MetricsEnabled
	db.PoolingEnabled = primary.PoolingEnabled

	imageName, env, defaultPort := s.getEngineConfig(primary.Engine, primary.Version, primary.Username, primary.Password, name, db.Role, primaryIP)

	networkID, err := s.resolveVpcNetwork(ctx, db.VpcID)
	if err != nil {
		return nil, err
	}

	// Create persistent volume for the replica
	volName := fmt.Sprintf("db-replica-vol-%s", db.ID.String()[:8])
	vol, err := s.volumeSvc.CreateVolume(ctx, volName, db.AllocatedStorage)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create persistent volume for replica", err)
	}

	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        fmt.Sprintf("cloud-db-%s-%s", name, db.ID.String()[:8]),
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: []string{fmt.Sprintf("%s:%s", s.getBackendVolName(vol), s.getMountPath(db.Engine))},
		Env:         env,
		Cmd:         nil,
	})
	if err != nil {
		s.cleanupVolumeQuietly(ctx, vol.ID.String())
		return nil, errors.Wrap(errors.Internal, "failed to launch replica container", err)
	}

	db.ContainerID = containerID
	if err := s.resolveDatabasePort(ctx, db, allocatedPorts, defaultPort); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to resolve replica database port", err))
	}
	db.Status = domain.DatabaseStatusRunning

	if db.MetricsEnabled || db.PoolingEnabled {
		dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
		if err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to get replica IP for sidecars", err))
		}

		if err := s.provisionSidecars(ctx, db, db.Engine, dbIP, db.Username, db.Password, networkID); err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
		}
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
	}

	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_REPLICA_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{
		"primary_id": primaryID,
		"name":       name,
	}); err != nil {
		s.logger.Warn("failed to record event", "action", "DATABASE_REPLICA_CREATE", "db_id", db.ID, "error", err)
	}

	return db, nil
}

func (s *DatabaseService) ModifyDatabase(ctx context.Context, req ports.ModifyDatabaseRequest) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBUpdate, req.ID.String()); err != nil {
		return nil, err
	}

	db, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Parameters != nil {
		db.Parameters = req.Parameters
	}

	if req.AllocatedStorage != nil {
		if *req.AllocatedStorage < db.AllocatedStorage {
			return nil, errors.New(errors.InvalidInput, "cannot decrease allocated storage")
		}
		vol, err := s.getVolumeForDatabase(ctx, db)
		if err != nil {
			return nil, err
		}
		if err := s.volumeSvc.ResizeVolume(ctx, vol.ID.String(), *req.AllocatedStorage); err != nil {
			return nil, err
		}
		db.AllocatedStorage = *req.AllocatedStorage
	}

	networkID, _ := s.resolveVpcNetwork(ctx, db.VpcID)
	dbIP, _ := s.compute.GetInstanceIP(ctx, db.ContainerID)

	if req.MetricsEnabled != nil && *req.MetricsEnabled != db.MetricsEnabled {
		if *req.MetricsEnabled {
			if err := s.provisionMetricsSidecar(ctx, db, db.Engine, dbIP, db.Username, db.Password, networkID); err != nil {
				return nil, err
			}
		} else if db.ExporterContainerID != "" {
			if err := s.compute.DeleteInstance(ctx, db.ExporterContainerID); err != nil {
				s.logger.Warn("failed to delete metrics sidecar during modification", "container_id", db.ExporterContainerID, "error", err)
			}
			db.ExporterContainerID = ""
			db.MetricsPort = 0
		}
		db.MetricsEnabled = *req.MetricsEnabled
	}

	if req.PoolingEnabled != nil && *req.PoolingEnabled != db.PoolingEnabled {
		if *req.PoolingEnabled {
			if db.Engine != domain.EnginePostgres {
				return nil, errors.New(errors.InvalidInput, "connection pooling is currently only supported for PostgreSQL")
			}
			if err := s.provisionPoolerSidecar(ctx, db, db.Engine, dbIP, db.Username, db.Password, networkID); err != nil {
				return nil, err
			}
		} else if db.PoolerContainerID != "" {
			if err := s.compute.DeleteInstance(ctx, db.PoolerContainerID); err != nil {
				s.logger.Warn("failed to delete pooler sidecar during modification", "container_id", db.PoolerContainerID, "error", err)
			}
			db.PoolerContainerID = ""
			db.PoolingPort = 0
		}
		db.PoolingEnabled = *req.PoolingEnabled
	}

	if err := s.repo.Update(ctx, db); err != nil {
		return nil, err
	}

	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_MODIFY", db.ID.String(), "DATABASE", nil); err != nil {
		s.logger.Warn("failed to record event", "action", "DATABASE_MODIFY", "db_id", db.ID, "error", err)
	}
	if err := s.auditSvc.Log(ctx, db.UserID, "database.modify", "database", db.ID.String(), map[string]interface{}{"name": db.Name}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "database.modify", "db_id", db.ID, "error", err)
	}

	return db, nil
}

func (s *DatabaseService) GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBRead, id.String()); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *DatabaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBRead, "*"); err != nil {
		return nil, err
	}

	return s.repo.List(ctx)
}

func (s *DatabaseService) DeleteDatabase(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBDelete, id.String()); err != nil {
		return err
	}

	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if db.ContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.ContainerID); err != nil {
			s.logger.Warn("failed to delete database container", "container_id", db.ContainerID, "error", err)
		}
	}
	if db.ExporterContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.ExporterContainerID); err != nil {
			s.logger.Warn("failed to delete exporter container", "container_id", db.ExporterContainerID, "error", err)
		}
	}
	if db.PoolerContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.PoolerContainerID); err != nil {
			s.logger.Warn("failed to delete pooler container", "container_id", db.PoolerContainerID, "error", err)
		}
	}

	vols, err := s.volumeSvc.ListVolumes(ctx)
	if err == nil {
		expectedPrefix := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
		if db.Role == domain.RoleReplica {
			expectedPrefix = fmt.Sprintf("db-replica-vol-%s", db.ID.String()[:8])
		}
		for _, v := range vols {
			if strings.HasPrefix(v.Name, expectedPrefix) {
				_ = s.volumeSvc.DeleteVolume(ctx, v.ID.String())
				break
			}
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_DELETE", id.String(), "DATABASE", nil); err != nil {
		s.logger.Warn("failed to record event", "action", "DATABASE_DELETE", "db_id", id, "error", err)
	}
	if err := s.auditSvc.Log(ctx, db.UserID, "database.delete", "database", db.ID.String(), map[string]interface{}{"name": db.Name}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "database.delete", "db_id", db.ID, "error", err)
	}
	platform.RDSInstancesTotal.WithLabelValues(string(db.Engine), "running").Dec()

	return nil
}

func (s *DatabaseService) PromoteToPrimary(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBUpdate, id.String()); err != nil {
		return err
	}

	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if db.Role == domain.RolePrimary {
		return errors.New(errors.InvalidInput, "database is already a primary")
	}
	db.Role = domain.RolePrimary
	db.PrimaryID = nil
	if err := s.repo.Update(ctx, db); err != nil {
		return err
	}
	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_PROMOTED", db.ID.String(), "DATABASE", nil); err != nil {
		s.logger.Warn("failed to record event", "action", "DATABASE_PROMOTED", "db_id", db.ID, "error", err)
	}
	return nil
}

func (s *DatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBRead, id.String()); err != nil {
		return "", err
	}

	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	port := db.Port
	if db.PoolingEnabled && db.PoolingPort != 0 {
		port = db.PoolingPort
	}
	switch db.Engine {
	case domain.EnginePostgres:
		return fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s", db.Username, db.Password, port, db.Name), nil
	case domain.EngineMySQL:
		return fmt.Sprintf("%s:%s@tcp(127.0.0.1:%d)/%s", db.Username, db.Password, port, db.Name), nil
	default:
		return "", errors.New(errors.Internal, "unknown engine")
	}
}

func (s *DatabaseService) CreateDatabaseSnapshot(ctx context.Context, databaseID uuid.UUID, description string) (*domain.Snapshot, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSnapshotCreate, "*"); err != nil {
		return nil, err
	}

	db, err := s.repo.GetByID(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	vol, err := s.getVolumeForDatabase(ctx, db)
	if err != nil {
		return nil, err
	}
	snapshotName := fmt.Sprintf("db-snap-%s-%s", db.Name, time.Now().Format("20060102150405"))
	snap, err := s.snapshotSvc.CreateSnapshot(ctx, vol.ID, snapshotName)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *DatabaseService) ListDatabaseSnapshots(ctx context.Context, databaseID uuid.UUID) ([]*domain.Snapshot, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSnapshotRead, "*"); err != nil {
		return nil, err
	}

	db, err := s.repo.GetByID(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	vol, err := s.getVolumeForDatabase(ctx, db)
	if err != nil {
		return nil, err
	}
	return s.snapshotRepo.ListByVolumeID(ctx, vol.ID)
}

func (s *DatabaseService) RestoreDatabase(ctx context.Context, req ports.RestoreDatabaseRequest) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionDBCreate, "*"); err != nil {
		return nil, err
	}

	snap, err := s.snapshotSvc.GetSnapshot(ctx, req.SnapshotID)
	if err != nil {
		return nil, err
	}
	dbEngine := domain.DatabaseEngine(req.Engine)
	password, _ := util.GenerateRandomPassword(16)
	username := s.getDefaultUsername(dbEngine)
	db := s.initialDatabaseRecord(userID, req.NewName, dbEngine, req.Version, username, password, req.VpcID)
	db.TenantID = tenantID

	// Preserving user requested storage if larger than snapshot
	db.AllocatedStorage = req.AllocatedStorage
	if snap.SizeGB > db.AllocatedStorage {
		db.AllocatedStorage = snap.SizeGB
	}

	db.MetricsEnabled = req.MetricsEnabled
	db.PoolingEnabled = req.PoolingEnabled

	vol, err := s.snapshotSvc.RestoreSnapshot(ctx, req.SnapshotID, fmt.Sprintf("db-vol-%s", db.ID.String()[:8]))
	if err != nil {
		return nil, err
	}

	// Re-verify volume size matches DB record (snapshot restore might resize)
	if vol.SizeGB > db.AllocatedStorage {
		db.AllocatedStorage = vol.SizeGB
	}

	networkID, _ := s.resolveVpcNetwork(ctx, req.VpcID)
	imageName, env, defaultPort := s.getEngineConfig(dbEngine, req.Version, username, password, req.NewName, domain.RolePrimary, "")

	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        fmt.Sprintf("cloud-db-%s-%s", req.NewName, db.ID.String()[:8]),
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: []string{fmt.Sprintf("%s:%s", s.getBackendVolName(vol), s.getMountPath(dbEngine))},
		Env:         env,
	})
	if err != nil {
		s.cleanupVolumeQuietly(ctx, vol.ID.String())
		return nil, errors.Wrap(errors.Internal, "failed to launch database container", err)
	}

	db.ContainerID = containerID
	if err := s.resolveDatabasePort(ctx, db, allocatedPorts, defaultPort); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to resolve restored database port", err))
	}
	db.Status = domain.DatabaseStatusRunning

	if db.MetricsEnabled || db.PoolingEnabled {
		dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
		if err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), errors.Wrap(errors.Internal, "failed to get restored database IP", err))
		}
		if err := s.provisionSidecars(ctx, db, dbEngine, dbIP, username, password, networkID); err != nil {
			return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
		}
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return s.performProvisioningRollback(ctx, db, vol.ID.String(), err)
	}
	return db, nil
}

// Internal helper methods

func (s *DatabaseService) resolveDatabasePort(ctx context.Context, db *domain.Database, allocatedPorts []string, defaultPort string) error {
	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, db.ContainerID, defaultPort)
		if err != nil {
			return err
		}
	}
	db.Port = hostPort
	return nil
}

func (s *DatabaseService) provisionSidecars(ctx context.Context, db *domain.Database, engine domain.DatabaseEngine, dbIP, username, password, networkID string) error {
	if db.MetricsEnabled {
		if err := s.provisionMetricsSidecar(ctx, db, engine, dbIP, username, password, networkID); err != nil {
			return err
		}
	}
	if db.PoolingEnabled {
		if err := s.provisionPoolerSidecar(ctx, db, engine, dbIP, username, password, networkID); err != nil {
			return err
		}
	}
	return nil
}

func (s *DatabaseService) provisionMetricsSidecar(ctx context.Context, db *domain.Database, engine domain.DatabaseEngine, dbIP, username, password, networkID string) error {
	image, env, internalPort := s.getExporterConfig(engine, dbIP, username, password, db.Name)
	cid, ports, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:      fmt.Sprintf("cloud-db-exporter-%s-%s", db.Name, db.ID.String()[:8]),
		ImageName: image,
		Ports:     []string{"0:" + internalPort},
		NetworkID: networkID,
		Env:       env,
	})
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to launch metrics exporter", err)
	}
	db.ExporterContainerID = cid
	hostPort, err := s.parseAllocatedPort(ports, internalPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, cid, internalPort)
		if err != nil {
			// Cleanup the sidecar on port resolution failure
			_ = s.compute.DeleteInstance(ctx, cid)
			db.ExporterContainerID = ""
			return errors.Wrap(errors.Internal, "failed to resolve metrics exporter port", err)
		}
	}
	db.MetricsPort = hostPort
	return nil
}

func (s *DatabaseService) provisionPoolerSidecar(ctx context.Context, db *domain.Database, engine domain.DatabaseEngine, dbIP, username, password, networkID string) error {
	image, env, internalPort := s.getPoolerConfig(engine, dbIP, username, password, db.Name)
	cid, ports, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:      fmt.Sprintf("cloud-db-pooler-%s-%s", db.Name, db.ID.String()[:8]),
		ImageName: image,
		Ports:     []string{"0:" + internalPort},
		NetworkID: networkID,
		Env:       env,
	})
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to launch connection pooler", err)
	}
	db.PoolerContainerID = cid
	hostPort, err := s.parseAllocatedPort(ports, internalPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, cid, internalPort)
		if err != nil {
			// Cleanup the sidecar on port resolution failure
			_ = s.compute.DeleteInstance(ctx, cid)
			db.PoolerContainerID = ""
			return errors.Wrap(errors.Internal, "failed to resolve connection pooler port", err)
		}
	}
	db.PoolingPort = hostPort
	return nil
}

func (s *DatabaseService) performProvisioningRollback(ctx context.Context, db *domain.Database, volID string, err error) (*domain.Database, error) {
	s.logger.Error("rolling back database provisioning due to failure", "error", err)
	if db.ContainerID != "" {
		if deleteErr := s.compute.DeleteInstance(ctx, db.ContainerID); deleteErr != nil {
			s.logger.Warn("failed to delete database container during rollback", "container_id", db.ContainerID, "error", deleteErr)
		}
	}
	if db.ExporterContainerID != "" {
		if deleteErr := s.compute.DeleteInstance(ctx, db.ExporterContainerID); deleteErr != nil {
			s.logger.Warn("failed to delete exporter container during rollback", "container_id", db.ExporterContainerID, "error", deleteErr)
		}
	}
	if db.PoolerContainerID != "" {
		if deleteErr := s.compute.DeleteInstance(ctx, db.PoolerContainerID); deleteErr != nil {
			s.logger.Warn("failed to delete pooler container during rollback", "container_id", db.PoolerContainerID, "error", deleteErr)
		}
	}
	if deleteErr := s.volumeSvc.DeleteVolume(ctx, volID); deleteErr != nil {
		s.logger.Warn("failed to delete volume during rollback", "volume_id", volID, "error", deleteErr)
	}
	return nil, err
}

func (s *DatabaseService) cleanupVolumeQuietly(ctx context.Context, volID string) {
	if err := s.volumeSvc.DeleteVolume(ctx, volID); err != nil {
		s.logger.Warn("failed to delete volume during cleanup", "volume_id", volID, "error", err)
	}
}

func (s *DatabaseService) getBackendVolName(vol *domain.Volume) string {
	if vol.BackendPath != "" {
		return vol.BackendPath
	}
	return "thecloud-vol-" + vol.ID.String()[:8]
}

func (s *DatabaseService) getMountPath(engine domain.DatabaseEngine) string {
	if engine == domain.EngineMySQL {
		return "/var/lib/mysql"
	}
	return "/var/lib/postgresql/data"
}

func (s *DatabaseService) isValidEngine(engine domain.DatabaseEngine) bool {
	return engine == domain.EnginePostgres || engine == domain.EngineMySQL
}

func (s *DatabaseService) getDefaultUsername(engine domain.DatabaseEngine) string {
	if engine == domain.EngineMySQL {
		return "root"
	}
	return "cloud_user"
}

func (s *DatabaseService) getExporterConfig(engine domain.DatabaseEngine, dbIP, username, password, dbName string) (string, []string, string) {
	switch engine {
	case domain.EnginePostgres:
		dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", username, password, dbIP, DefaultPostgresPort, dbName)
		return PostgresExporterImage, []string{"DATA_SOURCE_NAME=" + dsn}, PostgresExporterPort
	case domain.EngineMySQL:
		dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s", username, password, dbIP, DefaultMySQLPort, dbName)
		return MySQLExporterImage, []string{"DATA_SOURCE_NAME=" + dsn}, MySQLExporterPort
	}
	return "", nil, ""
}

func (s *DatabaseService) getPoolerConfig(engine domain.DatabaseEngine, dbIP, username, password, dbName string) (string, []string, string) {
	if engine == domain.EnginePostgres {
		env := []string{
			"DB_HOST=" + dbIP,
			"DB_PORT=" + DefaultPostgresPort,
			"DB_USER=" + username,
			"DB_PASSWORD=" + password,
			"DB_NAME=" + dbName,
			"POOL_MODE=" + DefaultPoolMode,
			"MAX_CLIENT_CONN=" + DefaultMaxClientConn,
			"DEFAULT_POOL_SIZE=" + DefaultPoolSize,
		}
		return PoolerImage, env, PoolerInternalPort
	}
	return "", nil, ""
}

func (s *DatabaseService) buildEngineCmd(engine domain.DatabaseEngine, parameters map[string]string) []string {
	if len(parameters) == 0 {
		return nil
	}
	var cmd []string
	switch engine {
	case domain.EnginePostgres:
		cmd = append(cmd, "postgres")
		for k, v := range parameters {
			cmd = append(cmd, "-c", fmt.Sprintf("%s=%s", k, v))
		}
	case domain.EngineMySQL:
		cmd = append(cmd, "mysqld")
		for k, v := range parameters {
			cmd = append(cmd, fmt.Sprintf("--%s=%s", k, v))
		}
	}
	return cmd
}

func (s *DatabaseService) getEngineConfig(engine domain.DatabaseEngine, version, username, password, name string, role domain.DatabaseRole, primaryIP string) (string, []string, string) {
	switch engine {
	case domain.EnginePostgres:
		env := []string{"POSTGRES_USER=" + username, "POSTGRES_PASSWORD=" + password, "POSTGRES_DB=" + name}
		if role == domain.RoleReplica {
			env = append(env, "PRIMARY_HOST="+primaryIP)
		}
		return fmt.Sprintf("postgres:%s-alpine", version), env, DefaultPostgresPort
	case domain.EngineMySQL:
		env := []string{"MYSQL_ROOT_PASSWORD=" + password, "MYSQL_DATABASE=" + name}
		if role == domain.RoleReplica {
			env = append(env, "MYSQL_MASTER_HOST="+primaryIP)
		}
		return fmt.Sprintf("mysql:%s", version), env, DefaultMySQLPort
	}
	return "", nil, ""
}

func (s *DatabaseService) resolveVpcNetwork(ctx context.Context, vpcID *uuid.UUID) (string, error) {
	if vpcID == nil {
		return "", nil
	}
	vpc, err := s.vpcRepo.GetByID(ctx, *vpcID)
	if err != nil {
		return "", err
	}
	// When networking falls back to no-op, VPCs still carry synthetic OVS bridge
	// names like br-vpc-*. Docker cannot join those, so use the default network.
	if s.compute != nil && s.compute.Type() == "docker" && strings.HasPrefix(vpc.NetworkID, "br-vpc-") {
		return "", nil
	}
	return vpc.NetworkID, nil
}

func (s *DatabaseService) initialDatabaseRecord(userID uuid.UUID, name string, engine domain.DatabaseEngine, version, username, password string, vpcID *uuid.UUID) *domain.Database {
	return &domain.Database{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Engine:    engine,
		Version:   version,
		Status:    domain.DatabaseStatusCreating,
		VpcID:     vpcID,
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *DatabaseService) recordDatabaseCreation(ctx context.Context, userID uuid.UUID, db *domain.Database, originalEngine string) {
	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{"name": db.Name, "engine": db.Engine}); err != nil {
		s.logger.Warn("failed to record event", "action", "DATABASE_CREATE", "db_id", db.ID, "error", err)
	}
	if err := s.auditSvc.Log(ctx, userID, "database.create", "database", db.ID.String(), map[string]interface{}{"name": db.Name, "engine": originalEngine}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "database.create", "db_id", db.ID, "error", err)
	}
	platform.RDSInstancesTotal.WithLabelValues(originalEngine, "running").Inc()
}

func (s *DatabaseService) getVolumeForDatabase(ctx context.Context, db *domain.Database) (*domain.Volume, error) {
	vols, err := s.volumeSvc.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}
	expectedPrefix := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
	if db.Role == domain.RoleReplica {
		expectedPrefix = fmt.Sprintf("db-replica-vol-%s", db.ID.String()[:8])
	}
	for _, v := range vols {
		if strings.HasPrefix(v.Name, expectedPrefix) {
			return v, nil
		}
	}
	return nil, errors.New(errors.NotFound, "volume not found")
}

func (s *DatabaseService) parseAllocatedPort(allocatedPorts []string, targetPort string) (int, error) {
	for _, p := range allocatedPorts {
		parts := strings.Split(p, ":")
		if len(parts) == 2 && parts[1] == targetPort {
			hp, err := strconv.Atoi(parts[0])
			if err != nil {
				return 0, err
			}
			return hp, nil
		}
	}
	return 0, nil
}
