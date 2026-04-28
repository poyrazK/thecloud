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
	allowedOrigins []string
}

// NewHandler creates a new WebSocket handler.
//
// allowedOrigins is the explicit allowlist of origins permitted to open a
// WebSocket connection. The list is required: an empty configuration is
// treated as "deny all". This is a fail-closed default — see #249. Operators
// who really want to allow everything must opt in explicitly by passing "*"
// (which is itself only safe for non-credential-bearing endpoints).
func NewHandler(hub *Hub, identitySvc ports.IdentityService, logger *slog.Logger, allowedOrigins ...string) *Handler {
	cleaned := make([]string, 0, len(allowedOrigins))
	for _, raw := range allowedOrigins {
		for _, o := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
	}
	if len(cleaned) == 0 && logger != nil {
		logger.Warn("websocket handler started with empty allowed-origins list; all cross-origin upgrades will be rejected")
	}
	return &Handler{
		hub:            hub,
		identitySvc:    identitySvc,
		logger:         logger,
		allowedOrigins: cleaned,
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
		CheckOrigin:     h.checkOrigin,
	}
}

// checkOrigin enforces the configured allowlist. Defaults are fail-closed:
//
//   - No allowlist configured → reject everything.
//   - Empty Origin header → reject (browsers always send it for cross-origin
//     WebSocket upgrades; the only requests without one are non-browser clients
//     that should authenticate via the API rather than the websocket route).
//   - Wildcard "*" entry → allow any origin. Intended only for development /
//     internal endpoints; logged as a warning at handler construction.
//
// Matches are byte-exact on the full Origin string, so attackers cannot bypass
// the check via subdomain or port confusion.
func (h *Handler) checkOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 {
		if h.logger != nil {
			h.logger.Warn("rejecting websocket upgrade: no allowed origins configured",
				slog.String("remote", r.RemoteAddr))
		}
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		if h.logger != nil {
			h.logger.Debug("rejecting websocket upgrade: missing Origin header",
				slog.String("remote", r.RemoteAddr))
		}
		return false
	}
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	if h.logger != nil {
		h.logger.Debug("rejecting websocket upgrade: origin not in allowlist",
			slog.String("origin", origin),
			slog.String("remote", r.RemoteAddr))
	}
	return false
}
