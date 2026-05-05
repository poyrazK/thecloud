// Package services implements core business workflows.
package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
)

// serverSecret is used as HMAC key to prevent rainbow table attacks on API key hashes.
var serverSecret = getServerSecret()

func getServerSecret() string {
	secret := platform.GetSecretsEncryptionKey()
	if secret != "" {
		return secret
	}
	// For tests, allow a fallback to avoid os.Exit in package init
	if os.Getenv("TEST_SECRETS") != "" {
		return os.Getenv("TEST_SECRETS")
	}
	slog.Default().Error("SECRETS_ENCRYPTION_KEY environment variable is required")
	os.Exit(1)
	return "" // unreachable but satisfies compiler
}

// computeKeyHash creates a HMAC-SHA256 hash of the API key using the server secret.
// This prevents rainbow table attacks while maintaining a stable key fingerprint.
// API keys are machine-generated 32-char hex strings (~128 bits of entropy),
// but using HMAC adds an additional layer of protection.
func computeKeyHash(key string) string {
	//nolint:codeql // HMAC-SHA256 is used for key fingerprinting, not password hashing.
	h := hmac.New(sha256.New, []byte(serverSecret))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// IdentityServiceParams defines the dependencies for IdentityService.
type IdentityServiceParams struct {
	Repo     ports.IdentityRepository
	RbacSvc  ports.RBACService
	AuditSvc ports.AuditService
	Logger   *slog.Logger
}

// IdentityService manages API key lifecycle and validation.
type IdentityService struct {
	repo     ports.IdentityRepository
	rbacSvc  ports.RBACService
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// NewIdentityService constructs an IdentityService with its dependencies.
func NewIdentityService(params IdentityServiceParams) *IdentityService {
	logger := params.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &IdentityService{
		repo:     params.Repo,
		rbacSvc:  params.RbacSvc,
		auditSvc: params.AuditSvc,
		logger:   logger,
	}
}

func (s *IdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Allow if:
	// 1. Context has no user (initial login flow where we create first key)
	// 2. User is creating for themselves
	// 3. User is authorized via RBAC
	if uID != uuid.Nil && uID != userID {
		if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionIdentityCreate, "*"); err != nil {
			return nil, err
		}
	}

	// Generate a secure random key
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate secure key", err)
	}
	keyStr := "thecloud_" + hex.EncodeToString(b)

	apiKey := &domain.APIKey{
		ID:              uuid.New(),
		UserID:          userID,
		Key:             keyStr,
		KeyHash:         computeKeyHash(keyStr),
		Name:            name,
		CreatedAt:       time.Now(),
		TenantID:        tenantID,
		DefaultTenantID: nil,
	}
	if tenantID != uuid.Nil {
		apiKey.DefaultTenantID = &tenantID
	}

	if err := s.repo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	// Log audit event
	if err := s.auditSvc.Log(ctx, userID, "api_key.create", "api_key", apiKey.ID.String(), map[string]interface{}{
		"name": name,
	}); err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to log audit event", "action", "api_key.create", "resource_id", apiKey.ID.String(), "error", err)
		}
	}

	return apiKey, nil
}

func (s *IdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	// Authentication path, no RBAC check yet as we are resolving identity
	keyHash := computeKeyHash(key)
	apiKey, err := s.repo.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, errors.New(errors.Unauthorized, "invalid api key")
	}

	// Check expiration if set
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New(errors.Unauthorized, "api key has expired")
	}

	platform.AuthAttemptsTotal.WithLabelValues("success_api_key").Inc()

	return apiKey, nil
}

func (s *IdentityService) GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	return s.repo.GetAPIKeyByID(ctx, id)
}

func (s *IdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionIdentityRead, "*"); err != nil {
		return nil, err
	}

	// Horizontal access check: if requesting keys for another user, need elevated permission
	if userID != uID {
		if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionIdentityReadAll, "*"); err != nil {
			return nil, errors.New(errors.Forbidden, "cannot list keys for another user")
		}
	}

	return s.repo.ListAPIKeysByUserID(ctx, userID)
}

func (s *IdentityService) RevokeKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// 1. Authorize: Check if they have specific permission for this resource, OR it is self-revoking
	isSelf := uID == userID
	authErr := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionIdentityDelete, id.String())
	if authErr != nil && !isSelf {
		return authErr
	}

	key, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Ownership check: If RBAC failed but it was self-revoking, verify they own the key
	// If RBAC succeeded, we bypass this.
	if authErr != nil && isSelf && key.UserID != uID {
		return errors.New(errors.Forbidden, "unauthorized access to api key")
	}

	if err := s.repo.DeleteAPIKey(ctx, id); err != nil {
		return err
	}

	// Log audit event - capture error
	if auditErr := s.auditSvc.Log(ctx, uID, "api_key.revoke", "api_key", id.String(), map[string]interface{}{
		"name": key.Name,
	}); auditErr != nil {
		s.logger.Error("failed to log audit event for api key revocation", "user_id", uID, "key_id", id, "error", auditErr)
	}

	return nil
}

func (s *IdentityService) RotateKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.APIKey, error) {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// 1. Authorize: Check if they have specific permission for this resource, OR it is self-rotating
	isSelf := uID == userID
	authErr := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionIdentityDelete, id.String())
	if authErr != nil && !isSelf {
		return nil, authErr
	}

	key, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Ownership check: If RBAC failed but it was self-rotating, verify they own the key
	if authErr != nil && isSelf && key.UserID != uID {
		return nil, errors.New(errors.Forbidden, "unauthorized access to api key")
	}

	// 3. Create new key with same name
	newKey, err := s.CreateKey(ctx, userID, key.Name)
	if err != nil {
		return nil, err
	}

	// 4. Delete old key with compensating rollback
	if err := s.repo.DeleteAPIKey(ctx, id); err != nil {
		s.logger.Error("failed to delete old api key during rotation, rolling back new key", "id", id, "error", err)
		// Rollback: delete the newly created key
		if rbErr := s.repo.DeleteAPIKey(ctx, newKey.ID); rbErr != nil {
			s.logger.Error("failed to rollback new api key after rotation failure", "new_key_id", newKey.ID, "error", rbErr)
			return nil, errors.Wrap(errors.Internal, "rotation failed and rollback failed", err)
		}
		return nil, errors.Wrap(errors.Internal, "failed to delete old key, rotation rolled back", err)
	}

	// Log audit event - capture error
	if auditErr := s.auditSvc.Log(ctx, uID, "api_key.rotate", "api_key", id.String(), map[string]interface{}{
		"name":   key.Name,
		"new_id": newKey.ID.String(),
	}); auditErr != nil {
		s.logger.Error("failed to log audit event for api key rotation", "user_id", uID, "key_id", id, "error", auditErr)
	}

	return newKey, nil
}
