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

const errInvalidEIPID = "invalid elastic ip id"

// ElasticIPHandler handles HTTP requests for Elastic IPs.
type ElasticIPHandler struct {
	svc ports.ElasticIPService
}

// AssociateIPRequest represents the body for associating an EIP.
type AssociateIPRequest struct {
	InstanceID string `json:"instance_id" validate:"required" binding:"required,uuid"`
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
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
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
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
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
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /elastic-ips/{id} [get]
func (h *ElasticIPHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidEIPID))
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
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /elastic-ips/{id} [delete]
func (h *ElasticIPHandler) Release(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidEIPID))
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
// @Param request body AssociateIPRequest true "Association Request"
// @Accept json
// @Produce json
// @Success 200 {object} domain.ElasticIP
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /elastic-ips/{id}/associate [post]
func (h *ElasticIPHandler) Associate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidEIPID))
		return
	}

	var req AssociateIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid instance_id"))
		return
	}

	instID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid instance_id format"))
		return
	}

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
// @Failure 400 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /elastic-ips/{id}/disassociate [post]
func (h *ElasticIPHandler) Disassociate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidEIPID))
		return
	}

	eip, err := h.svc.DisassociateIP(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, eip)
}
