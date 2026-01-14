// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// StackHandler handles stack HTTP endpoints.
type StackHandler struct {
	svc ports.StackService
}

// NewStackHandler constructs a StackHandler.
func NewStackHandler(svc ports.StackService) *StackHandler {
	return &StackHandler{svc: svc}
}

// CreateStackRequest is the payload for creating a stack.
type CreateStackRequest struct {
	Name       string            `json:"name" binding:"required"`
	Template   string            `json:"template" binding:"required"`
	Parameters map[string]string `json:"parameters"`
}

// Create godoc
// @Summary Create a new stack
// @Description Creates a stack from a YAML/JSON template
// @Tags IaC
// @Accept json
// @Produce json
// @Param request body CreateStackRequest true "Stack details"
// @Success 201 {object} domain.Stack
// @Router /iac/stacks [post]
func (h *StackHandler) Create(c *gin.Context) {
	var req CreateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	stack, err := h.svc.CreateStack(c.Request.Context(), req.Name, req.Template, req.Parameters)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, stack)
}

// List godoc
// @Summary List stacks
// @Description Returns all stacks for the user
// @Tags IaC
// @Produce json
// @Success 200 {array} domain.Stack
// @Router /iac/stacks [get]
func (h *StackHandler) List(c *gin.Context) {
	stacks, err := h.svc.ListStacks(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, stacks)
}

// Get godoc
// @Summary Get stack details
// @Description Returns a single stack by ID
// @Tags IaC
// @Produce json
// @Param id path string true "Stack ID"
// @Success 200 {object} domain.Stack
// @Router /iac/stacks/{id} [get]
func (h *StackHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	stack, err := h.svc.GetStack(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, stack)
}

// Delete godoc
// @Summary Delete stack
// @Description Deletes a stack and all its resources
// @Tags IaC
// @Param id path string true "Stack ID"
// @Success 204
// @Router /iac/stacks/{id} [delete]
func (h *StackHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DeleteStack(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "stack deleted"})
}

// Validate godoc
// @Summary Validate template
// @Description Validates an IaC template without creating a stack
// @Tags IaC
// @Accept json
// @Produce json
// @Param template body string true "Template content"
// @Success 200 {object} domain.TemplateValidateResponse
// @Router /iac/validate [post]
func (h *StackHandler) Validate(c *gin.Context) {
	var req struct {
		Template string `json:"template" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	resp, err := h.svc.ValidateTemplate(c.Request.Context(), req.Template)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
