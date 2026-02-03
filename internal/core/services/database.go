// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"log/slog"
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

// NewDatabaseService constructs a DatabaseService with its dependencies.
func NewDatabaseService(
	repo ports.DatabaseRepository,
	compute ports.ComputeBackend,
	vpcRepo ports.VpcRepository,
	eventSvc ports.EventService,
	auditSvc ports.AuditService,
	logger *slog.Logger,
) *DatabaseService {
	return &DatabaseService{
		repo:     repo,
		compute:  compute,
		vpcRepo:  vpcRepo,
		eventSvc: eventSvc,
		auditSvc: auditSvc,
		logger:   logger,
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
	containerID, err := s.compute.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   imageName,
		Ports:       []string{"0:" + defaultPort},
		NetworkID:   networkID,
		VolumeBinds: nil,
		Env:         env,
		Cmd:         nil,
	})
	if err != nil {
		s.logger.Error("failed to create database container", "error", err)
		return nil, errors.Wrap(errors.Internal, "failed to launch database container", err)
	}

	hostPort, _ := s.compute.GetInstancePort(ctx, containerID, defaultPort)
	db.ContainerID = containerID
	db.Port = hostPort
	db.Status = domain.DatabaseStatusRunning

	if err := s.repo.Create(ctx, db); err != nil {
		_ = s.compute.DeleteInstance(ctx, containerID)
		return nil, err
	}

	s.recordDatabaseCreation(ctx, userID, db, engine)
	return db, nil
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
	_ = s.eventSvc.RecordEvent(ctx, "DATABASE_CREATE", db.ID.String(), "DATABASE", map[string]interface{}{
		"name":   db.Name,
		"engine": db.Engine,
	})

	_ = s.auditSvc.Log(ctx, userID, "database.create", "database", db.ID.String(), map[string]interface{}{
		"name":   db.Name,
		"engine": originalEngine,
	})

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
