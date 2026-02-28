// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type rbacService struct {
	userRepo  ports.UserRepository
	roleRepo  ports.RoleRepository
	iamRepo   ports.IAMRepository
	evaluator ports.PolicyEvaluator
	logger    *slog.Logger
}

// NewRBACService constructs an RBAC service for role-based authorization.
func NewRBACService(userRepo ports.UserRepository, roleRepo ports.RoleRepository, iamRepo ports.IAMRepository, evaluator ports.PolicyEvaluator, logger *slog.Logger) *rbacService {
	return &rbacService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		iamRepo:   iamRepo,
		evaluator: evaluator,
		logger:    logger,
	}
}

func (s *rbacService) Authorize(ctx context.Context, userID uuid.UUID, permission domain.Permission, resource string) error {
	allowed, err := s.HasPermission(ctx, userID, permission, resource)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New(errors.Forbidden, fmt.Sprintf("permission denied: %s on %s", permission, resource))
	}
	return nil
}

func (s *rbacService) HasPermission(ctx context.Context, userID uuid.UUID, permission domain.Permission, resource string) (bool, error) {
	// 1. Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("IAM: failed to get user", "user_id", userID, "error", err)
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	s.logger.Debug("IAM: checking permission", "user_id", userID, "user_role", user.Role, "permission", permission, "resource", resource)

	// 2. Check Attached IAM Policies (New Logic)
	policies, err := s.iamRepo.GetPoliciesForUser(ctx, user.TenantID, userID)
	if err != nil {
		s.logger.Error("IAM: failed to get policies for user", "user_id", userID, "error", err)
		return false, err // Fail closed on error
	}

	if len(policies) > 0 {
		allowed, evalErr := s.evaluator.Evaluate(ctx, policies, string(permission), resource, nil)
		if evalErr != nil {
			s.logger.Error("IAM: policy evaluation error", "user_id", userID, "error", evalErr)
			return false, evalErr // Fail closed on error
		}
		if allowed {
			return true, nil
		}
		// If policies exist and evaluation resulted in Deny (explicit or implicit),
		// we do NOT fall through to legacy roles. Deny always wins.
		return false, nil
	}

	// 3. Fallback to Role-based logic (Legacy/Compatibility)
	role, err := s.roleRepo.GetRoleByName(ctx, user.Role)
	if err != nil {
		s.logger.Debug("IAM: role not found in DB, using default permissions", "role", user.Role, "error", err)
		return s.hasDefaultPermission(user.Role, permission)
	}

	s.logger.Debug("IAM: found role in DB", "role", role.Name, "permissions_count", len(role.Permissions))

	// 4. Check permissions
	for _, p := range role.Permissions {
		if p == domain.PermissionFullAccess {
			return true, nil
		}
		if p == permission {
			return true, nil
		}
	}

	s.logger.Warn("IAM: permission denied (role in DB but permission not listed)", "role", role.Name, "permission", permission)
	return false, nil
}

func (s *rbacService) CreateRole(ctx context.Context, role *domain.Role) error {
	// Check if role name already exists to provide a better error than 500
	existing, _ := s.roleRepo.GetRoleByName(ctx, role.Name)
	if existing != nil {
		return errors.New(errors.Conflict, fmt.Sprintf("role with name '%s' already exists", role.Name))
	}

	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	return s.roleRepo.CreateRole(ctx, role)
}

func (s *rbacService) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetRoleByID(ctx, id)
}

func (s *rbacService) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	return s.roleRepo.GetRoleByName(ctx, name)
}

func (s *rbacService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.ListRoles(ctx)
}

func (s *rbacService) UpdateRole(ctx context.Context, role *domain.Role) error {
	return s.roleRepo.UpdateRole(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.roleRepo.DeleteRole(ctx, id)
}

func (s *rbacService) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	return s.roleRepo.AddPermissionToRole(ctx, roleID, permission)
}

func (s *rbacService) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	return s.roleRepo.RemovePermissionFromRole(ctx, roleID, permission)
}

func (s *rbacService) BindRole(ctx context.Context, userIdentifier string, roleName string) error {
	// 1. Verify role exists
	_, err := s.roleRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("role %s does not exist: %w", roleName, err)
	}

	// 2. Get user (by ID or Email)
	var user *domain.User
	userID, err := uuid.Parse(userIdentifier)
	if err == nil {
		user, err = s.userRepo.GetByID(ctx, userID)
	} else {
		user, err = s.userRepo.GetByEmail(ctx, userIdentifier)
	}

	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 3. Update role
	user.Role = roleName
	return s.userRepo.Update(ctx, user)
}

func (s *rbacService) ListRoleBindings(ctx context.Context) ([]*domain.User, error) {
	// In this implementation, bindings are just users with their roles
	return s.userRepo.List(ctx)
}

func (s *rbacService) EvaluatePolicy(ctx context.Context, userID uuid.UUID, action string, resource string, evalCtx map[string]interface{}) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	policies, err := s.iamRepo.GetPoliciesForUser(ctx, user.TenantID, userID)
	if err != nil {
		return false, err
	}
	if len(policies) == 0 {
		return false, nil
	}
	return s.evaluator.Evaluate(ctx, policies, action, resource, evalCtx)
}

func (s *rbacService) hasDefaultPermission(roleName string, permission domain.Permission) (bool, error) {
	// Fallback to default roles if not found in DB
	s.logger.Debug("RBAC: checking default permission", "role", roleName, "permission", permission)

	switch roleName {
	case domain.RoleAdmin:
		s.logger.Debug("RBAC: admin role, granting permission", "permission", permission)
		return true, nil
	case domain.RoleViewer:
		if permission == domain.PermissionInstanceRead ||
			permission == domain.PermissionVolumeRead ||
			permission == domain.PermissionVpcRead {
			s.logger.Debug("RBAC: viewer role, granting read permission", "permission", permission)
			return true, nil
		}
	case domain.RoleDeveloper:
		// Developer gets most things except RBAC management
		if permission != domain.PermissionFullAccess {
			s.logger.Debug("RBAC: developer role, granting permission", "permission", permission)
			return true, nil
		}
	}

	s.logger.Warn("role not found in DB and no default fallback", "role", roleName, "permission", permission)
	return false, nil
}
