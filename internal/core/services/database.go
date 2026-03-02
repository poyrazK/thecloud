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

	dbEngine := domain.DatabaseEngine(req.Engine)
	if !s.isValidEngine(dbEngine) {
		return nil, errors.New(errors.InvalidInput, "unsupported database engine")
	}

	if req.AllocatedStorage < 10 {
		return nil, errors.New(errors.InvalidInput, "allocated storage must be at least 10GB")
	}

	// Verify pooling support (PgBouncer only for Postgres currently)
	if req.PoolingEnabled && dbEngine != domain.EnginePostgres {
		return nil, errors.New(errors.InvalidInput, "connection pooling is currently only supported for PostgreSQL")
	}

	password, err := util.GenerateRandomPassword(16)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate password", err)
	}

	username := s.getDefaultUsername(dbEngine)
	db := s.initialDatabaseRecord(userID, req.Name, dbEngine, req.Version, username, password, req.VpcID)
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

	// Create persistent volume for the database
	volName := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
	vol, err := s.volumeSvc.CreateVolume(ctx, volName, req.AllocatedStorage)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create persistent volume", err)
	}

	backendVolName := "thecloud-vol-" + vol.ID.String()[:8]
	if vol.BackendPath != "" {
		backendVolName = vol.BackendPath
	}

	mountPath := "/var/lib/postgresql/data"
	if dbEngine == domain.EngineMySQL {
		mountPath = "/var/lib/mysql"
	}
	volumeBinds := []string{fmt.Sprintf("%s:%s", backendVolName, mountPath)}

	cmd := s.buildEngineCmd(dbEngine, req.Parameters)

	dockerName := fmt.Sprintf("cloud-db-%s-%s", req.Name, db.ID.String()[:8])
	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: volumeBinds,
		Env:         env,
		Cmd:         cmd,
	})

	if err != nil {
		if delErr := s.volumeSvc.DeleteVolume(ctx, vol.ID.String()); delErr != nil {
			s.logger.Error("failed to delete volume after container launch failure", "volume_id", vol.ID, "error", delErr)
		}
		s.logger.Error("failed to launch database container", "error", err)
		return nil, errors.Wrap(errors.Internal, "failed to launch database container", err)
	}

	// Deterministic cleanup helper
	rollback := func(err error) (*domain.Database, error) {
		s.logger.Error("rolling back database provisioning due to failure", "error", err)
		if delErr := s.compute.DeleteInstance(ctx, containerID); delErr != nil {
			s.logger.Error("failed to clean up database container during rollback", "container_id", containerID, "error", delErr)
		}
		if db.ExporterContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.ExporterContainerID); delErr != nil {
				s.logger.Error("failed to clean up metrics exporter container during rollback", "container_id", db.ExporterContainerID, "error", delErr)
			}
		}
		if db.PoolerContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.PoolerContainerID); delErr != nil {
				s.logger.Error("failed to clean up pooler container during rollback", "container_id", db.PoolerContainerID, "error", delErr)
			}
		}
		if delErr := s.volumeSvc.DeleteVolume(ctx, vol.ID.String()); delErr != nil {
			s.logger.Error("failed to delete volume during rollback", "volume_id", vol.ID, "error", delErr)
		}
		return nil, err
	}

	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, containerID, defaultPort)
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to resolve database port", err))
		}
	}

	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	// Sidecars require the DB IP
	dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
	if err != nil {
		return rollback(errors.Wrap(errors.Internal, "failed to get database IP for sidecars", err))
	}

	// Launch metrics sidecar if enabled
	if req.MetricsEnabled {
		exporterImage, exporterEnv, exporterPort := s.getExporterConfig(dbEngine, dbIP, username, password, req.Name)
		exporterName := fmt.Sprintf("cloud-db-exporter-%s-%s", req.Name, db.ID.String()[:8])

		expCID, expPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      exporterName,
			ImageName: exporterImage,
			Ports:     []string{"0:" + exporterPort},
			NetworkID: networkID,
			Env:       exporterEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch metrics exporter", err))
		}

		db.ExporterContainerID = expCID
		expHostPort, err := s.parseAllocatedPort(expPorts, exporterPort)
		if err != nil || expHostPort == 0 {
			expHostPort, err = s.compute.GetInstancePort(ctx, expCID, exporterPort)
			if err != nil {
				return rollback(errors.Wrap(errors.Internal, "failed to resolve metrics exporter port", err))
			}
		}
		db.MetricsPort = expHostPort
	}

	// Launch pooler sidecar if enabled
	if req.PoolingEnabled {
		poolerImage, poolerEnv, poolerPort := s.getPoolerConfig(dbEngine, dbIP, username, password, req.Name)
		poolerName := fmt.Sprintf("cloud-db-pooler-%s-%s", req.Name, db.ID.String()[:8])

		pCID, pPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      poolerName,
			ImageName: poolerImage,
			Ports:     []string{"0:" + poolerPort},
			NetworkID: networkID,
			Env:       poolerEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch connection pooler", err))
		}

		db.PoolerContainerID = pCID
		pHostPort, err := s.parseAllocatedPort(pPorts, poolerPort)
		if err != nil || pHostPort == 0 {
			pHostPort, err = s.compute.GetInstancePort(ctx, pCID, poolerPort)
			if err != nil {
				return rollback(errors.Wrap(errors.Internal, "failed to resolve connection pooler port", err))
			}
		}
		db.PoolingPort = pHostPort
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return rollback(err)
	}

	s.recordDatabaseCreation(ctx, userID, db, req.Engine)
	return db, nil
}

