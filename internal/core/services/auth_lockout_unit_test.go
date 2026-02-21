package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_LockoutLogic(t *testing.T) {
	userRepo := new(MockUserRepo)
	apiKeySvc := new(MockIdentityService)
	auditSvc := new(MockAuditService)
	tenantSvc := new(MockTenantService)
	
	svc := services.NewAuthService(userRepo, apiKeySvc, auditSvc, tenantSvc)
	ctx := context.Background()
	email := "locked@example.com"
	password := "CorrectPassword123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{ID: uuid.New(), Email: email, PasswordHash: string(hashed)}

	t.Run("Lockout after 5 attempts", func(t *testing.T) {
		userRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)
		
		// 5 failed attempts
		for i := 0; i < 5; i++ {
			_, _, err := svc.Login(ctx, email, "wrong-password")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid email or password")
		}

		// 6th attempt should be locked out
		_, _, err := svc.Login(ctx, email, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account is locked")
	})

	t.Run("Success clears failures", func(t *testing.T) {
		email2 := "clean@example.com"
		user2 := &domain.User{ID: uuid.New(), Email: email2, PasswordHash: string(hashed)}
		userRepo.On("GetByEmail", mock.Anything, email2).Return(user2, nil)
		apiKeySvc.On("CreateKey", mock.Anything, user2.ID, mock.Anything).Return(&domain.APIKey{Key: "k1"}, nil).Once()
		auditSvc.On("Log", mock.Anything, user2.ID, "user.login", "user", user2.ID.String(), mock.Anything).Return(nil).Once()

		// 2 failed attempts
		for i := 0; i < 2; i++ {
			_, _, err := svc.Login(ctx, email2, "wrong")
			assert.Error(t, err)
		}

		// Success
		_, _, err := svc.Login(ctx, email2, password)
		assert.NoError(t, err)
	})
}
