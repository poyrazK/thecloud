// Package ws provides WebSocket handler components.
package ws

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client represents a WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	tenantID uuid.UUID
	logger   *slog.Logger
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, userID string, tenantID uuid.UUID, logger *slog.Logger) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		tenantID: tenantID,
		logger:   logger,
	}
}

// ReadPump handles incoming messages from the client.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		if err := c.conn.Close(); err != nil {
			c.logger.Error("failed to close connection", slog.String("error", err.Error()))
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Error("failed to set read deadline", slog.String("error", err.Error()))
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error("failed to set pong read deadline", slog.String("error", err.Error()))
		}
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("websocket error", slog.String("error", err.Error()))
			}
			break
		}
		// Currently read-only; future: handle client commands
	}
}

// WritePump sends messages to the client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.logger.Error("failed to set write deadline for ping", slog.String("error", err.Error()))
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