func (s *DatabaseService) CreateReplica(ctx context.Context, primaryID uuid.UUID, name string) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)

	primary, err := s.repo.GetByID(ctx, primaryID)
	if err != nil {
		return nil, err
	}

	primaryIP, err := s.compute.GetInstanceIP(ctx, primary.ContainerID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get primary IP", err)
	}

	db := s.initialDatabaseRecord(userID, name, primary.Engine, primary.Version, primary.Username, primary.Password, primary.VpcID)
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

	backendVolName := "thecloud-vol-" + vol.ID.String()[:8]
	if vol.BackendPath != "" {
		backendVolName = vol.BackendPath
	}

	mountPath := "/var/lib/postgresql/data"
	if db.Engine == domain.EngineMySQL {
		mountPath = "/var/lib/mysql"
	}
	volumeBinds := []string{fmt.Sprintf("%s:%s", backendVolName, mountPath)}

	dockerName := fmt.Sprintf("cloud-db-replica-%s-%s", name, db.ID.String()[:8])
	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: volumeBinds,
		Env:         env,
		Cmd:         nil,
	})
	if err != nil {
		_ = s.volumeSvc.DeleteVolume(ctx, vol.ID.String())
		return nil, errors.Wrap(errors.Internal, "failed to launch replica container", err)
	}

	// Deterministic cleanup helper for replica
	rollback := func(err error) (*domain.Database, error) {
		s.logger.Error("rolling back database replica provisioning due to failure", "error", err)
		if delErr := s.compute.DeleteInstance(ctx, containerID); delErr != nil {
			s.logger.Error("failed to clean up replica container during rollback", "container_id", containerID, "error", delErr)
		}
		if db.ExporterContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.ExporterContainerID); delErr != nil {
				s.logger.Error("failed to clean up replica metrics exporter container during rollback", "container_id", db.ExporterContainerID, "error", delErr)
			}
		}
		if db.PoolerContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.PoolerContainerID); delErr != nil {
				s.logger.Error("failed to clean up replica pooler container during rollback", "container_id", db.PoolerContainerID, "error", delErr)
			}
		}
		if delErr := s.volumeSvc.DeleteVolume(ctx, vol.ID.String()); delErr != nil {
			s.logger.Error("failed to delete replica volume during rollback", "volume_id", vol.ID, "error", delErr)
		}
		return nil, err
	}

	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, _ = s.compute.GetInstancePort(ctx, containerID, defaultPort)
	}

	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
	if err != nil {
		return rollback(errors.Wrap(errors.Internal, "failed to get replica IP for sidecars", err))
	}

	// Launch metrics sidecar for replica if enabled on primary
	if db.MetricsEnabled {
		exporterImage, exporterEnv, exporterPort := s.getExporterConfig(db.Engine, dbIP, db.Username, db.Password, name)
		exporterName := fmt.Sprintf("cloud-db-replica-exporter-%s-%s", name, db.ID.String()[:8])

		expCID, expPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      exporterName,
			ImageName: exporterImage,
			Ports:     []string{"0:" + exporterPort},
			NetworkID: networkID,
			Env:       exporterEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch replica metrics exporter", err))
		}

		db.ExporterContainerID = expCID
		expHostPort, err := s.parseAllocatedPort(expPorts, exporterPort)
		if err != nil || expHostPort == 0 {
			expHostPort, _ = s.compute.GetInstancePort(ctx, expCID, exporterPort)
		}
		db.MetricsPort = expHostPort
	}

	// Launch pooler sidecar for replica if enabled
	if db.PoolingEnabled {
		poolerImage, poolerEnv, poolerPort := s.getPoolerConfig(db.Engine, dbIP, db.Username, db.Password, name)
		poolerName := fmt.Sprintf("cloud-db-replica-pooler-%s-%s", name, db.ID.String()[:8])

		pCID, pPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      poolerName,
			ImageName: poolerImage,
			Ports:     []string{"0:" + poolerPort},
			NetworkID: networkID,
			Env:       poolerEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch replica connection pooler", err))
		}

		db.PoolerContainerID = pCID
		pHostPort, err := s.parseAllocatedPort(pPorts, poolerPort)
		if err != nil || pHostPort == 0 {
			pHostPort, _ = s.compute.GetInstancePort(ctx, pCID, poolerPort)
		}
		db.PoolingPort = pHostPort
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return rollback(err)
	}

	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_REPLICA_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{
		"primary_id": primaryID,
		"name":       name,
	})

	return db, nil
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

