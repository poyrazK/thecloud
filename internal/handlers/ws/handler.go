// Package ws implements the WebSocket handlers and hub.
package ws

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	websocketReadBufferSize  = 1024
	websocketWriteBufferSize = 1024
)

// Handler handles WebSocket connections.
type Handler struct {
	hub            *Hub
	identitySvc    ports.IdentityService
	logger         *slog.Logger
	allowedOrigins string
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, identitySvc ports.IdentityService, logger *slog.Logger, allowedOrigins ...string) *Handler {
	origins := strings.Join(allowedOrigins, ",")
	return &Handler{
		hub:            hub,
		identitySvc:    identitySvc,
		logger:         logger,
		allowedOrigins: origins,
	}
}

// ServeWS upgrades HTTP to WebSocket and registers the client.
func (h *Handler) ServeWS(c *gin.Context) {
	// Validate API key - prefer Authorization header, fall back to query param
	var apiKey string
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		lowerHeader := strings.ToLower(authHeader)
		switch {
		case strings.HasPrefix(lowerHeader, "bearer "):
			apiKey = strings.TrimSpace(authHeader[len("Bearer "):])
		case strings.Contains(authHeader, " "):
			// Contains a space but not Bearer prefix - malformed
			h.logger.Warn("malformed authorization header received")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		default:
			// No space - treat as raw API key
			apiKey = authHeader
		}
	}

	if apiKey == "" {
		// Fallback to query param (deprecated)
		apiKey = c.Query("api_key")
		if apiKey != "" {
			h.logger.Warn("websocket auth via query string is deprecated")
		}
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "api key or authorization header required"})
		return
	}

	apiKeyObj, err := h.identitySvc.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader().Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", slog.String("error", err.Error()))
		return
	}

	client := NewClient(h.hub, conn, apiKeyObj.UserID.String(), apiKeyObj.TenantID, h.logger)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) upgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  websocketReadBufferSize,
		WriteBufferSize: websocketWriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Browser clients always send Origin header for WebSocket handshakes.
			// If no origins are configured, allow all (backward compatible).
			// If origins are configured, only allow matching origins.
			if h.allowedOrigins == "" {
				return true
			}
			if origin == "" {
				// Reject requests with no origin when allowlist is enforced
				return false
			}
			allowed := strings.Split(h.allowedOrigins, ",")
			for _, o := range allowed {
				if strings.TrimSpace(o) == origin {
					return true
				}
			}
			return false
		},
	}
}
