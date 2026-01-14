// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// ImageHandler handles image registry HTTP endpoints.
type ImageHandler struct {
	svc ports.ImageService
}

// NewImageHandler constructs an ImageHandler.
func NewImageHandler(svc ports.ImageService) *ImageHandler {
	return &ImageHandler{svc: svc}
}

type registerImageRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	OS          string `json:"os" binding:"required"`
	Version     string `json:"version" binding:"required"`
	IsPublic    bool   `json:"is_public"`
}

// RegisterImage godoc
// @Summary Register a new image
// @Description Register a new VM image metadata before uploading the actual file
// @Tags Images
// @Accept json
// @Produce json
// @Param request body registerImageRequest true "Registration Info"
// @Success 201 {object} domain.Image
// @Router /api/v1/images [post]
func (h *ImageHandler) RegisterImage(c *gin.Context) {
	var req registerImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.Wrap(errors.InvalidInput, err.Error(), err))
		return
	}

	img, err := h.svc.RegisterImage(c.Request.Context(), req.Name, req.Description, req.OS, req.Version, req.IsPublic)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, img)
}

// UploadImage godoc
// @Summary Upload image file
// @Description Upload the actual qcow2 file for a registered image
// @Tags Images
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Image ID"
// @Param file formData file true "Image File"
// @Success 200 {object} map[string]string
// @Router /api/v1/images/{id}/upload [post]
func (h *ImageHandler) UploadImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid image id"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "no file uploaded"))
		return
	}

	src, err := file.Open()
	if err != nil {
		httputil.Error(c, errors.Wrap(errors.Internal, "failed to open uploaded file", err))
		return
	}
	defer func() {
		if err := src.Close(); err != nil {
			httputil.Error(c, errors.Wrap(errors.Internal, "failed to close uploaded file", err))
		}
	}()

	if err := h.svc.UploadImage(c.Request.Context(), id, src); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"status": "uploaded"})
}

// ListImages godoc
// @Summary List images
// @Description List available images for the current user and public images
// @Tags Images
// @Produce json
// @Success 200 {array} domain.Image
// @Router /api/v1/images [get]
func (h *ImageHandler) ListImages(c *gin.Context) {
	userID := httputil.GetUserID(c)
	images, err := h.svc.ListImages(c.Request.Context(), userID, true)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, images)
}

// GetImage godoc
// @Summary Get image detail
// @Description Get detailed information about a specific image
// @Tags Images
// @Produce json
// @Param id path string true "Image ID"
// @Success 200 {object} domain.Image
// @Router /api/v1/images/{id} [get]
func (h *ImageHandler) GetImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id"))
		return
	}

	img, err := h.svc.GetImage(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, img)
}

// DeleteImage godoc
// @Summary Delete image
// @Description Delete an image and its associated file
// @Tags Images
// @Param id path string true "Image ID"
// @Success 200 {object} map[string]string "message: image deleted"
// @Router /api/v1/images/{id} [delete]
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id"))
		return
	}

	if err := h.svc.DeleteImage(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "image deleted"})
}