func (s *DatabaseService) PromoteToPrimary(ctx context.Context, id uuid.UUID) error {
	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if db.Role == domain.RolePrimary {
		return errors.New(errors.InvalidInput, "database is already a primary")
	}

	// 1. Update metadata
	db.Role = domain.RolePrimary
	db.PrimaryID = nil

	// 2. Potentially update container (complex in Docker without restart)
	// For now, we'll assume the application logic handles the role change
	// or we'd need to re-launch/signal the container.
	// We'll record an event for manual/automated orchestration.

	if err := s.repo.Update(ctx, db); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_PROMOTED", db.ID.String(), "DATABASE", nil)

	_ = s.auditSvc.Log(ctx, db.UserID, "database.promote", "database", db.ID.String(), map[string]interface{}{
		"name": db.Name,
	})

	return nil
}

func (s *DatabaseService) parseAllocatedPort(allocatedPorts []string, targetPort string) (int, error) {
	for _, p := range allocatedPorts {
		parts := strings.Split(p, ":")
		// handle host:port or host:containerPort:hostPort or containerPort:hostPort
		if len(parts) >= 2 && parts[len(parts)-1] == targetPort {
			portStr := parts[0]
			if len(parts) == 3 {
				portStr = parts[1]
			}
			hp, err := strconv.Atoi(portStr)
			if err != nil {
				return 0, err
			}
			return hp, nil
		}
	}
	return 0, nil
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
		// postgres-exporter uses DATA_SOURCE_NAME
		dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", username, password, dbIP, DefaultPostgresPort, dbName)
		return PostgresExporterImage, []string{"DATA_SOURCE_NAME=" + dsn}, PostgresExporterPort
	case domain.EngineMySQL:
		// mysqld-exporter uses DATA_SOURCE_NAME
		dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s", username, password, dbIP, DefaultMySQLPort, dbName)
		return MySQLExporterImage, []string{"DATA_SOURCE_NAME=" + dsn}, MySQLExporterPort
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
			// postgres uses -c key=value
			cmd = append(cmd, "-c", fmt.Sprintf("%s=%s", k, v))
		}
	case domain.EngineMySQL:
		cmd = append(cmd, "mysqld")
		for k, v := range parameters {
			// mysql uses --key=value
			cmd = append(cmd, fmt.Sprintf("--%s=%s", k, v))
		}
	}

	return cmd
}

