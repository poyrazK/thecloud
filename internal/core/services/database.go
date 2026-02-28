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

// DatabaseService manages database instances and lifecycle.
type DatabaseService struct {
	repo      ports.DatabaseRepository
	compute   ports.ComputeBackend
	vpcRepo   ports.VpcRepository
	volumeSvc ports.VolumeService
	eventSvc  ports.EventService
	auditSvc  ports.AuditService
	logger    *slog.Logger
}

// DatabaseServiceParams holds dependencies for DatabaseService creation.
type DatabaseServiceParams struct {
	Repo      ports.DatabaseRepository
	Compute   ports.ComputeBackend
	VpcRepo   ports.VpcRepository
	VolumeSvc ports.VolumeService
	EventSvc  ports.EventService
	AuditSvc  ports.AuditService
	Logger    *slog.Logger
}

// NewDatabaseService constructs a DatabaseService with its dependencies.
func NewDatabaseService(params DatabaseServiceParams) *DatabaseService {
	return &DatabaseService{
		repo:      params.Repo,
		compute:   params.Compute,
		vpcRepo:   params.VpcRepo,
		volumeSvc: params.VolumeSvc,
		eventSvc:  params.EventSvc,
		auditSvc:  params.AuditSvc,
		logger:    params.Logger,
	}
}

func (s *DatabaseService) CreateDatabase(ctx context.Context, name, engine, version string, vpcID *uuid.UUID) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)

	dbEngine := domain.DatabaseEngine(engine)
	if !s.isValidEngine(dbEngine) {
		return nil, errors.New(errors.InvalidInput, "unsupported database engine")
	}

	password, err := util.GenerateRandomPassword(16)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate password", err)
	}

	username := s.getDefaultUsername(dbEngine)
	db := s.initialDatabaseRecord(userID, name, dbEngine, version, username, password, vpcID)
	db.Role = domain.RolePrimary

	imageName, env, defaultPort := s.getEngineConfig(dbEngine, version, username, password, name, db.Role, "")

	networkID, err := s.resolveVpcNetwork(ctx, vpcID)
	if err != nil {
		return nil, err
	}

	// Create persistent volume for the database
	volName := fmt.Sprintf("db-vol-%s", db.ID.String()[:8])
	vol, err := s.volumeSvc.CreateVolume(ctx, volName, 10) // 10GB default
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

	dockerName := fmt.Sprintf("cloud-db-%s-%s", name, db.ID.String()[:8])
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
		s.logger.Error("failed to launch database container", "error", err)
		return nil, errors.Wrap(errors.Internal, "failed to launch database container", err)
	}

	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, err = s.compute.GetInstancePort(ctx, containerID, defaultPort)
		if err != nil {
			s.logger.Error("failed to resolve database port", "container_id", containerID, "error", err)
			if delErr := s.compute.DeleteInstance(ctx, containerID); delErr != nil {
				s.logger.Error("failed to clean up database container after port resolution failure", "container_id", containerID, "error", delErr)
			}
			_ = s.volumeSvc.DeleteVolume(ctx, vol.ID.String())
			return nil, errors.Wrap(errors.Internal, "failed to resolve database port", err)
		}
	}

	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	if err := s.repo.Create(ctx, db); err != nil {
		if delErr := s.compute.DeleteInstance(ctx, containerID); delErr != nil {
			s.logger.Error("failed to clean up database container after repo create failure", "container_id", containerID, "error", delErr)
		}
		_ = s.volumeSvc.DeleteVolume(ctx, vol.ID.String())
		return nil, err
	}

	s.recordDatabaseCreation(ctx, userID, db, engine)
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

	imageName, env, defaultPort := s.getEngineConfig(primary.Engine, primary.Version, primary.Username, primary.Password, name, db.Role, primaryIP)

	networkID, err := s.resolveVpcNetwork(ctx, db.VpcID)
	if err != nil {
		return nil, err
	}

	// Create persistent volume for the replica
	volName := fmt.Sprintf("db-replica-vol-%s", db.ID.String()[:8])
	vol, err := s.volumeSvc.CreateVolume(ctx, volName, 10) // 10GB default
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

	hostPort, err := s.parseAllocatedPort(allocatedPorts, defaultPort)
	if err != nil || hostPort == 0 {
		hostPort, _ = s.compute.GetInstancePort(ctx, containerID, defaultPort)
	}

	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	if err := s.repo.Create(ctx, db); err != nil {
		_ = s.compute.DeleteInstance(ctx, containerID)
		_ = s.volumeSvc.DeleteVolume(ctx, vol.ID.String())
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_REPLICA_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{
		"primary_id": primaryID,
		"name":       name,
	})

	return db, nil
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
		return fmt.Sprintf("postgres:%s-alpine", version), env, "5432"
	case domain.EngineMySQL:
		env := []string{
			"MYSQL_ROOT_PASSWORD=" + password,
			"MYSQL_DATABASE=" + name,
		}
		if role == domain.RoleReplica {
			env = append(env, "MYSQL_MASTER_HOST="+primaryIP)
		}
		return fmt.Sprintf("mysql:%s", version), env, "3306"
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

	// 1. Remove container
	if db.ContainerID != "" {
		if err := s.compute.DeleteInstance(ctx, db.ContainerID); err != nil {
			s.logger.Warn("failed to remove database container", "container_id", db.ContainerID, "error", err)
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

func (s *DatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
	db, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	// In a real cloud, this would be the LB endpoint or internal VPC IP.
	// For this simulation, we'll return a localhost connection string with the mapped port.
	// Note: We need to get the mapped port from Docker if we don't store it accurately.
	// For now, let's assume we store it in db.Port (I should probably implement a way to retrieve it).

	switch db.Engine {
	case domain.EnginePostgres:
		return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", db.Username, db.Password, db.Port, db.Name), nil
	case domain.EngineMySQL:
		return fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s", db.Username, db.Password, db.Port, db.Name), nil
	default:
		return "", errors.New(errors.Internal, "unknown engine")
	}
}
