// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// RBACHandler handles role and permission HTTP endpoints.
type RBACHandler struct {
	svc ports.RBACService
}

// NewRBACHandler constructs an RBACHandler.
func NewRBACHandler(svc ports.RBACService) *RBACHandler {
	return &RBACHandler{svc: svc}
}

// CreateRoleRequest is the payload for creating a role.
type CreateRoleRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	Permissions []domain.Permission `json:"permissions"`
}

// CreateRole godoc
// @Summary Create a new role
// @Description Creates a new role with permissions
// @Tags RBAC
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "Role details"
// @Success 201 {object} domain.Role
// @Router /rbac/roles [post]
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	if err := h.svc.CreateRole(c.Request.Context(), role); err != nil {
		httputil.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, role)
}

// ListRoles godoc
// @Summary List roles
// @Description Returns all available roles
// @Tags RBAC
// @Produce json
// @Success 200 {array} domain.Role
// @Router /rbac/roles [get]
func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetRole godoc
// @Summary Get role details
// @Description Returns a single role by ID or Name
// @Tags RBAC
// @Produce json
// @Param id path string true "Role ID or Name"
// @Success 200 {object} domain.Role
// @Router /rbac/roles/{id} [get]
func (h *RBACHandler) GetRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err == nil {
		role, err := h.svc.GetRoleByID(c.Request.Context(), id)
		if err != nil {
			httputil.Error(c, err)
			return
		}
		c.JSON(http.StatusOK, role)
		return
	}

	role, err := h.svc.GetRoleByName(c.Request.Context(), idParam)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, role)
}

// UpdateRole godoc
// @Summary Update role details
// @Description Updates an existing role's description or permissions
// @Tags RBAC
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body CreateRoleRequest true "Role details"
// @Success 200 {object} domain.Role
// @Router /rbac/roles/{id} [put]
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	role := &domain.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	if err := h.svc.UpdateRole(c.Request.Context(), role); err != nil {
		httputil.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Delete role
// @Description Deletes a role by ID
// @Tags RBAC
// @Param id path string true "Role ID"
// @Success 204
// @Router /rbac/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		// Try to look up by name and delete? No, strictly by ID for safety.
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AddPermissionRequest is the payload for adding a permission to a role.
type AddPermissionRequest struct {
	Permission domain.Permission `json:"permission" binding:"required"`
}

// AddPermission godoc
// @Summary Add permission to role
// @Description Adds a permission to an existing role
// @Tags RBAC
// @Accept json
// @Param id path string true "Role ID"
// @Param request body AddPermissionRequest true "Permission details"
// @Success 204
// @Router /rbac/roles/{id}/permissions [post]
func (h *RBACHandler) AddPermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var req AddPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.AddPermissionToRole(c.Request.Context(), id, req.Permission); err != nil {
		httputil.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RemovePermission godoc
// @Summary Remove permission from role
// @Description Removes a permission from an existing role
// @Tags RBAC
// @Accept json
// @Param id path string true "Role ID"
// @Param permission path string true "Permission name"
// @Success 204
// @Router /rbac/roles/{id}/permissions/{permission} [delete]
func (h *RBACHandler) RemovePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	permission := domain.Permission(c.Param("permission"))

	if err := h.svc.RemovePermissionFromRole(c.Request.Context(), id, permission); err != nil {
		httputil.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// BindRoleRequest is the payload for assigning a role to a user.
type BindRoleRequest struct {
	UserIdentifier string `json:"user_identifier" binding:"required"`
	RoleName       string `json:"role_name" binding:"required"`
}

// BindRole godoc
// @Summary Bind role to user
// @Description Assigns a role to a user
// @Tags RBAC
// @Accept json
// @Param request body BindRoleRequest true "Binding details"
// @Success 204
// @Router /rbac/bindings [post]
func (h *RBACHandler) BindRole(c *gin.Context) {
	var req BindRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.BindRole(c.Request.Context(), req.UserIdentifier, req.RoleName); err != nil {
		httputil.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRoleBindings godoc
// @Summary List role bindings
// @Description Returns all user-role assignments
// @Tags RBAC
// @Produce json
// @Success 200 {array} domain.User
// @Router /rbac/bindings [get]
func (h *RBACHandler) ListRoleBindings(c *gin.Context) {
	bindings, err := h.svc.ListRoleBindings(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, bindings)
}