func (s *DatabaseService) getEngineConfig(engine domain.DatabaseEngine, version, username, password, name string, role domain.DatabaseRole, primaryIP string) (string, []string, string) {
	switch engine {
	case domain.EnginePostgres:
		env := []string{
			"POSTGRES_USER=" + username,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + name,
		}
		// Using official postgres image logic for simplicity in simulation.
		// In a real scenario, we might use bitnami/postgresql which has better built-in replication env vars.
		if role == domain.RoleReplica {
			// Simulating replication setup
			env = append(env, "PRIMARY_HOST="+primaryIP)
		}
		return fmt.Sprintf("postgres:%s-alpine", version), env, DefaultPostgresPort
	case domain.EngineMySQL:
		env := []string{
			"MYSQL_ROOT_PASSWORD=" + password,
			"MYSQL_DATABASE=" + name,
		}
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
		return "", errors.Wrap(errors.NotFound, "vpc not found", err)
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
	if err := s.eventSvc.RecordEvent(ctx, "DATABASE_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{
		"name":   db.Name,
		"engine": db.Engine,
	}); err != nil {
		s.logger.Warn("failed to record database creation event", "id", db.ID, "error", err)
	}

	if err := s.auditSvc.Log(ctx, userID, "database.create", "database", db.ID.String(), map[string]interface{}{
		"name":   db.Name,
		"engine": originalEngine,
	}); err != nil {
		s.logger.Warn("failed to log database creation to audit", "id", db.ID, "error", err)
	}

	platform.RDSInstancesTotal.WithLabelValues(originalEngine, "running").Inc()
}

func (s *DatabaseService) GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DatabaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	return s.repo.List(ctx)
}

func (s *DatabaseService) DeleteDatabase(ctx context.Context, id uuid.UUID) error {
	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. Remove containers
	if db.ContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.ContainerID); err != nil {
			s.logger.Warn("failed to remove database container", "container_id", db.ContainerID, "error", err)
		}
	}
	if db.ExporterContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.ExporterContainerID); err != nil {
			s.logger.Warn("failed to remove metrics exporter container", "container_id", db.ExporterContainerID, "error", err)
		}
	}
	if db.PoolerContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.PoolerContainerID); err != nil {
			s.logger.Warn("failed to remove connection pooler container", "container_id", db.PoolerContainerID, "error", err)
		}
	}

	// 2. Clean up volume
	// We try to find the volume by the expected name
	vols, err := s.volumeSvc.ListVolumes(ctx)
	if err == nil {
		expectedPrefix := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
		if db.Role == domain.RoleReplica {
			expectedPrefix = fmt.Sprintf("db-replica-vol-%s", db.ID.String()[:8])
		}
		for _, v := range vols {
			if strings.HasPrefix(v.Name, expectedPrefix) {
				if err := s.volumeSvc.DeleteVolume(ctx, v.ID.String()); err != nil {
					s.logger.Warn("failed to delete database volume", "volume_id", v.ID, "error", err)
				}
				break
			}
		}
	}

	// 3. Delete from repo
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_DELETE", id.String(), "DATABASE", nil)

	_ = s.auditSvc.Log(ctx, db.UserID, "database.delete", "database", db.ID.String(), map[string]interface{}{
		"name": db.Name,
	})

	platform.RDSInstancesTotal.WithLabelValues(string(db.Engine), "running").Dec()

	return nil
}

