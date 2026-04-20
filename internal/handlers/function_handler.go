// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

const invalidFunctionIDMsg = "invalid function id"

// EnvVarRequest represents a single environment variable in requests.
type EnvVarRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

// FunctionUpdateRequest mirrors domain.FunctionUpdate for JSON binding.
type FunctionUpdateRequest struct {
	Handler  *string           `json:"handler,omitempty"`
	Timeout  *int              `json:"timeout,omitempty"`
	MemoryMB *int              `json:"memory_mb,omitempty"`
	Status   *string           `json:"status,omitempty"`
	EnvVars  []*EnvVarRequest `json:"env_vars,omitempty"`
}

// FunctionHandler handles serverless function HTTP endpoints.
type FunctionHandler struct {
	svc ports.FunctionService
}

// NewFunctionHandler constructs a FunctionHandler.
func NewFunctionHandler(svc ports.FunctionService) *FunctionHandler {
	return &FunctionHandler{svc: svc}
}

// CreateFunctionRequest is the payload for function creation.
type CreateFunctionRequest struct {
	Name    string `form:"name" binding:"required"`
	Runtime string `form:"runtime" binding:"required"`
	Handler string `form:"handler" binding:"required"`
}

func (h *FunctionHandler) Create(c *gin.Context) {
	var req CreateFunctionRequest
	if err := c.ShouldBind(&req); err != nil {
		httputil.Error(c, errors.Wrap(errors.InvalidInput, "invalid request", err))
		return
	}

	file, err := c.FormFile("code")
	if err != nil {
		httputil.Error(c, errors.Wrap(errors.InvalidInput, "code file is required", err))
		return
	}

	f, err := file.Open()
	if err != nil {
		httputil.Error(c, errors.Wrap(errors.Internal, "failed to open code file", err))
		return
	}
	defer func() { _ = f.Close() }()

	code, err := io.ReadAll(f)
	if err != nil {
		httputil.Error(c, errors.Wrap(errors.Internal, "failed to read code file", err))
		return
	}

	function, err := h.svc.CreateFunction(c.Request.Context(), req.Name, req.Runtime, req.Handler, code)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, function)
}

func (h *FunctionHandler) List(c *gin.Context) {
	functions, err := h.svc.ListFunctions(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, functions)
}

func (h *FunctionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidFunctionIDMsg))
		return
	}

	function, err := h.svc.GetFunction(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, function)
}

func (h *FunctionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidFunctionIDMsg))
		return
	}

	if err := h.svc.DeleteFunction(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, gin.H{"message": "function deleted"})
}

func (h *FunctionHandler) Invoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidFunctionIDMsg))
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httputil.Error(c, errors.Wrap(errors.InvalidInput, "failed to read payload", err))
		return
	}

	async := c.Query("async") == "true"

	invocation, err := h.svc.InvokeFunction(c.Request.Context(), id, payload, async)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	status := http.StatusOK
	if async {
		status = http.StatusAccepted
	}
	httputil.Success(c, status, invocation)
}

func (h *FunctionHandler) GetLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidFunctionIDMsg))
		return
	}

	logs, err := h.svc.GetFunctionLogs(c.Request.Context(), id, 100)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, logs)
}

// Update updates an existing function's configuration.
// @Summary Update function
// @Description Updates a function's timeout, memory, handler, status, or environment variables
// @Tags functions
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Function ID"
// @Param request body FunctionUpdateRequest true "Update request"
// @Success 200 {object} domain.Function
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /functions/{id} [patch]
func (h *FunctionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidFunctionIDMsg))
		return
	}

	var req FunctionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.Wrap(errors.InvalidInput, "invalid request body", err))
		return
	}

	update := &domain.FunctionUpdate{
		Handler:  req.Handler,
		Timeout:  req.Timeout,
		MemoryMB: req.MemoryMB,
		Status:   req.Status,
	}
	if req.EnvVars != nil {
		envVars := make([]*domain.EnvVar, len(req.EnvVars))
		for i, e := range req.EnvVars {
			envVars[i] = &domain.EnvVar{Key: e.Key, Value: e.Value}
		}
		update.EnvVars = envVars
	}

	fn, err := h.svc.UpdateFunction(c.Request.Context(), id, update)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, fn)
}
