package httphandlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
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
// @Summary Create IAM Policy
// @Description Create a new granular IAM policy with specified statements.
// @Tags iam
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param policy body domain.Policy true "Policy configuration"
// @Success 201 {object} domain.Policy
// @Failure 400 {object} httputil.Response
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /iam/policies [post]
func (h *IAMHandler) CreatePolicy(c *gin.Context) {
	var policy domain.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.validatePolicy(&policy); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	if err := h.svc.CreatePolicy(c.Request.Context(), &policy); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, policy)
}

// GetPolicyByID returns a specific IAM policy.
// @Summary Get IAM Policy
// @Description Get details of a specific IAM policy by its ID.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 {object} domain.Policy
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /iam/policies/{id} [get]
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
// @Summary List IAM Policies
// @Description List all IAM policies for the tenant.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Success 200 {array} domain.Policy
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /iam/policies [get]
func (h *IAMHandler) ListPolicies(c *gin.Context) {
	policies, err := h.svc.ListPolicies(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, policies)
}

// UpdatePolicy updates an existing IAM policy.
// @Summary Update IAM Policy
// @Description Update the configuration of an existing IAM policy.
// @Tags iam
// @Security APIKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Policy ID"
// @Param policy body domain.Policy true "Updated policy configuration"
// @Success 200 {object} domain.Policy
// @Failure 400 {object} httputil.Response
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /iam/policies/{id} [put]
func (h *IAMHandler) UpdatePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var policy domain.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		httputil.Error(c, err)
		return
	}
	policy.ID = id

	if err := h.validatePolicy(&policy); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	if err := h.svc.UpdatePolicy(c.Request.Context(), &policy); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, policy)
}

// DeletePolicy removes an IAM policy.
// @Summary Delete IAM Policy
// @Description Delete an existing IAM policy by its ID.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /iam/policies/{id} [delete]
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
// @Summary Attach IAM Policy to User
// @Description Attach a specific IAM policy to a user.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Param userId path string true "User ID"
// @Param policyId path string true "Policy ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /iam/users/{userId}/policies/{policyId} [post]
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
// @Summary Detach IAM Policy from User
// @Description Remove a specific IAM policy assignment from a user.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Param userId path string true "User ID"
// @Param policyId path string true "Policy ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Router /iam/users/{userId}/policies/{policyId} [delete]
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
// @Summary List User IAM Policies
// @Description List all IAM policies currently attached to a specific user.
// @Tags iam
// @Security APIKeyAuth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} domain.Policy
// @Failure 401 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Router /iam/users/{userId}/policies [get]
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

func (h *IAMHandler) validatePolicy(policy *domain.Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}
	if len(policy.Statements) == 0 {
		return fmt.Errorf("at least one policy statement is required")
	}
	for i, stmt := range policy.Statements {
		if stmt.Effect != domain.EffectAllow && stmt.Effect != domain.EffectDeny {
			return fmt.Errorf("statement %d has invalid effect: must be 'Allow' or 'Deny'", i)
		}
		if len(stmt.Action) == 0 {
			return fmt.Errorf("statement %d must specify at least one action", i)
		}
		if len(stmt.Resource) == 0 {
			return fmt.Errorf("statement %d must specify at least one resource", i)
		}
	}
	return nil
}
