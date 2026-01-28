// Package httphandlers exposes HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// LifecycleHandler handles lifecycle rule endpoints.
type LifecycleHandler struct {
	svc ports.LifecycleService
}

// NewLifecycleHandler constructs a LifecycleHandler.
func NewLifecycleHandler(svc ports.LifecycleService) *LifecycleHandler {
	return &LifecycleHandler{svc: svc}
}

// CreateRule adds a new lifecycle rule to the bucket
// @Summary Create lifecycle rule
// @Description Adds an expiration rule for objects with a specific prefix
// @Tags lifecycle
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param request body domain.LifecycleRule true "Lifecycle rule"
// @Success 201 {object} domain.LifecycleRule
// @Failure 400 {object} httputil.Response
// @Router /storage/buckets/{bucket}/lifecycle [post]
func (h *LifecycleHandler) CreateRule(c *gin.Context) {
	bucket := c.Param("bucket")
	var req struct {
		Prefix         string `json:"prefix"`
		ExpirationDays int    `json:"expiration_days" binding:"required"`
		Enabled        bool   `json:"enabled"`
	}
	// Default enabled to true if not provided is tricky with bool unmarshal (defaults to false).
	// Let's assume user must specify or we handle logic.
	// Actually struct tag `default:"true"` isn't standard in gin binding without validator.
	// Logic: default Enabled to true?
	// The problem is `false` is zero value.
	// Let's use pointer for Enabled or just default logic.

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	// Default Enable to true? No, let's respect the input. If they send nothing, it's false.
	// Users usually want to enable it immediately.
	// But let's assume if they don't provide it, it's false (safe).
	// Actually implementation plans said "Enabled: true" default in SQL.
	// I'll stick to input.

	rule, err := h.svc.CreateRule(c.Request.Context(), bucket, req.Prefix, req.ExpirationDays, req.Enabled)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, rule)
}

// ListRules lists all lifecycle rules for a bucket
// @Summary List lifecycle rules
// @Description Lists all lifecycle rules configured for a bucket
// @Tags lifecycle
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Success 200 {array} domain.LifecycleRule
// @Router /storage/buckets/{bucket}/lifecycle [get]
func (h *LifecycleHandler) ListRules(c *gin.Context) {
	bucket := c.Param("bucket")
	rules, err := h.svc.ListRules(c.Request.Context(), bucket)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, rules)
}

// DeleteRule removes a lifecycle rule
// @Summary Delete lifecycle rule
// @Description Deletes a specific lifecycle rule
// @Tags lifecycle
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param id path string true "Rule ID"
// @Success 204
// @Router /storage/buckets/{bucket}/lifecycle/{id} [delete]
func (h *LifecycleHandler) DeleteRule(c *gin.Context) {
	bucket := c.Param("bucket")
	id := c.Param("id")

	if err := h.svc.DeleteRule(c.Request.Context(), bucket, id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}
