// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// VolumeHandler handles volume HTTP endpoints.
type VolumeHandler struct {
	svc ports.VolumeService
}

// NewVolumeHandler constructs a VolumeHandler.
func NewVolumeHandler(svc ports.VolumeService) *VolumeHandler {
	return &VolumeHandler{svc: svc}
}

// CreateVolumeRequest is the payload for volume creation.
type CreateVolumeRequest struct {
	Name   string `json:"name" binding:"required,min=3,max=64"`
	SizeGB int    `json:"size_gb" binding:"required,min=1,max=16000"`
}

// Create creates a new volume
// @Summary Create a new volume
// @Description Creates a new block storage volume
// @Tags volumes
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param request body CreateVolumeRequest true "Volume creation request"
// @Success 201 {object} domain.Volume
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /volumes [post]
func (h *VolumeHandler) Create(c *gin.Context) {
	var req CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	vol, err := h.svc.CreateVolume(c.Request.Context(), req.Name, req.SizeGB)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, vol)
}

// List returns all volumes
// @Summary List all volumes
// @Description Gets a list of all existing block storage volumes
// @Tags volumes
// @Produce json
// @Security APIKeyAuth
// @Success 200 {array} domain.Volume
// @Failure 500 {object} httputil.Response
// @Router /volumes [get]
func (h *VolumeHandler) List(c *gin.Context) {
	volumes, err := h.svc.ListVolumes(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, volumes)
}

// Get returns volume details
// @Summary Get volume details
// @Description Gets detailed information about a specific block storage volume
// @Tags volumes
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Volume ID"
// @Success 200 {object} domain.Volume
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /volumes/{id} [get]
func (h *VolumeHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	vol, err := h.svc.GetVolume(c.Request.Context(), idStr)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, vol)
}

// Delete deletes a volume
// @Summary Delete a volume
// @Description Removes a block storage volume
// @Tags volumes
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Volume ID"
// @Success 200 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /volumes/{id} [delete]
func (h *VolumeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	if err := h.svc.DeleteVolume(c.Request.Context(), idStr); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "volume deleted"})
}

// AttachRequest is the payload for attaching a volume.
type AttachRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	MountPath  string `json:"mount_path" binding:"required"`
}

// Attach attaches a volume to an instance
// @Summary Attach volume
// @Description Attaches a block storage volume to a compute instance
// @Tags volumes
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Volume ID"
// @Param request body AttachRequest true "Attachment request"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /volumes/{id}/attach [post]
func (h *VolumeHandler) Attach(c *gin.Context) {
	idStr := c.Param("id")
	var req AttachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.AttachVolume(c.Request.Context(), idStr, req.InstanceID, req.MountPath); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "volume attached"})
}

// Detach detaches a volume from an instance
// @Summary Detach volume
// @Description Detaches a block storage volume from its current compute instance
// @Tags volumes
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Volume ID"
// @Success 200 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /volumes/{id}/detach [post]
func (h *VolumeHandler) Detach(c *gin.Context) {
	idStr := c.Param("id")
	if err := h.svc.DetachVolume(c.Request.Context(), idStr); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "volume detached"})
}
