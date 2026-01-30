// Package services implements core business workflows.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// InstanceService manages compute instance lifecycle (containers or VMs).
//
// Supports multiple backends (Docker, Libvirt) and networking modes (bridge, VPC).
// Handles instance CRUD, port mapping, volume attachment, and resource monitoring.
//
// All methods are safe for concurrent use and return domain errors.
type InstanceService struct {
	repo       ports.InstanceRepository
	vpcRepo    ports.VpcRepository
	subnetRepo ports.SubnetRepository
	volumeRepo ports.VolumeRepository
	compute    ports.ComputeBackend
	network    ports.NetworkBackend
	eventSvc   ports.EventService
	auditSvc   ports.AuditService
	dnsSvc     ports.DNSService
	taskQueue  ports.TaskQueue
	logger     *slog.Logger
}

// InstanceServiceParams holds dependencies for InstanceService creation.
// Uses parameter object pattern for cleaner dependency injection.
type InstanceServiceParams struct {
	Repo       ports.InstanceRepository
	VpcRepo    ports.VpcRepository
	SubnetRepo ports.SubnetRepository
	VolumeRepo ports.VolumeRepository
	Compute    ports.ComputeBackend
	Network    ports.NetworkBackend
	EventSvc   ports.EventService
	AuditSvc   ports.AuditService
	DNSSvc     ports.DNSService
	TaskQueue  ports.TaskQueue // Optional
	Logger     *slog.Logger
}

// NewInstanceService creates a new InstanceService with the given dependencies.
func NewInstanceService(params InstanceServiceParams) *InstanceService {
	return &InstanceService{
		repo:       params.Repo,
		vpcRepo:    params.VpcRepo,
		subnetRepo: params.SubnetRepo,
		volumeRepo: params.VolumeRepo,
		compute:    params.Compute,
		network:    params.Network,
		eventSvc:   params.EventSvc,
		auditSvc:   params.AuditSvc,
		dnsSvc:     params.DNSSvc,
		taskQueue:  params.TaskQueue,
		logger:     params.Logger,
	}
}

// LaunchInstance provisions a new instance, sets up its network (if VPC/Subnet provided),
// and attaches any requested volumes.
func (s *InstanceService) LaunchInstance(ctx context.Context, name, image, ports string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	ctx, span := otel.Tracer("instance-service").Start(ctx, "LaunchInstance")
	defer span.End()

	span.SetAttributes(
		attribute.String("instance.name", name),
		attribute.String("instance.image", image),
	)

	// 1. Validate ports if provided
	_, err := s.parseAndValidatePorts(ports)
	if err != nil {
		return nil, err
	}

	// 2. Create domain entity
	inst := &domain.Instance{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      name,
		Image:     image,
		Status:    domain.StatusStarting,
		Ports:     ports,
		VpcID:     vpcID,
		SubnetID:  subnetID,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, inst); err != nil {
		return nil, err
	}

	// 4. Enqueue provision task
	job := domain.ProvisionJob{
		InstanceID: inst.ID,
		UserID:     inst.UserID,
		TenantID:   inst.TenantID,
		Volumes:    volumes,
	}

	s.logger.Info("enqueueing provision job", "instance_id", inst.ID, "queue", "provision_queue", "tenant_id", inst.TenantID)
	if err := s.taskQueue.Enqueue(ctx, "provision_queue", job); err != nil {
		s.logger.Error("failed to enqueue provision job", "instance_id", inst.ID, "error", err)
		// Fallback to sync if queue fails or just return error
		// For now, return error as we want to guarantee the queue works for 1k users
		return nil, errors.Wrap(errors.Internal, "failed to enqueue provisioning task", err)
	}

	return inst, nil
}

