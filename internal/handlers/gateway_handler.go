// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// GatewayHandler handles API gateway HTTP endpoints.
type GatewayHandler struct {
	svc ports.GatewayService
}

// NewGatewayHandler constructs a GatewayHandler.
func NewGatewayHandler(svc ports.GatewayService) *GatewayHandler {
	return &GatewayHandler{svc: svc}
}

func (h *GatewayHandler) CreateRoute(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		PathPrefix  string   `json:"path_prefix" binding:"required"`
		TargetURL   string   `json:"target_url" binding:"required"`
		Methods     []string `json:"methods"`
		StripPrefix bool     `json:"strip_prefix"`
		RateLimit   int      `json:"rate_limit"`
		Priority    int      `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid request body"))
		return
	}

	if req.RateLimit == 0 {
		req.RateLimit = 100
	}

	params := ports.CreateRouteParams{
		Name:        req.Name,
		Pattern:     req.PathPrefix,
		Target:      req.TargetURL,
		Methods:     req.Methods,
		StripPrefix: req.StripPrefix,
		RateLimit:   req.RateLimit,
		Priority:    req.Priority,
	}

	route, err := h.svc.CreateRoute(c.Request.Context(), params)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, route)
}

func (h *GatewayHandler) ListRoutes(c *gin.Context) {
	routes, err := h.svc.ListRoutes(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, routes)
}

func (h *GatewayHandler) DeleteRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid route ID"))
		return
	}

	if err := h.svc.DeleteRoute(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Route deleted"})
}

func (h *GatewayHandler) Proxy(c *gin.Context) {
	path := c.Param("proxy") // Expecting routes like /gw/*proxy
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	proxy, params, ok := h.svc.GetProxy(c.Request.Method, path)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "No route found for " + path})
		return
	}

	// Inject parameters into request context for downstream services if needed
	if len(params) > 0 {
		for k, v := range params {
			c.Set("path_param_"+k, v)
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
