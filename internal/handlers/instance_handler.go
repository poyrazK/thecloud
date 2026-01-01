package httphandlers

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/poyraz/cloud/internal/core/ports"
	"github.com/poyraz/cloud/internal/errors"
	"github.com/poyraz/cloud/pkg/httputil"
)

const (
	maxNameLength  = 64
	minNameLength  = 1
	maxImageLength = 256
)

type InstanceHandler struct {
	svc ports.InstanceService
}

func NewInstanceHandler(svc ports.InstanceService) *InstanceHandler {
	return &InstanceHandler{svc: svc}
}

type VolumeAttachmentRequest struct {
	VolumeID  string `json:"volume_id"`
	MountPath string `json:"mount_path"`
}

type LaunchRequest struct {
	Name    string                    `json:"name" binding:"required"`
	Image   string                    `json:"image" binding:"required"`
	Ports   string                    `json:"ports"`
	VpcID   string                    `json:"vpc_id"`
	Volumes []VolumeAttachmentRequest `json:"volumes"`
}

// validateLaunchRequest performs custom validation beyond struct tags
func validateLaunchRequest(req *LaunchRequest) error {
	// Validate name
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < minNameLength || len(req.Name) > maxNameLength {
		return errors.New(errors.InvalidInput, "name must be between 1 and 64 characters")
	}
	if !isValidResourceName(req.Name) {
		return errors.New(errors.InvalidInput, "name must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Validate image
	req.Image = strings.TrimSpace(req.Image)
	if req.Image == "" {
		return errors.New(errors.InvalidInput, "image is required")
	}
	if len(req.Image) > maxImageLength {
		return errors.New(errors.InvalidInput, "image name too long (max 256 characters)")
	}

	// Validate volume attachments
	for i, v := range req.Volumes {
		if strings.TrimSpace(v.VolumeID) == "" {
			return errors.New(errors.InvalidInput, "volume_id is required for volume attachment")
		}
		if strings.TrimSpace(v.MountPath) == "" {
			return errors.New(errors.InvalidInput, "mount_path is required for volume attachment")
		}
		if !strings.HasPrefix(v.MountPath, "/") {
			return errors.New(errors.InvalidInput, "mount_path must be an absolute path starting with /")
		}
		req.Volumes[i].VolumeID = strings.TrimSpace(v.VolumeID)
		req.Volumes[i].MountPath = strings.TrimSpace(v.MountPath)
	}

	return nil
}

// isValidResourceName checks if name contains only valid characters (alphanumeric, hyphen, underscore)
func isValidResourceName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func (h *InstanceHandler) Launch(c *gin.Context) {
	var req LaunchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	// Custom validation
	if err := validateLaunchRequest(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	var vpcUUID *uuid.UUID
	if req.VpcID != "" {
		id, err := uuid.Parse(req.VpcID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vpc_id format"})
			return
		}
		vpcUUID = &id
	}

	// Convert volume attachments
	var volumes []domain.VolumeAttachment
	for _, v := range req.Volumes {
		volumes = append(volumes, domain.VolumeAttachment{
			VolumeIDOrName: v.VolumeID,
			MountPath:      v.MountPath,
		})
	}

	inst, err := h.svc.LaunchInstance(c.Request.Context(), req.Name, req.Image, req.Ports, vpcUUID, volumes)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, inst)
}

func (h *InstanceHandler) List(c *gin.Context) {
	instances, err := h.svc.ListInstances(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, instances)
}

func (h *InstanceHandler) Stop(c *gin.Context) {
	idStr := c.Param("id")

	if err := h.svc.StopInstance(c.Request.Context(), idStr); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "instance stop initiated"})
}

func (h *InstanceHandler) GetLogs(c *gin.Context) {
	idStr := c.Param("id")

	logs, err := h.svc.GetInstanceLogs(c.Request.Context(), idStr)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	c.String(http.StatusOK, logs)
}

func (h *InstanceHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	inst, err := h.svc.GetInstance(c.Request.Context(), idStr)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, inst)
}

func (h *InstanceHandler) Terminate(c *gin.Context) {
	idStr := c.Param("id")

	if err := h.svc.TerminateInstance(c.Request.Context(), idStr); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "instance terminated"})
}

func (h *InstanceHandler) GetStats(c *gin.Context) {
	idStr := c.Param("id")
	stats, err := h.svc.GetInstanceStats(c.Request.Context(), idStr)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, stats)
}
