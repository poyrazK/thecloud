// Package services implements core business workflows.
package services

import (
	"context"
	"encoding/json"
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

// CacheService manages cache clusters and their lifecycle.
type CacheService struct {
	repo     ports.CacheRepository
	compute  ports.ComputeBackend
	vpcRepo  ports.VpcRepository
	eventSvc ports.EventService
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// NewCacheService constructs a CacheService with its dependencies.
func NewCacheService(
	repo ports.CacheRepository,
	compute ports.ComputeBackend,
	vpcRepo ports.VpcRepository,
	eventSvc ports.EventService,
	auditSvc ports.AuditService,
	logger *slog.Logger,
) *CacheService {
	return &CacheService{
		repo:     repo,
		compute:  compute,
		vpcRepo:  vpcRepo,
		eventSvc: eventSvc,
		auditSvc: auditSvc,
		logger:   logger,
	}
}

func (s *CacheService) CreateCache(ctx context.Context, name, version string, memoryMB int, vpcID *uuid.UUID) (*domain.Cache, error) {
	userID := appcontext.UserIDFromContext(ctx)

	password, err := util.GenerateRandomPassword(16)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate password", err)
	}

	cache := &domain.Cache{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Engine:    domain.EngineRedis,
		Version:   version,
		Status:    domain.CacheStatusCreating,
		VpcID:     vpcID,
		Password:  password,
		MemoryMB:  memoryMB,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	networkID, err := s.resolveNetworkID(ctx, vpcID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, cache); err != nil {
		return nil, err
	}

	containerID, err := s.launchCacheContainer(ctx, cache, networkID)
	if err != nil {
		cache.Status = domain.CacheStatusFailed
		_ = s.repo.Delete(ctx, cache.ID)
		return nil, errors.Wrap(errors.Internal, "failed to launch cache container", err)
	}

	port, err := s.compute.GetInstancePort(ctx, containerID, "6379")
	if err != nil {
		s.logger.Error("failed to get cache port", "error", err)
	}

	cache.Status = domain.CacheStatusRunning
	cache.ContainerID = containerID
	cache.Port = port

	if err := s.repo.Update(ctx, cache); err != nil {
		s.logger.Warn("failed to update cache status after launch", "id", cache.ID, "error", err)
	}

	s.logCacheCreation(ctx, cache, name)

	return cache, nil
}

func (s *CacheService) resolveNetworkID(ctx context.Context, vpcID *uuid.UUID) (string, error) {
	if vpcID == nil {
		return "", nil
	}
	vpc, err := s.vpcRepo.GetByID(ctx, *vpcID)
	if err != nil {
		s.logger.Error("failed to get VPC", "vpc_id", vpcID, "error", err)
		return "", err
	}
	return vpc.NetworkID, nil
}

func (s *CacheService) launchCacheContainer(ctx context.Context, cache *domain.Cache, networkID string) (string, error) {
	dockerName := fmt.Sprintf("thecloud-cache-%s", cache.ID.String()[:8])
	imageName := fmt.Sprintf("redis:%s-alpine", cache.Version)

	cmd := []string{
		"redis-server",
		"--appendonly", "yes",
		"--save", "",
		"--requirepass", cache.Password,
		"--maxmemory", fmt.Sprintf("%dmb", cache.MemoryMB),
		"--maxmemory-policy", "allkeys-lru",
		"--tcp-keepalive", "300",
	}

	containerID, err := s.compute.CreateInstance(ctx, ports.CreateInstanceOptions{
		Name:      dockerName,
		ImageName: imageName,
		Ports:     []string{"0:6379"},
		NetworkID: networkID,
		Cmd:       cmd,
	})
	if err != nil {
		s.logger.Error("failed to create cache container", "error", err)
		return "", err
	}
	return containerID, nil
}

func (s *CacheService) logCacheCreation(ctx context.Context, cache *domain.Cache, originalName string) {
	_ = s.eventSvc.RecordEvent(ctx, "CACHE_CREATE", cache.ID.String(), "CACHE", map[string]interface{}{
		"name":    cache.Name,
		"version": cache.Version,
		"memory":  cache.MemoryMB,
	})

	_ = s.auditSvc.Log(ctx, cache.UserID, "cache.create", "cache", cache.ID.String(), map[string]interface{}{
		"name": originalName,
	})

	platform.CacheInstancesTotal.WithLabelValues("running").Inc()
}

func (s *CacheService) GetCache(ctx context.Context, idOrName string) (*domain.Cache, error) {
	return s.getCacheByIDOrName(ctx, idOrName)
}

func (s *CacheService) ListCaches(ctx context.Context) ([]*domain.Cache, error) {
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.List(ctx, userID)
}

func (s *CacheService) DeleteCache(ctx context.Context, idOrName string) error {
	cache, err := s.getCacheByIDOrName(ctx, idOrName)
	if err != nil {
		return err
	}

	if cache.ContainerID != "" {
		if err := s.compute.StopInstance(ctx, cache.ContainerID); err != nil {
			s.logger.Warn("failed to stop cache container", "container_id", cache.ContainerID, "error", err)
		}
		if err := s.compute.DeleteInstance(ctx, cache.ContainerID); err != nil {
			s.logger.Warn("failed to remove cache container", "container_id", cache.ContainerID, "error", err)
		}
	}

	if err := s.repo.Delete(ctx, cache.ID); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "CACHE_DELETE", cache.ID.String(), "CACHE", nil)

	_ = s.auditSvc.Log(ctx, cache.UserID, "cache.delete", "cache", cache.ID.String(), map[string]interface{}{
		"name": cache.Name,
	})

	platform.CacheInstancesTotal.WithLabelValues("running").Dec()

	return nil
}

func (s *CacheService) GetConnectionString(ctx context.Context, idOrName string) (string, error) {
	cache, err := s.getCacheByIDOrName(ctx, idOrName)
	if err != nil {
		return "", err
	}
	// format: redis://:password@host:port
	// We assume localhost for now as we don't have public IPs yet
	return fmt.Sprintf("redis://:%s@localhost:%d", cache.Password, cache.Port), nil
}

func (s *CacheService) getCacheByIDOrName(ctx context.Context, idOrName string) (*domain.Cache, error) {
	id, err := uuid.Parse(idOrName)
	if err == nil {
		return s.repo.GetByID(ctx, id)
	}
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.GetByName(ctx, userID, idOrName)
}

func (s *CacheService) FlushCache(ctx context.Context, idOrName string) error {
	cache, err := s.getCacheByIDOrName(ctx, idOrName)
	if err != nil {
		return err
	}

	if cache.Status != domain.CacheStatusRunning {
		return errors.New(errors.InstanceNotRunning, "cache is not running")
	}

	// Exec FLUSHALL inside the container
	// We need to pass the password if set.
	cmd := []string{"redis-cli"}
	if cache.Password != "" {
		cmd = append(cmd, "-a", cache.Password)
	}
	cmd = append(cmd, "FLUSHALL")

	output, err := s.compute.Exec(ctx, cache.ContainerID, cmd)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to flush cache: "+output, err)
	}

	_ = s.auditSvc.Log(ctx, cache.UserID, "cache.flush", "cache", cache.ID.String(), map[string]interface{}{})

	return nil
}