func (s *DatabaseService) getVolumeForDatabase(ctx context.Context, db *domain.Database) (*domain.Volume, error) {
	vols, err := s.volumeSvc.ListVolumes(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list volumes", err)
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

	return nil, errors.New(errors.NotFound, "underlying database volume not found")
}

func (s *DatabaseService) CreateDatabaseSnapshot(ctx context.Context, databaseID uuid.UUID, description string) (*domain.Snapshot, error) {
	db, err := s.repo.GetByID(ctx, databaseID)
	if err != nil {
		return nil, err
	}

	if db.Status != domain.DatabaseStatusRunning && db.Status != domain.DatabaseStatusStopped {
		return nil, errors.New(errors.InvalidInput, "database must be running or stopped to create a snapshot")
	}

	vol, err := s.getVolumeForDatabase(ctx, db)
	if err != nil {
		return nil, err
	}

	snapshotName := fmt.Sprintf("db-snap-%s-%s", db.Name, time.Now().Format("20060102150405"))
	if description != "" {
		snapshotName = fmt.Sprintf("%s-%s", snapshotName, description)
	}

	snap, err := s.snapshotSvc.CreateSnapshot(ctx, vol.ID, snapshotName)
	if err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, db.UserID, "database.snapshot.create", "database", db.ID.String(), map[string]interface{}{
		"snapshot_id": snap.ID,
	})

	return snap, nil
}

