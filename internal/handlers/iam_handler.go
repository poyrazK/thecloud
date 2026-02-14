package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// IAMHandler handles HTTP requests for identity and access management.
type IAMHandler struct {
	svc ports.IAMService
}

// NewIAMHandler creates a new IAM handler.
func NewIAMHandler(svc ports.IAMService) *IAMHandler {
	return &IAMHandler{svc: svc}
}

// CreatePolicy creates a new IAM policy.
func (h *IAMHandler) CreatePolicy(c *gin.Context) {
	var policy domain.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.CreatePolicy(c.Request.Context(), &policy); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, policy)
}

// GetPolicyByID returns a specific IAM policy.
func (h *IAMHandler) GetPolicyByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	policy, err := h.svc.GetPolicyByID(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, policy)
}

// ListPolicies returns all IAM policies.
func (h *IAMHandler) ListPolicies(c *gin.Context) {
	policies, err := h.svc.ListPolicies(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, policies)
}

// DeletePolicy removes an IAM policy.
func (h *IAMHandler) DeletePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DeletePolicy(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"status": "deleted"})
}

// AttachPolicyToUser attaches a policy to a user.
func (h *IAMHandler) AttachPolicyToUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.Error(c, err)
		return
	}
	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.AttachPolicyToUser(c.Request.Context(), userID, policyID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"status": "attached"})
}

// DetachPolicyFromUser removes a policy from a user.
func (h *IAMHandler) DetachPolicyFromUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.Error(c, err)
		return
	}
	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DetachPolicyFromUser(c.Request.Context(), userID, policyID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"status": "detached"})
}

// GetUserPolicies lists all policies attached to a specific user.
func (h *IAMHandler) GetUserPolicies(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	policies, err := h.svc.GetPoliciesForUser(c.Request.Context(), userID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, policies)
}
