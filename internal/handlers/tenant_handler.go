// Package httphandlers exposes HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// TenantHandler handles tenant management endpoints.
type TenantHandler struct {
	svc ports.TenantService
}

// NewTenantHandler constructs a TenantHandler.
func NewTenantHandler(svc ports.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// CreateTenantRequest defines the payload for tenant creation.
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// InviteMemberRequest defines the payload for inviting a tenant member.
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// Create godoc
// @Summary Create a new tenant
// @Tags Tenant
// @Security APIKeyAuth
// @Router /tenants [post]
func (h *TenantHandler) Create(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request"))
		return
	}

	userID := httputil.GetUserID(c)
	tenant, err := h.svc.CreateTenant(c.Request.Context(), req.Name, req.Slug, userID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, tenant)
}

// InviteMember godoc
// @Summary Invite member to tenant
// @Tags Tenant
// @Security APIKeyAuth
// @Router /tenants/:id/members [post]
func (h *TenantHandler) InviteMember(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid tenant ID"))
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request"))
		return
	}

	if err := h.svc.InviteMember(c.Request.Context(), tenantID, req.Email, req.Role); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "member invited"})
}

// SwitchTenant godoc
// @Summary Switch active tenant
// @Tags Tenant
// @Security APIKeyAuth
// @Router /tenants/:id/switch [post]
func (h *TenantHandler) SwitchTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid tenant ID"))
		return
	}

	userID := httputil.GetUserID(c)
	if err := h.svc.SwitchTenant(c.Request.Context(), userID, tenantID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "tenant switched"})
}
