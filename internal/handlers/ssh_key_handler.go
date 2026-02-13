package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

type SSHKeyHandler struct {
	svc ports.SSHKeyService
}

func NewSSHKeyHandler(svc ports.SSHKeyService) *SSHKeyHandler {
	return &SSHKeyHandler{svc: svc}
}

type CreateSSHKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

// Create handles the creation of a new SSH key.
// @Summary Create SSH Key
// @Tags ssh_keys
// @Accept json
// @Produce json
// @Param request body CreateSSHKeyRequest true "SSH Key details"
// @Success 201 {object} domain.SSHKey
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /ssh-keys [post]
func (h *SSHKeyHandler) Create(c *gin.Context) {
	var req CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	key, err := h.svc.CreateKey(c.Request.Context(), req.Name, req.PublicKey)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, key)
}

// List handles listing of SSH keys for the current tenant.
// @Summary List SSH Keys
// @Tags ssh_keys
// @Produce json
// @Success 200 {array} domain.SSHKey
// @Failure 500 {object} httputil.Response
// @Router /ssh-keys [get]
func (h *SSHKeyHandler) List(c *gin.Context) {
	keys, err := h.svc.ListKeys(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, keys)
}

// Get handles retrieving a specific SSH key.
// @Summary Get SSH Key
// @Tags ssh_keys
// @Produce json
// @Param id path string true "SSH Key ID"
// @Success 200 {object} domain.SSHKey
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /ssh-keys/{id} [get]
func (h *SSHKeyHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid ssh key id"))
		return
	}

	key, err := h.svc.GetKey(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, key)
}

// Delete handles deletion of an SSH key.
// @Summary Delete SSH Key
// @Tags ssh_keys
// @Param id path string true "SSH Key ID"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /ssh-keys/{id} [delete]
func (h *SSHKeyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid ssh key id"))
		return
	}

	if err := h.svc.DeleteKey(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