// Provision contains the heavy lifting of instance launch, called by background workers.
func (s *InstanceService) Provision(ctx context.Context, instanceID uuid.UUID, volumes []domain.VolumeAttachment) error {
	inst, err := s.repo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}

	// 1. Resolve Networking
	networkID, err := s.provisionNetwork(ctx, inst)
	if err != nil {
		s.updateStatus(ctx, inst, domain.StatusError)
		return err
	}

	// 2. Resolve Volumes
	volumeBinds, attachedVolumes, err := s.resolveVolumes(ctx, volumes)
	if err != nil {
		s.updateStatus(ctx, inst, domain.StatusError)
		return err
	}

	// 3. Create Instance
	dockerName := s.formatContainerName(inst.ID)
	portList, _ := s.parseAndValidatePorts(inst.Ports)
	containerID, err := s.compute.CreateInstance(ctx, ports.CreateInstanceOptions{
		Name:        dockerName,
		ImageName:   inst.Image,
		Ports:       portList,
		NetworkID:   networkID,
		VolumeBinds: volumeBinds,
		Env:         nil,
		Cmd:         nil,
	})
	if err != nil {
		platform.InstanceOperationsTotal.WithLabelValues("launch", "failure").Inc()
		s.updateStatus(ctx, inst, domain.StatusError)
		return errors.Wrap(errors.Internal, "failed to launch container", err)
	}

	// 4. Finalize
	return s.finalizeProvision(ctx, inst, containerID, attachedVolumes)
}

func (s *InstanceService) provisionNetwork(ctx context.Context, inst *domain.Instance) (string, error) {
	if s.compute.Type() == "noop" {
		inst.PrivateIP = "127.0.0.1"
		return "", nil
	}

	networkID, allocatedIP, ovsPort, err := s.resolveNetworkConfig(ctx, inst.VpcID, inst.SubnetID)
	if err != nil {
		return "", err
	}

	inst.PrivateIP = allocatedIP
	inst.OvsPort = ovsPort
	return networkID, nil
}

func (s *InstanceService) finalizeProvision(ctx context.Context, inst *domain.Instance, containerID string, attachedVolumes []*domain.Volume) error {
	if err := s.plumbNetwork(ctx, inst, containerID); err != nil {
		s.logger.Warn("failed to plumb network", "error", err)
	}

	inst.Status = domain.StatusRunning
	inst.ContainerID = containerID

	// If IP was not allocated during provision (e.g. Docker dynamic), fetch it now
	if inst.PrivateIP == "" {
		ip, err := s.compute.GetInstanceIP(ctx, containerID)
		if err == nil && ip != "" {
			inst.PrivateIP = ip
		} else {
			s.logger.Warn("failed to get instance IP from backend", "instance_id", inst.ID, "error", err)
		}
	}

	// 5. Register DNS (if applicable)
	if s.dnsSvc != nil && inst.PrivateIP != "" {
		if err := s.dnsSvc.RegisterInstance(ctx, inst, inst.PrivateIP); err != nil {
			s.logger.Warn("failed to register instance DNS", "error", err, "instance", inst.Name)
			// Don't fail provisioning for DNS failure
		}
	}

	if err := s.repo.Update(ctx, inst); err != nil {
		return err
	}

	s.updateVolumesAfterLaunch(ctx, attachedVolumes, inst.ID)

	_ = s.eventSvc.RecordEvent(ctx, "INSTANCE_LAUNCH", inst.ID.String(), "INSTANCE", map[string]interface{}{
		"name":  inst.Name,
		"image": inst.Image,
		"ip":    inst.PrivateIP,
	})

	_ = s.auditSvc.Log(ctx, inst.UserID, "instance.launch", "instance", inst.ID.String(), map[string]interface{}{
		"name":  inst.Name,
		"image": inst.Image,
		"ip":    inst.PrivateIP,
	})

	return nil
}

func (s *InstanceService) updateStatus(ctx context.Context, inst *domain.Instance, status domain.InstanceStatus) {
	inst.Status = status
	_ = s.repo.Update(ctx, inst)
}

func (s *InstanceService) parseAndValidatePorts(ports string) ([]string, error) {
	if ports == "" {
		return nil, nil
	}

	portList := strings.Split(ports, ",")
	if len(portList) > domain.MaxPortsPerInstance {
		return nil, errors.New(errors.TooManyPorts, fmt.Sprintf("max %d ports allowed", domain.MaxPortsPerInstance))
	}

	for _, p := range portList {
		if err := validatePortMapping(p); err != nil {
			return nil, err
		}
	}

	return portList, nil
}

