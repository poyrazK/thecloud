// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// ElasticIPHandler handles HTTP requests for Elastic IPs.
type ElasticIPHandler struct {
	svc ports.ElasticIPService
}

// NewElasticIPHandler creates a new ElasticIPHandler.
func NewElasticIPHandler(svc ports.ElasticIPService) *ElasticIPHandler {
	return &ElasticIPHandler{svc: svc}
}

// Allocate reserves a new Elastic IP.
// @Summary Allocate an Elastic IP
// @Tags elastic-ips
// @Security APIKeyAuth
// @Produce json
// @Success 201 {object} domain.ElasticIP
// @Router /elastic-ips [post]
func (h *ElasticIPHandler) Allocate(c *gin.Context) {
	eip, err := h.svc.AllocateIP(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusCreated, eip)
}

// List returns all Elastic IPs for the tenant.
// @Summary List Elastic IPs
// @Tags elastic-ips
// @Security APIKeyAuth
// @Produce json
// @Success 200 {array} domain.ElasticIP
// @Router /elastic-ips [get]
func (h *ElasticIPHandler) List(c *gin.Context) {
	eips, err := h.svc.ListElasticIPs(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, eips)
}

// Get retrieves a specific Elastic IP.
// @Summary Get Elastic IP
// @Tags elastic-ips
// @Security APIKeyAuth
// @Param id path string true "EIP ID"
// @Produce json
// @Success 200 {object} domain.ElasticIP
// @Router /elastic-ips/{id} [get]
func (h *ElasticIPHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid elastic ip id"))
		return
	}

	eip, err := h.svc.GetElasticIP(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, eip)
}

// Release returns an Elastic IP to the pool.
// @Summary Release Elastic IP
// @Tags elastic-ips
// @Security APIKeyAuth
// @Param id path string true "EIP ID"
// @Success 200 {object} httputil.Response
// @Router /elastic-ips/{id} [delete]
func (h *ElasticIPHandler) Release(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid elastic ip id"))
		return
	}

	if err := h.svc.ReleaseIP(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, gin.H{"message": "elastic ip released"})
}

// Associate maps an Elastic IP to an instance.
// @Summary Associate Elastic IP
// @Tags elastic-ips
// @Security APIKeyAuth
// @Param id path string true "EIP ID"
// @Param request body object{instance_id=string} true "Association Request"
// @Produce json
// @Success 200 {object} domain.ElasticIP
// @Router /elastic-ips/{id}/associate [post]
func (h *ElasticIPHandler) Associate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid elastic ip id"))
		return
	}

	var req struct {
		InstanceID string `json:"instance_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid instance_id"))
		return
	}

	instID := uuid.MustParse(req.InstanceID)
	eip, err := h.svc.AssociateIP(c.Request.Context(), id, instID)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, eip)
}

// Disassociate removes an Elastic IP mapping.
// @Summary Disassociate Elastic IP
// @Tags elastic-ips
// @Security APIKeyAuth
// @Param id path string true "EIP ID"
// @Produce json
// @Success 200 {object} domain.ElasticIP
// @Router /elastic-ips/{id}/disassociate [post]
func (h *ElasticIPHandler) Disassociate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid elastic ip id"))
		return
	}

	eip, err := h.svc.DisassociateIP(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, eip)
}
