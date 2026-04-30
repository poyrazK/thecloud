// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// FunctionScheduleHandler handles serverless function schedule HTTP endpoints.
type FunctionScheduleHandler struct {
	svc ports.FunctionScheduleService
}

// NewFunctionScheduleHandler constructs a FunctionScheduleHandler.
func NewFunctionScheduleHandler(svc ports.FunctionScheduleService) *FunctionScheduleHandler {
	return &FunctionScheduleHandler{svc: svc}
}

// CreateFunctionScheduleRequest is the payload for schedule creation.
type CreateFunctionScheduleRequest struct {
	FunctionID uuid.UUID       `json:"function_id" binding:"required"`
	Name       string          `json:"name" binding:"required"`
	Schedule   string          `json:"schedule" binding:"required"`
	Payload    json.RawMessage `json:"payload"`
}

// Create handles POST /function-schedules
func (h *FunctionScheduleHandler) Create(c *gin.Context) {
	var req CreateFunctionScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	sched, err := h.svc.CreateSchedule(c.Request.Context(), req.FunctionID, req.Name, req.Schedule, []byte(req.Payload))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, sched)
}

// List handles GET /function-schedules
func (h *FunctionScheduleHandler) List(c *gin.Context) {
	schedules, err := h.svc.ListSchedules(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, schedules)
}

// Get handles GET /function-schedules/:id
func (h *FunctionScheduleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid schedule id"))
		return
	}

	schedule, err := h.svc.GetSchedule(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, schedule)
}

// Delete handles DELETE /function-schedules/:id
func (h *FunctionScheduleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid schedule id"))
		return
	}

	if err := h.svc.DeleteSchedule(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "schedule deleted"})
}

// Pause handles POST /function-schedules/:id/pause
func (h *FunctionScheduleHandler) Pause(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid schedule id"))
		return
	}

	if err := h.svc.PauseSchedule(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "schedule paused"})
}

// Resume handles POST /function-schedules/:id/resume
func (h *FunctionScheduleHandler) Resume(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid schedule id"))
		return
	}

	if err := h.svc.ResumeSchedule(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "schedule resumed"})
}

// GetRuns handles GET /function-schedules/:id/runs
func (h *FunctionScheduleHandler) GetRuns(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid schedule id"))
		return
	}

	runs, err := h.svc.GetScheduleRuns(c.Request.Context(), id, 100)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, runs)
}
