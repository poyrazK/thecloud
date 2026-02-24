package services_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSSHKeyService_Unit(t *testing.T) {
	mockRepo := new(MockSSHKeyRepo)
	svc := services.NewSSHKeyService(mockRepo)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)
	ctx = appcontext.WithUserID(ctx, userID)

	// Generate a valid RSA public key for testing
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicRsaKey, _ := ssh.NewPublicKey(&privateKey.PublicKey)
	pubKey := string(ssh.MarshalAuthorizedKey(publicRsaKey))

	t.Run("CreateKey_Success", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, tenantID, "test-key").Return(nil, errors.New(errors.NotFound, "not found")).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		key, err := svc.CreateKey(ctx, "test-key", pubKey)
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, "test-key", key.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateKey_Duplicate", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, tenantID, "test-key").Return(&domain.SSHKey{}, nil).Once()

		key, err := svc.CreateKey(ctx, "test-key", pubKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		assert.Nil(t, key)
	})
}
