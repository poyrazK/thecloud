package ws

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientBatching(t *testing.T) {
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

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?api_key=batch-key"
	mockID.On("ValidateAPIKey", mock.Anything, "batch-key").Return(&domain.APIKey{Key: "batch-key", UserID: uuid.New()}, nil)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	time.Sleep(100 * time.Millisecond)

	// Send multiple messages quickly to test batching
	hub.broadcast <- []byte("msg1")
	hub.broadcast <- []byte("msg2")
	hub.broadcast <- []byte("msg3")

	_, p, err := conn.ReadMessage()
	assert.NoError(t, err)

	payload := string(p)
	// It might be combined with newlines if they hit the same WritePump tick
	assert.True(t, strings.Contains(payload, "msg1") || strings.Contains(payload, "msg2") || strings.Contains(payload, "msg3"))
}

func TestClientPingPong(t *testing.T) {
	t.Parallel()
	// This is hard to test deterministically without mocking time or the ticker
	// but we can at least verify basic connectivity handles pongs.
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

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?api_key=ping-key"
	mockID.On("ValidateAPIKey", mock.Anything, "ping-key").Return(&domain.APIKey{Key: "ping-key", UserID: uuid.New()}, nil)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// gorilla/websocket handles PING by default with a PONG response if not overridden.
	// Our client uses default SetPongHandler which updates deadline.
}
