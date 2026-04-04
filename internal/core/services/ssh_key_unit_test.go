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
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc, err := services.NewSSHKeyService(services.SSHKeyServiceParams{
		Repo:    mockRepo,
		RBACSvc: rbacSvc,
	})
	require.NoError(t, err)

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

	t.Run("GetKey_Success", func(t *testing.T) {
		id := uuid.New()
		expected := &domain.SSHKey{ID: id, TenantID: tenantID}
		mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil).Once()

		res, err := svc.GetKey(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("GetKey_NotFound", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		_, err := svc.GetKey(ctx, id)
		require.Error(t, err)
	})

	t.Run("ListKeys_Success", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, tenantID).Return([]*domain.SSHKey{{ID: uuid.New()}}, nil).Once()

		res, err := svc.ListKeys(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("DeleteKey_Success", func(t *testing.T) {
		id := uuid.New()
		key := &domain.SSHKey{ID: id, TenantID: tenantID}
		mockRepo.On("GetByID", mock.Anything, id).Return(key, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()

		err := svc.DeleteKey(ctx, id)
		require.NoError(t, err)
	})
}
