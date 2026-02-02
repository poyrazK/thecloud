package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

type GlobalLBHandler struct {
	svc ports.GlobalLBService
}

func NewGlobalLBHandler(svc ports.GlobalLBService) *GlobalLBHandler {
	return &GlobalLBHandler{svc: svc}
}

// CreateRequest defines the payload for creating a Global LB
type CreateGlobalLBRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Hostname    string                   `json:"hostname" binding:"required"`
	Policy      domain.RoutingPolicy     `json:"policy" binding:"required,oneof=LATENCY GEOLOCATION WEIGHTED FAILOVER"`
	HealthCheck domain.HealthCheckConfig `json:"health_check" binding:"required"`
}

// AddEndpointRequest defines payload for adding an endpoint
type AddGlobalEndpointRequest struct {
	Region     string     `json:"region" binding:"required"`
	TargetType string     `json:"target_type" binding:"required,oneof=LB IP"`
	TargetID   *uuid.UUID `json:"target_id,omitempty"`
	TargetIP   *string    `json:"target_ip,omitempty"`
	Weight     int        `json:"weight,omitempty"`
	Priority   int        `json:"priority,omitempty"` // Default 1
}

// Create handles POST /global-lb
func (h *GlobalLBHandler) Create(c *gin.Context) {
	var req CreateGlobalLBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	glb, err := h.svc.Create(c.Request.Context(), req.Name, req.Hostname, req.Policy, req.HealthCheck)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 201, glb)
}

// Get handles GET /global-lb/:id
func (h *GlobalLBHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id"))
		return
	}

	glb, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 200, glb)
}

// List handles GET /global-lb
func (h *GlobalLBHandler) List(c *gin.Context) {
	glbs, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 200, glbs)
}

// Delete handles DELETE /global-lb/:id
func (h *GlobalLBHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 204, nil)
}

// AddEndpoint handles POST /global-lb/:id/endpoints
func (h *GlobalLBHandler) AddEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid global lb id"))
		return
	}

	var req AddGlobalEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	// Set defaults
	weight := 1
	if req.Weight > 0 {
		weight = req.Weight
	}
	priority := 1
	if req.Priority > 0 {
		priority = req.Priority
	}

	ep, err := h.svc.AddEndpoint(
		c.Request.Context(),
		id,
		req.Region,
		req.TargetType,
		req.TargetID,
		req.TargetIP,
		weight,
		priority,
	)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 201, ep)
}

// RemoveEndpoint handles DELETE /global-lb/:id/endpoints/:epID
func (h *GlobalLBHandler) RemoveEndpoint(c *gin.Context) {
	// We don't use GLB ID in the service call currently, but it's part of the route
	_ = c.Param("id")

	epID, err := uuid.Parse(c.Param("epID"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid endpoint id"))
		return
	}

	if err := h.svc.RemoveEndpoint(c.Request.Context(), epID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 204, nil)
}

func (h *GlobalLBHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/global-lb")
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/:id", h.Get)
		group.DELETE("/:id", h.Delete)

		group.POST("/:id/endpoints", h.AddEndpoint)
		group.DELETE("/:id/endpoints/:epID", h.RemoveEndpoint)
	}
}
