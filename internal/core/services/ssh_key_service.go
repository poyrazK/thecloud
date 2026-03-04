package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"golang.org/x/crypto/ssh"
)

type SSHKeyService struct {
	repo    ports.SSHKeyRepository
	rbacSvc ports.RBACService
}

func NewSSHKeyService(repo ports.SSHKeyRepository, rbacSvc ports.RBACService) *SSHKeyService {
	return &SSHKeyService{
		repo:    repo,
		rbacSvc: rbacSvc,
	}
}

func (s *SSHKeyService) CreateKey(ctx context.Context, name, publicKey string) (*domain.SSHKey, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	userID := appcontext.UserIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSSHKeyCreate); err != nil {
		return nil, err
	}

	// Validate Key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("invalid public key format: %v", err))
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)

	// Check if key with same name exists in tenant
	_, err = s.repo.GetByName(ctx, tenantID, name)
	if err == nil {
		return nil, errors.New(errors.Conflict, "ssh key with this name already exists")
	}

	// If error is NOT NotFound, return it
	if !isNotFound(err) {
		return nil, err
	}

	key := &domain.SSHKey{
		ID:          uuid.New(),
		UserID:      userID,
		TenantID:    tenantID,
		Name:        name,
		PublicKey:   strings.TrimSpace(publicKey),
		Fingerprint: fingerprint,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *SSHKeyService) GetKey(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSSHKeyRead); err != nil {
		return nil, err
	}

	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if key.TenantID != tenantID {
		return nil, errors.New(errors.NotFound, "ssh key not found")
	}
	return key, nil
}

func (s *SSHKeyService) ListKeys(ctx context.Context) ([]*domain.SSHKey, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSSHKeyRead); err != nil {
		return nil, err
	}

	return s.repo.List(ctx, tenantID)
}

func (s *SSHKeyService) DeleteKey(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionSSHKeyDelete); err != nil {
		return err
	}

	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if key.TenantID != tenantID {
		return errors.New(errors.NotFound, "ssh key not found")
	}

	return s.repo.Delete(ctx, key.ID)
}

func isNotFound(err error) bool {
	// Helper to check for NotFound error using internal/errors
	return errors.Is(err, errors.NotFound)
}
