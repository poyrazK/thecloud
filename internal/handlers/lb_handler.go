package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/ports"
	"github.com/poyraz/cloud/internal/errors"
	"github.com/poyraz/cloud/pkg/httputil"
)

type LBHandler struct {
	svc ports.LBService
}

func NewLBHandler(svc ports.LBService) *LBHandler {
	return &LBHandler{svc: svc}
}

type CreateLBRequest struct {
	Name      string `json:"name" binding:"required"`
	VpcID     string `json:"vpc_id" binding:"required"`
	Port      int    `json:"port" binding:"required"`
	Algorithm string `json:"algorithm"`
}

type AddTargetRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	Weight     int    `json:"weight"`
}

func (h *LBHandler) Create(c *gin.Context) {
	var req CreateLBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	vpcID, err := uuid.Parse(req.VpcID)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid vpc_id format"))
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	lb, err := h.svc.Create(c.Request.Context(), req.Name, vpcID, req.Port, req.Algorithm, idempotencyKey)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, lb)
}

func (h *LBHandler) List(c *gin.Context) {
	lbs, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, lbs)
}

func (h *LBHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id format"))
		return
	}

	lb, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, lb)
}

func (h *LBHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid id format"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, gin.H{"message": "load balancer deletion initiated"})
}

func (h *LBHandler) AddTarget(c *gin.Context) {
	lbIDStr := c.Param("id")
	lbID, err := uuid.Parse(lbIDStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid lb_id format"))
		return
	}

	var req AddTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	instID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid instance_id format"))
		return
	}

	if err := h.svc.AddTarget(c.Request.Context(), lbID, instID, req.Port, req.Weight); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, gin.H{"message": "target added"})
}

func (h *LBHandler) RemoveTarget(c *gin.Context) {
	lbIDStr := c.Param("id")
	lbID, err := uuid.Parse(lbIDStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid lb_id format"))
		return
	}

	instIDStr := c.Param("instanceId")
	instID, err := uuid.Parse(instIDStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid instance_id format"))
		return
	}

	if err := h.svc.RemoveTarget(c.Request.Context(), lbID, instID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "target removed"})
}

func (h *LBHandler) ListTargets(c *gin.Context) {
	lbIDStr := c.Param("id")
	lbID, err := uuid.Parse(lbIDStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid lb_id format"))
		return
	}

	targets, err := h.svc.ListTargets(c.Request.Context(), lbID)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, targets)
}
