package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSecretRepo struct {
	mock.Mock
}

func (m *MockSecretRepo) Create(ctx context.Context, secret *domain.Secret) error {
	return m.Called(ctx, secret).Error(0)
}
func (m *MockSecretRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Secret)
	return r0, args.Error(1)
}
func (m *MockSecretRepo) GetByName(ctx context.Context, name string) (*domain.Secret, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Secret)
	return r0, args.Error(1)
}
func (m *MockSecretRepo) List(ctx context.Context) ([]*domain.Secret, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Secret)
	return r0, args.Error(1)
}
func (m *MockSecretRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockSecretRepo) Update(ctx context.Context, secret *domain.Secret) error {
	return m.Called(ctx, secret).Error(0)
}

func TestSecretService_Unit(t *testing.T) {
	mockRepo := new(MockSecretRepo)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	// Initialize service with a test master key for deterministic encryption
	svc := services.NewSecretService(mockRepo, mockEventSvc, mockAuditSvc, slog.Default(), "test-master-key-12345678", "test")

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateSecret", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Secret) bool {
			// Verify name and ensure encrypted value is populated
			return s.Name == "api-key" && s.EncryptedValue != ""
		})).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "SECRET_CREATE", mock.Anything, "SECRET", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "secret.create", "secret", mock.Anything, mock.Anything).Return(nil).Once()

		sec, err := svc.CreateSecret(ctx, "api-key", "my-secret", "test secret")
		assert.NoError(t, err)
		assert.NotNil(t, sec)
		assert.Equal(t, "api-key", sec.Name)
	})

	t.Run("GetSecret", func(t *testing.T) {
		id := uuid.New()
		// Generate valid encrypted data using the service helper to ensure decryption succeeds
		realEncrypted, _ := svc.Encrypt(ctx, userID, "decrypted")
		
		secret := &domain.Secret{ID: id, UserID: userID, EncryptedValue: realEncrypted, Name: "s1"}
		
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()
		// Update call expected for LastAccessedAt timestamp
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once() 
		mockEventSvc.On("RecordEvent", mock.Anything, "SECRET_ACCESS", mock.Anything, "SECRET", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "secret.access", "secret", id.String(), mock.Anything).Return(nil).Once()

		res, err := svc.GetSecret(ctx, id)
		assert.NoError(t, err)
		// Verify that the service correctly decrypts and returns the plaintext value
		assert.Equal(t, "decrypted", res.EncryptedValue)
	})

	t.Run("ListSecrets", func(t *testing.T) {
		secrets := []*domain.Secret{{Name: "s1"}, {Name: "s2"}}
		mockRepo.On("List", mock.Anything).Return(secrets, nil).Once()
		
		res, err := svc.ListSecrets(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		// Ensure values are redacted in list view for security
		assert.Equal(t, "[REDACTED]", res[0].EncryptedValue)
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		id := uuid.New()
		secret := &domain.Secret{ID: id, UserID: userID, Name: "s1"}
		
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "SECRET_DELETE", id.String(), "SECRET", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "secret.delete", "secret", id.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteSecret(ctx, id)
		assert.NoError(t, err)
	})
}
