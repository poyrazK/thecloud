package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	sshKeysPath = "/ssh-keys"
)

type mockSSHKeyService struct {
	mock.Mock
}

func (m *mockSSHKeyService) CreateKey(ctx context.Context, name, publicKey string) (*domain.SSHKey, error) {
	args := m.Called(ctx, name, publicKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SSHKey), args.Error(1)
}

func (m *mockSSHKeyService) GetKey(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SSHKey), args.Error(1)
}

func (m *mockSSHKeyService) ListKeys(ctx context.Context) ([]*domain.SSHKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SSHKey), args.Error(1)
}

func (m *mockSSHKeyService) DeleteKey(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func setupSSHKeyHandlerTest() (*mockSSHKeyService, *SSHKeyHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockSSHKeyService)
	handler := NewSSHKeyHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestSSHKeyHandlerCreate(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupSSHKeyHandlerTest()

	r.POST(sshKeysPath, handler.Create)

	reqBody := CreateSSHKeyRequest{
		Name:      "test-key",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
	}
	body, _ := json.Marshal(reqBody)

	expectedKey := &domain.SSHKey{
		ID:        uuid.New(),
		Name:      reqBody.Name,
		PublicKey: reqBody.PublicKey,
		CreatedAt: time.Now(),
	}

	svc.On("CreateKey", mock.Anything, reqBody.Name, reqBody.PublicKey).Return(expectedKey, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", sshKeysPath, bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.SSHKey
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKey.Name, response.Name)
}

func TestSSHKeyHandlerList(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupSSHKeyHandlerTest()

	r.GET(sshKeysPath, handler.List)

	expectedKeys := []*domain.SSHKey{
		{
			ID:   uuid.New(),
			Name: "key-1",
		},
		{
			ID:   uuid.New(),
			Name: "key-2",
		},
	}

	svc.On("ListKeys", mock.Anything).Return(expectedKeys, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", sshKeysPath, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*domain.SSHKey
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestSSHKeyHandlerGet(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupSSHKeyHandlerTest()

	r.GET(sshKeysPath+"/:id", handler.Get)

	id := uuid.New()
	expectedKey := &domain.SSHKey{
		ID:   id,
		Name: "test-key",
	}

	svc.On("GetKey", mock.Anything, id).Return(expectedKey, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", sshKeysPath+"/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.SSHKey
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKey.ID, response.ID)
}

func TestSSHKeyHandlerDelete(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupSSHKeyHandlerTest()

	r.DELETE(sshKeysPath+"/:id", handler.Delete)

	id := uuid.New()

	svc.On("DeleteKey", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", sshKeysPath+"/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
