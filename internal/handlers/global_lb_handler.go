package httphandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
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
	Name        string                         `json:"name" binding:"required"`
	Hostname    string                         `json:"hostname" binding:"required"`
	Policy      domain.RoutingPolicy           `json:"policy" binding:"required,oneof=LATENCY GEOLOCATION WEIGHTED FAILOVER"`
	HealthCheck domain.GlobalHealthCheckConfig `json:"health_check" binding:"required"`
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
// @Summary Create a new global load balancer
// @Tags global-lb
// @Accept json
// @Produce json
// @Param request body CreateGlobalLBRequest true "GLB creation request"
// @Success 201 {object} domain.GlobalLoadBalancer
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /global-lb [post]
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
// @Summary Get a global load balancer by ID
// @Tags global-lb
// @Produce json
// @Param id path string true "Global LB ID"
// @Success 200 {object} domain.GlobalLoadBalancer
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /global-lb/{id} [get]
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
// @Summary List all global load balancers
// @Tags global-lb
// @Produce json
// @Success 200 {array} domain.GlobalLoadBalancer
// @Failure 500 {object} httputil.Response
// @Router /global-lb [get]
func (h *GlobalLBHandler) List(c *gin.Context) {
	userID := appcontext.UserIDFromContext(c.Request.Context())
	glbs, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 200, glbs)
}

// Delete handles DELETE /global-lb/:id
// @Summary Delete a global load balancer
// @Tags global-lb
// @Param id path string true "Global LB ID"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /global-lb/{id} [delete]
func (h *GlobalLBHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id"))
		return
	}

	userID := appcontext.UserIDFromContext(c.Request.Context())
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, 204, nil)
}

// AddEndpoint handles POST /global-lb/:id/endpoints
// @Summary Add an endpoint to a global load balancer
// @Tags global-lb
// @Accept json
// @Produce json
// @Param id path string true "Global LB ID"
// @Param request body AddGlobalEndpointRequest true "Endpoint details"
// @Success 201 {object} domain.GlobalEndpoint
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /global-lb/{id}/endpoints [post]
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
// @Summary Remove an endpoint from a global load balancer
// @Tags global-lb
// @Param id path string true "Global LB ID"
// @Param epID path string true "Endpoint ID"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /global-lb/{id}/endpoints/{epID} [delete]
func (h *GlobalLBHandler) RemoveEndpoint(c *gin.Context) {
	// The GLB ID is provided via the route parameters for API consistency,
	// though the underlying service call utilizes the unique endpoint identifier.
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