func (s *CacheService) GetCacheStats(ctx context.Context, idOrName string) (*ports.CacheStats, error) {
	cache, err := s.getCacheByIDOrName(ctx, idOrName)
	if err != nil {
		return nil, err
	}

	if cache.Status != domain.CacheStatusRunning {
		return nil, errors.New(errors.InstanceNotRunning, "cache is not running")
	}

	stream, err := s.compute.GetInstanceStats(ctx, cache.ContainerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = stream.Close() }()

	// Parse Docker Stats (standard JSON)
	var dockerStats struct {
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}
	if err := json.NewDecoder(stream).Decode(&dockerStats); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to decode stats", err)
	}

	result := &ports.CacheStats{
		UsedMemoryBytes:  int64(dockerStats.MemoryStats.Usage),
		MaxMemoryBytes:   int64(dockerStats.MemoryStats.Limit),
		ConnectedClients: 0,
		TotalKeys:        0,
	}

	// Try to get Redis Internal Stats
	cmd := []string{"redis-cli"}
	if cache.Password != "" {
		cmd = append(cmd, "-a", cache.Password)
	}
	cmd = append(cmd, "INFO")

	output, err := s.compute.Exec(ctx, cache.ContainerID, cmd)
	if err == nil {
		result.ConnectedClients = parseRedisClients(output)
		result.TotalKeys = parseRedisKeys(output)
	} else {
		s.logger.Warn("failed to get redis internal stats", "error", err)
	}

	return result, nil
}

func parseRedisClients(info string) int {
	// Look for connected_clients:N
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "connected_clients:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				val, _ := strconv.Atoi(parts[1])
				return val
			}
		}
	}
	return 0
}

func parseRedisKeys(info string) int64 {
	var total int64
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "db") && strings.Contains(line, "keys=") {
			total += parseKeysFromLine(line)
		}
	}
	return total
}

func parseKeysFromLine(line string) int64 {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return 0
	}
	stats := parts[1]
	pairs := strings.Split(stats, ",")
	for _, pair := range pairs {
		if strings.HasPrefix(pair, "keys=") {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 {
				val, _ := strconv.ParseInt(kv[1], 10, 64)
				return val
			}
		}
	}
	return 0
}