func validatePortMapping(p string) error {
	idx := strings.Index(p, ":")
	if idx == -1 || strings.Contains(p[idx+1:], ":") {
		return errors.New(errors.InvalidPortFormat, "port format must be host:container")
	}

	hostPart := p[:idx]
	containerPart := p[idx+1:]

	hostPort, err := parsePort(hostPart)
	if err != nil {
		return errors.New(errors.InvalidPortFormat, fmt.Sprintf("invalid host port: %s", hostPart))
	}
	containerPort, err := parsePort(containerPart)
	if err != nil {
		return errors.New(errors.InvalidPortFormat, fmt.Sprintf("invalid container port: %s", containerPart))
	}

	if hostPort < domain.MinPort || hostPort > domain.MaxPort {
		return errors.New(errors.InvalidPortFormat, fmt.Sprintf("host port %d out of range (%d-%d)", hostPort, domain.MinPort, domain.MaxPort))
	}
	if containerPort < domain.MinPort || containerPort > domain.MaxPort {
		return errors.New(errors.InvalidPortFormat, fmt.Sprintf("container port %d out of range (%d-%d)", containerPort, domain.MinPort, domain.MaxPort))
	}

	return nil
}

func parsePort(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty port")
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return port, nil
}

// StopInstance halts a running instance's associated compute resource (e.g., container).
func (s *InstanceService) StopInstance(ctx context.Context, idOrName string) error {
	// 1. Get from DB (handles both Name and UUID)
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return err
	}

	if inst.Status == domain.StatusStopped {
		return nil // Already stopped
	}

	// 2. Call Docker stop
	target := inst.ContainerID
	if target == "" {
		// Fallback to Reconstruction
		target = s.formatContainerName(inst.ID)
	}

	if err := s.compute.StopInstance(ctx, target); err != nil {
		platform.InstanceOperationsTotal.WithLabelValues("stop", "failure").Inc()
		s.logger.Error("failed to stop docker container", "container_id", target, "error", err)
		return errors.Wrap(errors.Internal, "failed to stop container", err)
	}

	platform.InstancesTotal.WithLabelValues("running", s.compute.Type()).Dec()
	platform.InstancesTotal.WithLabelValues("stopped", s.compute.Type()).Inc()
	platform.InstanceOperationsTotal.WithLabelValues("stop", "success").Inc()

	s.logger.Info("instance stopped", "instance_id", inst.ID)

	// 3. Update DB
	inst.Status = domain.StatusStopped
	if err := s.repo.Update(ctx, inst); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, inst.UserID, "instance.stop", "instance", inst.ID.String(), map[string]interface{}{
		"name": inst.Name,
	})

	return nil
}

// ListInstances returns all instances owned by the current user.
func (s *InstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	return s.repo.List(ctx)
}

// GetInstance retrieves an instance by its UUID or name.
func (s *InstanceService) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	// 1. Try to parse as UUID
	id, uuidErr := uuid.Parse(idOrName)
	if uuidErr == nil {
		return s.repo.GetByID(ctx, id)
	}
	// 2. Fallback to name lookup
	return s.repo.GetByName(ctx, idOrName)
}

// GetInstanceLogs retrieves the execution logs from the instance's compute resource.
func (s *InstanceService) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return "", err
	}

	if inst.ContainerID == "" {
		return "", errors.New(errors.InstanceNotRunning, "instance has no active container")
	}

	stream, err := s.compute.GetInstanceLogs(ctx, inst.ContainerID)
	if err != nil {
		return "", err
	}
	defer func() { _ = stream.Close() }()

	bytes, err := io.ReadAll(stream)
	if err != nil {
		return "", errors.Wrap(errors.Internal, "failed to read logs", err)
	}

	return string(bytes), nil
}