func (s *DatabaseService) ListDatabaseSnapshots(ctx context.Context, databaseID uuid.UUID) ([]*domain.Snapshot, error) {
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

	snap, err := s.snapshotSvc.GetSnapshot(ctx, req.SnapshotID)
	if err != nil {
		return nil, err
	}

	if req.AllocatedStorage < snap.SizeGB {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("allocated storage (%dGB) must be at least the snapshot size (%dGB)", req.AllocatedStorage, snap.SizeGB))
	}

	dbEngine := domain.DatabaseEngine(req.Engine)
	if !s.isValidEngine(dbEngine) {
		return nil, errors.New(errors.InvalidInput, "unsupported database engine")
	}

	if req.PoolingEnabled && dbEngine != domain.EnginePostgres {
		return nil, errors.New(errors.InvalidInput, "connection pooling is currently only supported for PostgreSQL")
	}

	password, err := util.GenerateRandomPassword(16)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate password", err)
	}

	username := s.getDefaultUsername(dbEngine)
	db := s.initialDatabaseRecord(userID, req.NewName, dbEngine, req.Version, username, password, req.VpcID)
	db.Role = domain.RolePrimary
	db.AllocatedStorage = req.AllocatedStorage
	db.Parameters = req.Parameters
	db.MetricsEnabled = req.MetricsEnabled
	db.PoolingEnabled = req.PoolingEnabled

	imageName, env, defaultPort := s.getEngineConfig(dbEngine, req.Version, username, password, req.NewName, db.Role, "")

	networkID, err := s.resolveVpcNetwork(ctx, req.VpcID)
	if err != nil {
		return nil, err
	}

	// Create persistent volume FOR THE NEW DATABASE from the snapshot
	volName := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
	vol, err := s.snapshotSvc.RestoreSnapshot(ctx, req.SnapshotID, volName)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to restore volume from snapshot", err)
	}

	// Resizing is handled by volumeSvc.
	db.AllocatedStorage = vol.SizeGB // Ensure DB record matches actual volume

	backendVolName := "thecloud-vol-" + vol.ID.String()[:8]
	if vol.BackendPath != "" {
		backendVolName = vol.BackendPath
	}

	mountPath := "/var/lib/postgresql/data"
	if dbEngine == domain.EngineMySQL {
		mountPath = "/var/lib/mysql"
	}
	volumeBinds := []string{fmt.Sprintf("%s:%s", backendVolName, mountPath)}

	cmd := s.buildEngineCmd(dbEngine, req.Parameters)

	dockerName := fmt.Sprintf("cloud-db-%s-%s", req.NewName, db.ID.String()[:8])
	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: volumeBinds,
		Env:         env,
		Cmd:         cmd,
	})

	if err != nil {
		_ = s.volumeSvc.DeleteVolume(ctx, vol.ID.String())
		s.logger.Error("failed to launch restored database container", "error", err)
		return nil, errors.Wrap(errors.Internal, "failed to launch restored database container", err)
	}

	// Deterministic cleanup helper
	rollback := func(err error) (*domain.Database, error) {
		s.logger.Error("rolling back restored database provisioning due to failure", "error", err)
		if delErr := s.compute.DeleteInstance(ctx, containerID); delErr != nil {
			s.logger.Error("failed to clean up database container during rollback", "container_id", containerID, "error", delErr)
		}
		if db.ExporterContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.ExporterContainerID); delErr != nil {
				s.logger.Error("failed to clean up metrics exporter container during rollback", "container_id", db.ExporterContainerID, "error", delErr)
			}
		}
		if db.PoolerContainerID != "" {
			if delErr := s.compute.DeleteInstance(ctx, db.PoolerContainerID); delErr != nil {
				s.logger.Error("failed to clean up pooler container during rollback", "container_id", db.PoolerContainerID, "error", delErr)
			}
		}
		if delErr := s.volumeSvc.DeleteVolume(ctx, vol.ID.String()); delErr != nil {
			s.logger.Error("failed to delete volume during rollback", "volume_id", vol.ID, "error", delErr)
		}
		return nil, err
	}

	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, containerID, defaultPort)
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to resolve restored database port", err))
		}
	}

	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	dbIP, err := s.compute.GetInstanceIP(ctx, containerID)
	if err != nil {
		return rollback(errors.Wrap(errors.Internal, "failed to get restored database IP for sidecars", err))
	}

	// Launch metrics sidecar if enabled
	if req.MetricsEnabled {
		exporterImage, exporterEnv, exporterPort := s.getExporterConfig(dbEngine, dbIP, username, password, req.NewName)
		exporterName := fmt.Sprintf("cloud-db-exporter-%s-%s", req.NewName, db.ID.String()[:8])

		expCID, expPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      exporterName,
			ImageName: exporterImage,
			Ports:     []string{"0:" + exporterPort},
			NetworkID: networkID,
			Env:       exporterEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch restored metrics exporter", err))
		}

		db.ExporterContainerID = expCID
		expHostPort, err := s.parseAllocatedPort(expPorts, exporterPort)
		if err != nil || expHostPort == 0 {
			expHostPort, _ = s.compute.GetInstancePort(ctx, expCID, exporterPort)
		}
		db.MetricsPort = expHostPort
	}

	// Launch pooler sidecar if enabled
	if req.PoolingEnabled {
		poolerImage, poolerEnv, poolerPort := s.getPoolerConfig(dbEngine, dbIP, username, password, req.NewName)
		poolerName := fmt.Sprintf("cloud-db-pooler-%s-%s", req.NewName, db.ID.String()[:8])

		pCID, pPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      poolerName,
			ImageName: poolerImage,
			Ports:     []string{"0:" + poolerPort},
			NetworkID: networkID,
			Env:       poolerEnv,
		})
		if err != nil {
			return rollback(errors.Wrap(errors.Internal, "failed to launch restored connection pooler", err))
		}

		db.PoolerContainerID = pCID
		pHostPort, err := s.parseAllocatedPort(pPorts, poolerPort)
		if err != nil || pHostPort == 0 {
			pHostPort, _ = s.compute.GetInstancePort(ctx, pCID, poolerPort)
		}
		db.PoolingPort = pHostPort
	}

	if err := s.repo.Create(ctx, db); err != nil {
		return rollback(err)
	}

	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_RESTORE", db.ID.String(), "DATABASE", map[string]interface{}{
		"snapshot_id": req.SnapshotID,
	})

	_ = s.auditSvc.Log(ctx, userID, "database.restore", "database", db.ID.String(), map[string]interface{}{
		"snapshot_id": req.SnapshotID,
	})

	platform.RDSInstancesTotal.WithLabelValues(req.Engine, "running").Inc()

	return db, nil
}
func (s *DatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
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
		return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", db.Username, db.Password, port, db.Name), nil
	case domain.EngineMySQL:
		return fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s", db.Username, db.Password, port, db.Name), nil
	default:
		return "", errors.New(errors.Internal, "unknown engine")
	}
}
