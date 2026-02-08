// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	errs "github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// IdentityHandler handles API key management endpoints.
type IdentityHandler struct {
	svc ports.IdentityService
}

// NewIdentityHandler constructs an IdentityHandler.
func NewIdentityHandler(svc ports.IdentityService) *IdentityHandler {
	return &IdentityHandler{svc: svc}
}

// CreateKeyRequest is the payload for API key creation.
type CreateKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateKey generates a new API key
// @Summary Create a new API key
// @Tags identity
// @Accept json
// @Produce json
// @Param request body CreateKeyRequest true "Key creation request"
// @Success 201 {object} domain.APIKey
// @Failure 401 {object} httputil.Response
// @Router /auth/keys [post]
func (h *IdentityHandler) CreateKey(c *gin.Context) {
	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errs.New(errs.InvalidInput, "invalid request body"))
		return
	}

	userID := httputil.GetUserID(c)
	key, err := h.svc.CreateKey(c.Request.Context(), userID, req.Name)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, key)
}

// ListKeys lists all API keys for the current user
// @Summary List API keys
// @Tags identity
// @Produce json
// @Success 200 {array} domain.APIKey
// @Failure 401 {object} httputil.Response
// @Router /auth/keys [get]
func (h *IdentityHandler) ListKeys(c *gin.Context) {
	userID := httputil.GetUserID(c)
	keys, err := h.svc.ListKeys(c.Request.Context(), userID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, keys)
}

// RevokeKey deletes an API key
// @Summary Revoke an API key
// @Tags identity
// @Param id path string true "Key ID"
// @Success 204 "No Content"
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /auth/keys/{id} [delete]
func (h *IdentityHandler) RevokeKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errs.New(errs.InvalidInput, "invalid id"))
		return
	}

	userID := httputil.GetUserID(c)
	if err := h.svc.RevokeKey(c.Request.Context(), userID, id); err != nil {
		httputil.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RotateKey rotates an API key
// @Summary Rotate an API key
// @Tags identity
// @Param id path string true "Key ID"
// @Success 200 {object} domain.APIKey
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /auth/keys/{id}/rotate [post]
func (h *IdentityHandler) RotateKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errs.New(errs.InvalidInput, "invalid id"))
		return
	}

	userID := httputil.GetUserID(c)
	newKey, err := h.svc.RotateKey(c.Request.Context(), userID, id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, newKey)
}

// RegenerateKey regenerates an API key (alias for RotateKey)
// @Summary Regenerate an API key
// @Tags identity
// @Param id path string true "Key ID"
// @Success 200 {object} domain.APIKey
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /auth/keys/{id}/regenerate [post]
func (h *IdentityHandler) RegenerateKey(c *gin.Context) {
	h.RotateKey(c)
}