// GetConsoleURL returns the VNC console URL for an instance.
func (s *InstanceService) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return "", err
	}

	id := inst.ID.String()
	if inst.ContainerID != "" {
		id = inst.ContainerID
	}

	return s.compute.GetConsoleURL(ctx, id)
}

func (s *InstanceService) TerminateInstance(ctx context.Context, idOrName string) error {
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return err
	}

	if err := s.removeInstanceContainer(ctx, inst); err != nil {
		platform.InstanceOperationsTotal.WithLabelValues("terminate", "failure").Inc()
		return err
	}

	s.updateTerminationMetrics(inst)

	if err := s.releaseAttachedVolumes(ctx, inst.ID); err != nil {
		s.logger.Warn("failed to release volumes during termination", "instance_id", inst.ID, "error", err)
	}

	return s.finalizeTermination(ctx, inst)
}

func (s *InstanceService) updateTerminationMetrics(inst *domain.Instance) {
	switch inst.Status {
	case domain.StatusRunning:
		platform.InstancesTotal.WithLabelValues("running", s.compute.Type()).Dec()
	case domain.StatusStopped:
		platform.InstancesTotal.WithLabelValues("stopped", s.compute.Type()).Dec()
	}
	platform.InstanceOperationsTotal.WithLabelValues("terminate", "success").Inc()

	if s.dnsSvc != nil {
		_ = s.dnsSvc.UnregisterInstance(context.Background(), inst.ID)
	}
}

func (s *InstanceService) finalizeTermination(ctx context.Context, inst *domain.Instance) error {
	if err := s.repo.Delete(ctx, inst.ID); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "INSTANCE_TERMINATE", inst.ID.String(), "INSTANCE", map[string]interface{}{})

	_ = s.auditSvc.Log(ctx, inst.UserID, "instance.terminate", "instance", inst.ID.String(), map[string]interface{}{
		"name": inst.Name,
	})

	return nil
}

func (s *InstanceService) removeInstanceContainer(ctx context.Context, inst *domain.Instance) error {
	containerID := inst.ContainerID
	if containerID == "" {
		// Fallback to Reconstruction for legacy or missing ID
		containerID = s.formatContainerName(inst.ID)
	}

	if err := s.compute.DeleteInstance(ctx, containerID); err != nil {
		s.logger.Warn("failed to remove docker container", "container_id", containerID, "error", err)
		return errors.Wrap(errors.Internal, "failed to remove container", err)
	}

	s.logger.Info("instance terminated", "instance_id", inst.ID)
	return nil
}

// releaseAttachedVolumes marks all volumes attached to an instance as available
func (s *InstanceService) releaseAttachedVolumes(ctx context.Context, instanceID uuid.UUID) error {
	volumes, err := s.volumeRepo.ListByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	for _, vol := range volumes {
		vol.Status = domain.VolumeStatusAvailable
		vol.InstanceID = nil
		vol.MountPath = ""
		vol.UpdatedAt = time.Now()

		if err := s.volumeRepo.Update(ctx, vol); err != nil {
			s.logger.Warn("failed to release volume", "volume_id", vol.ID, "error", err)
			continue
		}
		s.logger.Info("volume released during instance termination", "volume_id", vol.ID, "instance_id", instanceID)
	}

	return nil
}

// GetInstanceStats retrieves real-time CPU and Memory usage for an instance.
func (s *InstanceService) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return nil, err
	}

	if inst.ContainerID == "" {
		return nil, errors.New(errors.InstanceNotRunning, "instance not running")
	}

	stream, err := s.compute.GetInstanceStats(ctx, inst.ContainerID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get stats stream", err)
	}
	defer func() { _ = stream.Close() }()

	var stats domain.RawDockerStats
	if err := json.NewDecoder(stream).Decode(&stats); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to decode stats", err)
	}

	return s.calculateInstanceStats(&stats), nil
}

