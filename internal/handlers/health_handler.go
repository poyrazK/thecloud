// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	svc ports.HealthService
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(svc ports.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// Live checks if the process is running
// @Summary Liveness check
// @Tags health
// @Success 200 {object} map[string]string
// @Router /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready checks if dependencies are connected
// @Summary Readiness check
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	result := h.svc.Check(c.Request.Context())

	status := http.StatusOK
	if result.Status == "DEGRADED" || result.Status == "DOWN" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, result)
}
