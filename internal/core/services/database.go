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
	repo     ports.DatabaseRepository
	compute  ports.ComputeBackend
	vpcRepo  ports.VpcRepository
	eventSvc ports.EventService
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// DatabaseServiceParams holds dependencies for DatabaseService creation.
type DatabaseServiceParams struct {
	Repo     ports.DatabaseRepository
	Compute  ports.ComputeBackend
	VpcRepo  ports.VpcRepository
	EventSvc ports.EventService
	AuditSvc ports.AuditService
	Logger   *slog.Logger
}

// NewDatabaseService constructs a DatabaseService with its dependencies.
func NewDatabaseService(params DatabaseServiceParams) *DatabaseService {
	return &DatabaseService{
		repo:     params.Repo,
		compute:  params.Compute,
		vpcRepo:  params.VpcRepo,
		eventSvc: params.EventSvc,
		auditSvc: params.AuditSvc,
		logger:   params.Logger,
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

	imageName, env, defaultPort := s.getEngineConfig(dbEngine, version, username, password, name)

	networkID, err := s.resolveVpcNetwork(ctx, vpcID)
	if err != nil {
		return nil, err
	}

	dockerName := fmt.Sprintf("cloud-db-%s-%s", name, db.ID.String()[:8])
	containerID, allocatedPorts, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: nil,
		Env:         env,
		Cmd:         nil,
	})
	if err != nil {
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
		return nil, err
	}

	s.recordDatabaseCreation(ctx, userID, db, engine)
	return db, nil
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

func (s *DatabaseService) getEngineConfig(engine domain.DatabaseEngine, version, username, password, name string) (string, []string, string) {
	switch engine {
	case domain.EnginePostgres:
		return fmt.Sprintf("postgres:%s-alpine", version),
			[]string{"POSTGRES_USER=" + username, "POSTGRES_PASSWORD=" + password, "POSTGRES_DB=" + name},
			"5432"
	case domain.EngineMySQL:
		return fmt.Sprintf("mysql:%s", version),
			[]string{"MYSQL_ROOT_PASSWORD=" + password, "MYSQL_DATABASE=" + name},
			"3306"
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

	// 2. Delete from repo
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