func (s *InstanceService) calculateInstanceStats(stats *domain.RawDockerStats) *domain.InstanceStats {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage) - float64(stats.PreCPUStats.SystemCPUUsage)

	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * 100.0
	}

	memUsage := float64(stats.MemoryStats.Usage)
	memLimit := float64(stats.MemoryStats.Limit)
	memPercent := 0.0
	if memLimit > 0 {
		memPercent = (memUsage / memLimit) * 100.0
	}

	return &domain.InstanceStats{
		CPUPercentage:    cpuPercent,
		MemoryUsageBytes: memUsage,
		MemoryLimitBytes: memLimit,
		MemoryPercentage: memPercent,
	}
}

func (s *InstanceService) getVolumeByIDOrName(ctx context.Context, idOrName string) (*domain.Volume, error) {
	id, err := uuid.Parse(idOrName)
	if err == nil {
		return s.volumeRepo.GetByID(ctx, id)
	}
	return s.volumeRepo.GetByName(ctx, idOrName)
}

func (s *InstanceService) updateVolumesAfterLaunch(ctx context.Context, volumes []*domain.Volume, instanceID uuid.UUID) {
	for _, vol := range volumes {
		vol.Status = domain.VolumeStatusInUse
		vol.InstanceID = &instanceID
		vol.UpdatedAt = time.Now()
		if err := s.volumeRepo.Update(ctx, vol); err != nil {
			s.logger.Warn("failed to update volume status", "volume_id", vol.ID, "error", err)
		}
	}
}
func (s *InstanceService) allocateIP(ctx context.Context, subnet *domain.Subnet) (string, error) {
	_, ipNet, err := net.ParseCIDR(subnet.CIDRBlock)
	if err != nil {
		return "", err
	}

	instances, err := s.repo.ListBySubnet(ctx, subnet.ID)
	if err != nil {
		return "", err
	}

	usedIPs := make(map[string]bool)
	for _, inst := range instances {
		if inst.PrivateIP != "" {
			usedIPs[inst.PrivateIP] = true
		}
	}
	usedIPs[subnet.GatewayIP] = true

	// Find first available IP
	ip, err := s.findAvailableIP(ipNet, usedIPs)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (s *InstanceService) isValidHostIP(ip net.IP, n *net.IPNet) bool {
	// Simple check: not network address and not broadcast address (if /30 or larger)
	// For simplicity in this demo, we just ensure it's in range and not gateway
	return n.Contains(ip)
}

func (s *InstanceService) resolveNetworkConfig(ctx context.Context, vpcID, subnetID *uuid.UUID) (string, string, string, error) {
	var networkID string
	if vpcID != nil {
		vpc, err := s.vpcRepo.GetByID(ctx, *vpcID)
		if err != nil {
			s.logger.Error("failed to get VPC", "vpc_id", vpcID, "error", err)
			return "", "", "", err
		}
		networkID = vpc.NetworkID
	}

	// HACK: For Docker-based demo without full OVS integration, force use of shared network
	// because 'br-vpc-xxx' OVS bridge doesn't exist as a Docker network.
	if s.compute.Type() == "docker" {
		networkID = "cloud-network"

		// If no subnet is configured, we let the backend assign an IP (dynamic).
		// We return empty string here, and LaunchInstance should fetch the real IP later.
		if subnetID == nil {
			return networkID, "", "", nil
		}
	}

	if subnetID == nil || s.network == nil {
		return networkID, "", "", nil
	}

	subnet, err := s.subnetRepo.GetByID(ctx, *subnetID)
	if err != nil {
		return "", "", "", errors.Wrap(errors.NotFound, "subnet not found", err)
	}

	// Dynamic IP allocation
	allocatedIP, err := s.allocateIP(ctx, subnet)
	if err != nil {
		return "", "", "", errors.Wrap(errors.ResourceLimitExceeded, "failed to allocate IP in subnet", err)
	}

	ovsPort := "veth-" + uuid.New().String()[:8]
	return networkID, allocatedIP, ovsPort, nil
}

func (s *InstanceService) resolveVolumes(ctx context.Context, volumes []domain.VolumeAttachment) ([]string, []*domain.Volume, error) {
	var volumeBinds []string
	var attachedVolumes []*domain.Volume
	for _, va := range volumes {
		vol, err := s.getVolumeByIDOrName(ctx, va.VolumeIDOrName)
		if err != nil {
			s.logger.Error("failed to get volume", "volume", va.VolumeIDOrName, "error", err)
			return nil, nil, errors.Wrap(errors.NotFound, "volume "+va.VolumeIDOrName+" not found", err)
		}
		if vol.Status != domain.VolumeStatusAvailable {
			return nil, nil, errors.New(errors.InvalidInput, "volume "+vol.Name+" is not available")
		}
		volName := "thecloud-vol-" + vol.ID.String()[:8]
		if vol.BackendPath != "" {
			volName = vol.BackendPath
		}
		volumeBinds = append(volumeBinds, volName+":"+va.MountPath)
		attachedVolumes = append(attachedVolumes, vol)
	}
	return volumeBinds, attachedVolumes, nil
}

func (s *InstanceService) plumbNetwork(ctx context.Context, inst *domain.Instance, _ string) error {
	if inst.OvsPort == "" || s.network == nil {
		return nil
	}

	vethContainer := "eth0-" + inst.ID.String()[:8]
	if err := s.network.CreateVethPair(ctx, inst.OvsPort, vethContainer); err != nil {
		// In Docker/Dev mode without real OVS, this might fail. We log and continue
		// to allow the instance to run (albeit without custom networking).
		s.logger.Warn("failed to create veth pair (networking might be limited)", "error", err)
		return nil
	}

	if inst.VpcID != nil {
		if err := s.attachToVpcBridge(ctx, *inst.VpcID, inst.OvsPort); err != nil {
			return err
		}
	}

	if inst.SubnetID != nil {
		return s.configureVethIP(ctx, *inst.SubnetID, vethContainer, inst.PrivateIP)
	}
	return nil
}

func (s *InstanceService) attachToVpcBridge(ctx context.Context, vpcID uuid.UUID, ovsPort string) error {
	vpc, err := s.vpcRepo.GetByID(ctx, vpcID)
	if err != nil || vpc == nil {
		return err
	}
	return s.network.AttachVethToBridge(ctx, vpc.NetworkID, ovsPort)
}

func (s *InstanceService) configureVethIP(ctx context.Context, subnetID uuid.UUID, vethContainer, privateIP string) error {
	subnet, err := s.subnetRepo.GetByID(ctx, subnetID)
	if err != nil || subnet == nil {
		return err
	}
	_, ipNet, _ := net.ParseCIDR(subnet.CIDRBlock)
	ones, _ := ipNet.Mask.Size()
	return s.network.SetVethIP(ctx, vethContainer, privateIP, strconv.Itoa(ones))
}

func (s *InstanceService) formatContainerName(id uuid.UUID) string {
	return "thecloud-" + id.String()[:8]
}

func (s *InstanceService) findAvailableIP(ipNet *net.IPNet, usedIPs map[string]bool) (string, error) {
	ip := make(net.IP, len(ipNet.IP))
	copy(ip, ipNet.IP)

	for {
		// Increment IP
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] > 0 {
				break
			}
		}

		if !ipNet.Contains(ip) {
			break
		}

		if !usedIPs[ip.String()] && s.isValidHostIP(ip, ipNet) {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no available IPs in subnet")
}

func (s *InstanceService) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	inst, err := s.GetInstance(ctx, idOrName)
	if err != nil {
		return "", err
	}

	if inst.ContainerID == "" {
		return "", errors.New(errors.InstanceNotRunning, "instance not running")
	}

	// Permission check?
	// Implicitly checked by GetInstance if we had per-instance ACLs,
	// but context UserID is checked in GetInstance -> repo.Get... (actually Repo typically scopes or checks, but GetInstance checks GetByID/GetByName).
	// Let's assume caller (Handler) checked PermissionInstanceUpdate/Execute.

	output, err := s.compute.Exec(ctx, inst.ContainerID, cmd)
	if err != nil {
		return "", errors.Wrap(errors.Internal, "failed to execute command", err)
	}

	return output, nil
}
