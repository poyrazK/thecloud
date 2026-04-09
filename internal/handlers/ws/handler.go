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

// Handler handles WebSocket connections.
type Handler struct {
	hub            *Hub
	identitySvc    ports.IdentityService
	logger         *slog.Logger
	allowedOrigins string
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, identitySvc ports.IdentityService, logger *slog.Logger, allowedOrigins ...string) *Handler {
	origins := ""
	if len(allowedOrigins) > 0 {
		origins = allowedOrigins[0]
	}
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
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			apiKey = authHeader
		}
	} else {
		// Fallback to query param (deprecated)
		apiKey = c.Query("api_key")
		if apiKey != "" {
			h.logger.Warn("websocket auth via query string is deprecated")
		}
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "api_key required"})
		return
	}

	apiKeyObj, err := h.identitySvc.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api_key"})
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
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Same-origin requests have no Origin header
			}
			if h.allowedOrigins == "" {
				return false // No origins allowed if not configured
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

