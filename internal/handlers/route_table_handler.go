// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// RouteTableHandler handles HTTP requests for route tables.
type RouteTableHandler struct {
	svc ports.RouteTableService
}

// NewRouteTableHandler creates a new RouteTableHandler.
func NewRouteTableHandler(svc ports.RouteTableService) *RouteTableHandler {
	return &RouteTableHandler{svc: svc}
}

// CreateRouteTableRequest represents the body for creating a route table.
type CreateRouteTableRequest struct {
	VPCID  string `json:"vpc_id" binding:"required,uuid"`
	Name   string `json:"name" binding:"required"`
	IsMain bool   `json:"is_main"`
}

// Create creates a new custom route table.
// @Summary Create Route Table
// @Tags route-tables
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param request body CreateRouteTableRequest true "Route Table Request"
// @Success 201 {object} domain.RouteTable
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables [post]
func (h *RouteTableHandler) Create(c *gin.Context) {
	var req CreateRouteTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	vpcID, _ := uuid.Parse(req.VPCID)
	rt, err := h.svc.CreateRouteTable(c.Request.Context(), vpcID, req.Name, req.IsMain)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusCreated, rt)
}

// List returns all route tables for a VPC.
// @Summary List Route Tables
// @Tags route-tables
// @Security APIKeyAuth
// @Produce json
// @Param vpc_id query string false "VPC ID to filter by"
// @Success 200 {array} domain.RouteTable
// @Failure 500 {object} httputil.Response
// @Router /route-tables [get]
func (h *RouteTableHandler) List(c *gin.Context) {
	vpcIDStr := c.Query("vpc_id")
	if vpcIDStr == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, "vpc_id is required"))
		return
	}

	vpcID, _ := uuid.Parse(vpcIDStr)
	rts, err := h.svc.ListRouteTables(c.Request.Context(), vpcID)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, rts)
}

// Get retrieves a specific route table.
// @Summary Get Route Table
// @Tags route-tables
// @Security APIKeyAuth
// @Produce json
// @Param id path string true "Route Table ID"
// @Success 200 {object} domain.RouteTable
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id} [get]
func (h *RouteTableHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	rt, err := h.svc.GetRouteTable(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, rt)
}

// Delete removes a custom route table.
// @Summary Delete Route Table
// @Tags route-tables
// @Security APIKeyAuth
// @Param id path string true "Route Table ID"
// @Success 204
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id} [delete]
func (h *RouteTableHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	if err := h.svc.DeleteRouteTable(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusNoContent, nil)
}

// AddRouteRequest represents the body for adding a route.
type AddRouteRequest struct {
	DestinationCIDR string `json:"destination_cidr" binding:"required"`
	TargetType      string `json:"target_type" binding:"required"`
	TargetID        string `json:"target_id,omitempty"`
}

// AddRoute adds a route to a route table.
// @Summary Add Route
// @Tags route-tables
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Route Table ID"
// @Param request body AddRouteRequest true "Route Request"
// @Success 201 {object} domain.Route
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id}/routes [post]
func (h *RouteTableHandler) AddRoute(c *gin.Context) {
	idStr := c.Param("id")
	rtID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	var req AddRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	var targetID *uuid.UUID
	if req.TargetID != "" {
		id, _ := uuid.Parse(req.TargetID)
		targetID = &id
	}

	route, err := h.svc.AddRoute(c.Request.Context(), rtID, req.DestinationCIDR, domain.RouteTargetType(req.TargetType), targetID)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusCreated, route)
}

// RemoveRoute removes a route from a route table.
// @Summary Remove Route
// @Tags route-tables
// @Security APIKeyAuth
// @Param id path string true "Route Table ID"
// @Param route_id query string true "Route ID"
// @Success 204
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id}/routes/{route_id} [delete]
func (h *RouteTableHandler) RemoveRoute(c *gin.Context) {
	rtIDStr := c.Param("id")
	rtID, err := uuid.Parse(rtIDStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	routeIDStr := c.Query("route_id")
	if routeIDStr == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, "route_id is required"))
		return
	}
	routeID, _ := uuid.Parse(routeIDStr)

	if err := h.svc.RemoveRoute(c.Request.Context(), rtID, routeID); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusNoContent, nil)
}

// AssociateSubnetRequest represents the body for associating a subnet.
type AssociateSubnetRequest struct {
	SubnetID string `json:"subnet_id" binding:"required,uuid"`
}

// AssociateSubnet links a subnet to a route table.
// @Summary Associate Subnet
// @Tags route-tables
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Route Table ID"
// @Param request body AssociateSubnetRequest true "Subnet Association Request"
// @Success 200
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id}/associate [post]
func (h *RouteTableHandler) AssociateSubnet(c *gin.Context) {
	idStr := c.Param("id")
	rtID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	var req AssociateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	subnetID, _ := uuid.Parse(req.SubnetID)
	if err := h.svc.AssociateSubnet(c.Request.Context(), rtID, subnetID); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, nil)
}

// DisassociateSubnet removes a subnet's association.
// @Summary Disassociate Subnet
// @Tags route-tables
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Route Table ID"
// @Param request body AssociateSubnetRequest true "Subnet Disassociation Request"
// @Success 200
// @Failure 400 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Router /route-tables/{id}/disassociate [post]
func (h *RouteTableHandler) DisassociateSubnet(c *gin.Context) {
	idStr := c.Param("id")
	rtID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid route table id"))
		return
	}

	var req AssociateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	subnetID, _ := uuid.Parse(req.SubnetID)
	if err := h.svc.DisassociateSubnet(c.Request.Context(), rtID, subnetID); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, nil)
}