// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// CreateRouteRequest define the payload for creating a route.
type CreateRouteRequest struct {
	Name                 string   `json:"name" binding:"required"`
	PathPrefix           string   `json:"path_prefix" binding:"required"`
	TargetURL            string   `json:"target_url" binding:"required"`
	Methods              []string `json:"methods"`
	StripPrefix          bool     `json:"strip_prefix"`
	RateLimit            int      `json:"rate_limit"`
	DialTimeout          int64    `json:"dial_timeout"`
	ResponseHeaderTimeout int64    `json:"response_header_timeout"`
	IdleConnTimeout      int64    `json:"idle_conn_timeout"`
	TLSSkipVerify        bool     `json:"tls_skip_verify"`
	RequireTLS          bool     `json:"require_tls"`
	AllowedCIDRs         []string `json:"allowed_cidrs"`
	BlockedCIDRs         []string `json:"blocked_cidrs"`
	MaxBodySize          int64    `json:"max_body_size"`
	Priority             int      `json:"priority"`
}

// GatewayHandler handles API gateway HTTP endpoints.
type GatewayHandler struct {
	svc ports.GatewayService
}

// NewGatewayHandler constructs a GatewayHandler.
func NewGatewayHandler(svc ports.GatewayService) *GatewayHandler {
	return &GatewayHandler{svc: svc}
}

// CreateRoute establishes a new ingress mapping
// @Summary Create a new gateway route
// @Description Registers a new path pattern for the API gateway to proxy to a backend
// @Tags gateway
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param request body CreateRouteRequest true "Create route request"
// @Success 201 {object} domain.GatewayRoute
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /gateway/routes [post]
func (h *GatewayHandler) CreateRoute(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid request body"))
		return
	}

	if req.RateLimit == 0 {
		req.RateLimit = 100
	}

	params := ports.CreateRouteParams{
		Name:                 req.Name,
		Pattern:              req.PathPrefix,
		Target:               req.TargetURL,
		Methods:              req.Methods,
		StripPrefix:          req.StripPrefix,
		RateLimit:            req.RateLimit,
		DialTimeout:          req.DialTimeout,
		ResponseHeaderTimeout: req.ResponseHeaderTimeout,
		IdleConnTimeout:      req.IdleConnTimeout,
		TLSSkipVerify:        req.TLSSkipVerify,
		RequireTLS:          req.RequireTLS,
		AllowedCIDRs:         req.AllowedCIDRs,
		BlockedCIDRs:         req.BlockedCIDRs,
		MaxBodySize:          req.MaxBodySize,
		Priority:             req.Priority,
	}

	route, err := h.svc.CreateRoute(c.Request.Context(), params)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, route)
}

// ListRoutes returns all gateway routes
// @Summary List all gateway routes
// @Description Gets a list of all registered API gateway routes
// @Tags gateway
// @Produce json
// @Security APIKeyAuth
// @Success 200 {array} domain.GatewayRoute
// @Failure 500 {object} httputil.Response
// @Router /gateway/routes [get]
func (h *GatewayHandler) ListRoutes(c *gin.Context) {
	routes, err := h.svc.ListRoutes(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, routes)
}

// DeleteRoute removes a gateway route
// @Summary Delete a gateway route
// @Description Removes an existing API gateway route by ID
// @Tags gateway
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Route ID"
// @Success 200 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /gateway/routes/{id} [delete]
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

	proxy, route, params, ok := h.svc.GetProxy(c.Request.Method, path)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "No route found for " + path})
		return
	}

	// Apply IP allowlist/denylist (nil route means no route-specific rules apply)
	if route != nil && !h.checkCIDR(c, route) {
		return
	}

	// Apply request size limit
	if route != nil && route.MaxBodySize > 0 {
		c.Request.Body = &limitedReader{
			ReadCloser: c.Request.Body,
			limit:      route.MaxBodySize,
		}
	}

	// Inject parameters into request context for downstream services if needed
	if len(params) > 0 {
		for k, v := range params {
			c.Set("path_param_"+k, v)
		}
	}

	// Inject trace headers
	h.injectTraceHeaders(c)

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *GatewayHandler) injectTraceHeaders(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Request.Header.Set("X-Request-ID", requestID)
	c.Header("X-Request-ID", requestID)

	// W3C TraceContext
	traceID := generateTraceID()
	spanID := generateSpanID()
	c.Request.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", traceID, spanID))
	c.Request.Header.Set("tracestate", "")
}

func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *GatewayHandler) checkCIDR(c *gin.Context, route *domain.GatewayRoute) bool {
	clientIP := net.ParseIP(c.ClientIP())
	if clientIP == nil {
		return true // Allow if we can't parse IP
	}

	// Check blocked CIDRs first (takes precedence)
	for _, cidrStr := range route.BlockedCIDRs {
		if _, ipNet, err := net.ParseCIDR(cidrStr); err == nil && ipNet.Contains(clientIP) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return false
		}
	}

	// If allowlist is non-empty, only allow matched IPs
	if len(route.AllowedCIDRs) > 0 {
		allowed := false
		for _, cidrStr := range route.AllowedCIDRs {
			if _, ipNet, err := net.ParseCIDR(cidrStr); err == nil && ipNet.Contains(clientIP) {
				allowed = true
				break
			}
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return false
		}
	}

	return true
}

// limitedReader wraps an io.ReadCloser and enforces a byte limit.
type limitedReader struct {
	io.ReadCloser
	limit int64
	read  int64
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if l.read >= l.limit {
		return 0, io.EOF
	}
	toRead := l.limit - l.read
	if int64(len(p)) > toRead {
		p = p[:toRead]
	}
	n, err = l.ReadCloser.Read(p)
	l.read += int64(n)
	if l.read >= l.limit && err == nil {
		err = io.EOF
	}
	return
}
