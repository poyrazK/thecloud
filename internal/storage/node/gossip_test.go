package node

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"github.com/stretchr/testify/assert"
)

func TestGossipProtocol_AddPeer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	g := NewGossipProtocol("node1", "localhost:8080", logger)

	g.AddPeer("node2", "localhost:8081")

	g.mu.RLock()
	defer g.mu.RUnlock()
	assert.Contains(t, g.members, "node2")
	assert.Equal(t, "localhost:8081", g.members["node2"].Address)
}

func TestGossipProtocol_OnGossip(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	g := NewGossipProtocol("node1", "localhost:8080", logger)

	// Update coming from node2 about node3
	msg := &pb.GossipMessage{
		SenderId: "node2",
		Members: map[string]*pb.MemberState{
			"node3": {
				Addr:      "localhost:8082",
				Status:    "alive",
				Heartbeat: 10,
			},
		},
	}

	g.OnGossip(msg)

	g.mu.RLock()
	node3, exists := g.members["node3"]
	g.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, "localhost:8082", node3.Address)
	assert.Equal(t, uint64(10), node3.Heartbeat)

	// Newer heartbeat
	msg2 := &pb.GossipMessage{
		Members: map[string]*pb.MemberState{
			"node3": {
				Addr:      "localhost:8082",
				Status:    "alive",
				Heartbeat: 15,
			},
		},
	}
	g.OnGossip(msg2)

	g.mu.RLock()
	assert.Equal(t, uint64(15), g.members["node3"].Heartbeat)
	g.mu.RUnlock()
}

func TestGossipProtocol_DetectFailures(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	g := NewGossipProtocol("node1", "localhost:8080", logger)

	// Add a node that was seen long ago
	g.members["node2"] = &MemberState{
		Address:  "localhost:8081",
		Status:   "alive",
		LastSeen: time.Now().Add(-10 * time.Second), // Timeout is 5s
	}

	g.detectFailures()

	g.mu.RLock()
	assert.Equal(t, "suspect", g.members["node2"].Status)
	g.mu.RUnlock()

	// Advance time more to make it dead ( > 3*timeout = 15s)
	g.members["node2"].LastSeen = time.Now().Add(-20 * time.Second)
	g.detectFailures()

	g.mu.RLock()
	assert.Equal(t, "dead", g.members["node2"].Status)
	g.mu.RUnlock()
}
