package ws

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockIdentityService struct {
	mock.Mock
}

func (m *mockIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.APIKey)
	return r0, args.Error(1)
}

func (m *mockIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.APIKey)
	return r0, args.Error(1)
}
func (m *mockIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, userID)
	r0, _ := args.Get(0).([]*domain.APIKey)
	return r0, args.Error(1)
}
func (m *mockIdentityService) RevokeKey(ctx context.Context, userID, id uuid.UUID) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}
func (m *mockIdentityService) RotateKey(ctx context.Context, userID, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, id)
	r0, _ := args.Get(0).(*domain.APIKey)
	return r0, args.Error(1)
}

func TestWebSocketLifecycle(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	mockID := new(mockIdentityService)
	handler := NewHandler(hub, mockID, logger)

	r := gin.New()
	r.GET("/ws", handler.ServeWS)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?api_key=valid-key"
	mockID.On("ValidateAPIKey", mock.Anything, "valid-key").Return(&domain.APIKey{Key: "valid-key", UserID: uuid.New()}, nil)

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	defer func() { _ = conn.Close() }()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	event := &domain.WSEvent{Type: domain.WSEventInstanceCreated, Timestamp: time.Now()}
	hub.BroadcastEvent(event)

	_, p, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(p), "INSTANCE_CREATED")

	_ = conn.Close()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestWebSocketAuthFailure(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	mockID := new(mockIdentityService)
	handler := NewHandler(hub, mockID, logger)

	r := gin.New()
	r.GET("/ws", handler.ServeWS)

	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("Missing API Key", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
		dialer := websocket.Dialer{}
		_, resp, err := dialer.Dial(wsURL, nil)
		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			_ = resp.Body.Close()
		}
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?api_key=invalid"
		mockID.On("ValidateAPIKey", mock.Anything, "invalid").Return(nil, errors.New(errors.Unauthorized, "invalid key"))
		dialer := websocket.Dialer{}
		_, resp, err := dialer.Dial(wsURL, nil)
		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			_ = resp.Body.Close()
		}
	})
}
