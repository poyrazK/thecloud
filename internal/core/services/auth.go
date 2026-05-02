// Package services implements the core business logic.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxFailedAttempts  = 5
	defaultLockout     = 15 * time.Minute
	maxLockoutMapSize  = 10000  // prevents unbounded map growth DoS
	maxFailedMapSize   = 10000  // prevents unbounded map growth DoS
)

// AuthService handles registration and authentication workflows.
type AuthService struct {
	userRepo        ports.UserRepository
	apiKeySvc       ports.IdentityService
	auditSvc        ports.AuditService
	tenantSvc       ports.TenantService
	logger          *slog.Logger
	failedAttempts  map[string]int
	lockouts        map[string]time.Time
	lockoutDuration time.Duration
	mu              sync.Mutex
}

// NewAuthService constructs an AuthService with its dependencies.
func NewAuthService(userRepo ports.UserRepository, apiKeySvc ports.IdentityService, auditSvc ports.AuditService, tenantSvc ports.TenantService, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		apiKeySvc:       apiKeySvc,
		auditSvc:        auditSvc,
		tenantSvc:       tenantSvc,
		logger:          logger,
		failedAttempts:  make(map[string]int),
		lockouts:        make(map[string]time.Time),
		lockoutDuration: defaultLockout,
	}
}

// SetLockoutDuration overrides the default lockout duration. Useful for testing.
func (s *AuthService) SetLockoutDuration(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lockoutDuration = d
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "Register")
	defer span.End()
	span.SetAttributes(attribute.String("user.email", email))
	// Validate password strength
	if err := passwordvalidator.Validate(password, 50); err != nil {
		return nil, errors.New(errors.InvalidInput, "password is too weak: "+err.Error())
	}

	// Check if user already exists
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New(errors.Conflict, "user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         domain.RoleDeveloper,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Transactionality would be better here, but avoiding for simplicity unless needed
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create Personal Tenant
	tenantName := fmt.Sprintf("%s's Personal Tenant", name)

	// Simple slugify: lowercase, replace spaces with hyphens, keep only alphanumeric
	slugName := strings.ToLower(name)
	slugName = strings.ReplaceAll(slugName, " ", "-")

	// Remove non-alphanumeric chars
	var cleanSlug strings.Builder
	for _, r := range slugName {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanSlug.WriteRune(r)
		}
	}
	slugName = cleanSlug.String()

	if slugName == "" {
		slugName = "personal"
	}
	tenantSlug := fmt.Sprintf("personal-%s-%s", slugName, user.ID.String()[:8])

	tenant, err := s.tenantSvc.CreateTenant(appcontext.WithInternalCall(ctx), tenantName, tenantSlug, user.ID)
	if err != nil {
		rollbackErr := s.userRepo.Delete(ctx, user.ID)
		if rollbackErr != nil {
			return nil, fmt.Errorf("failed to create personal tenant: %w; rollback failed: %w", err, rollbackErr)
		}
		return nil, fmt.Errorf("failed to create personal tenant: %w", err)
	}
	ctx = appcontext.WithTenantID(ctx, tenant.ID)

	// Reload user to reflect changes made during tenant creation (e.g. DefaultTenantID)
	updatedUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err == nil {
		user = updatedUser
	}

	if err := s.auditSvc.Log(ctx, user.ID, "user.register", "user", user.ID.String(), map[string]interface{}{
		"email": email,
	}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "user.register", "user_id", user.ID, "error", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "Login")
	defer span.End()
	span.SetAttributes(attribute.String("user.email", email))
	s.mu.Lock()
	if lockoutTime, ok := s.lockouts[email]; ok {
		if time.Now().Before(lockoutTime) {
			s.mu.Unlock()
			return nil, "", errors.New(errors.Unauthorized, "account is locked due to too many failed attempts")
		}
		delete(s.lockouts, email)
		delete(s.failedAttempts, email)
	}
	s.mu.Unlock()

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		s.incrementFailure(email)
		platform.AuthAttemptsTotal.WithLabelValues("failure_not_found").Inc()
		return nil, "", errors.New(errors.Unauthorized, "invalid email or password")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.incrementFailure(email)
		return nil, "", errors.New(errors.Unauthorized, "invalid email or password")
	}

	// Clear failures on success
	s.mu.Lock()
	delete(s.failedAttempts, email)
	s.mu.Unlock()

	if user.DefaultTenantID != nil {
		ctx = appcontext.WithTenantID(ctx, *user.DefaultTenantID)
	}
	// or just return a fresh one. In a real platform, login gives you a JWT and
	// you manage API keys separately.
	// For now, let's create a default key for them.
	key, err := s.apiKeySvc.CreateKey(ctx, user.ID, "Default Key")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create initial API key: %w", err)
	}

	if err := s.auditSvc.Log(ctx, user.ID, "user.login", "user", user.ID.String(), map[string]interface{}{}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "user.login", "user_id", user.ID, "error", err)
	}

	platform.AuthAttemptsTotal.WithLabelValues("success").Inc()

	return user, key.Key, nil
}

func (s *AuthService) incrementFailure(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failedAttempts[email]++
	platform.AuthAttemptsTotal.WithLabelValues("failure_incorrect_password").Inc()
	if s.failedAttempts[email] >= maxFailedAttempts {
		s.lockouts[email] = time.Now().Add(s.lockoutDuration)
		platform.AuthAttemptsTotal.WithLabelValues("failure_lockout").Inc()
	}
	// Deterministic size-based eviction to prevent unbounded map growth DoS.
	if len(s.lockouts) > maxLockoutMapSize || len(s.failedAttempts) > maxFailedMapSize {
		s.purgeExpiredLocked()
	}
	// Probabilistic purge as secondary cleanup.
	if len(s.lockouts) > 0 && time.Now().Nanosecond()%10 == 0 {
		s.purgeExpiredLocked()
	}
}

// purgeExpiredLocked removes expired lockouts and stale failure records.
// Caller must hold s.mu.
func (s *AuthService) purgeExpiredLocked() {
	now := time.Now()
	for email, lockoutTime := range s.lockouts {
		if now.After(lockoutTime) {
			delete(s.lockouts, email)
			delete(s.failedAttempts, email)
		}
	}
	// Also purge failure-only entries that have no lockout and are very old
	// (these are users who failed but didn't reach lockout threshold)
	for email, count := range s.failedAttempts {
		// If count is suspiciously high, purge to prevent unbounded growth
		if count > maxFailedAttempts*10 {
			delete(s.failedAttempts, email)
		}
	}
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			return nil, errors.New(errors.NotFound, "user not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to fetch user", err)
	}
	return user, nil
}
