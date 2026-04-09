// Package ws provides WebSocket handler components.
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/platform"
)

// Hub maintains active WebSocket connections and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *slog.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			platform.WSConnectionsActive.Inc()
			h.logger.Debug("client connected", slog.Int("total", len(h.clients)))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			platform.WSConnectionsActive.Dec()
			h.logger.Debug("client disconnected", slog.Int("total", len(h.clients)))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent sends a WSEvent to all connected clients.
func (h *Hub) BroadcastEvent(event *domain.WSEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", slog.String("error", err.Error()))
		return
	}
	h.broadcast <- data
}

// BroadcastEventToTenant sends a WSEvent only to clients matching the given tenant.
// If userID is not nil, further filter to that specific user.
// Collects clients to remove under RLock, then removes them via unregister channel.
func (h *Hub) BroadcastEventToTenant(ctx context.Context, event *domain.WSEvent, tenantID uuid.UUID, userID *uuid.UUID) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Collect clients to notify under read lock
	h.mu.RLock()
	var toRemove []*Client
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		if userID != nil && client.userID != userID.String() {
			continue
		}
		select {
		case client.send <- data:
		default:
			// Client buffer full - mark for removal
			toRemove = append(toRemove, client)
		}
	}
	h.mu.RUnlock()

	// Remove dead clients via unregister channel (safe removal)
	for _, client := range toRemove {
		select {
		case h.unregister <- client:
		case <-ctx.Done():
			// Context cancelled - client will be cleaned up when hub shuts down
			return ctx.Err()
		}
	}

	return nil
}

// PublishEvent implements ports.RealtimePublisher.
func (h *Hub) PublishEvent(ctx context.Context, event *domain.WSEvent, tenantID uuid.UUID, userID *uuid.UUID) error {
	return h.BroadcastEventToTenant(ctx, event, tenantID, userID)
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
