package node

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MemberState struct {
	Address   string
	Status    string
	LastSeen  time.Time
	Heartbeat uint64
}

type GossipProtocol struct {
	nodeID   string
	address  string
	members  map[string]*MemberState
	mu       sync.RWMutex
	stopCh   chan struct{}
	logger   *slog.Logger
	dialOpts []grpc.DialOption
	peers    map[string]pb.StorageNodeClient
}

func NewGossipProtocol(nodeID, address string, logger *slog.Logger) *GossipProtocol {
	g := &GossipProtocol{
		nodeID:   nodeID,
		address:  address,
		members:  make(map[string]*MemberState),
		stopCh:   make(chan struct{}),
		logger:   logger,
		peers:    make(map[string]pb.StorageNodeClient),
		dialOpts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	}
	// Add self
	g.members[nodeID] = &MemberState{
		Address:   address,
		Status:    "alive",
		LastSeen:  time.Now(),
		Heartbeat: 0,
	}
	return g
}

func (g *GossipProtocol) AddPeer(id, addr string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.members[id]; !ok {
		g.members[id] = &MemberState{
			Address:   addr,
			Status:    "alive",
			LastSeen:  time.Now(),
			Heartbeat: 0,
		}
	}
}

func (g *GossipProtocol) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	failTicker := time.NewTicker(2 * time.Second) // Check failures often
	go func() {
		for {
			select {
			case <-ticker.C:
				g.gossip()
			case <-failTicker.C:
				g.detectFailures()
			case <-g.stopCh:
				ticker.Stop()
				failTicker.Stop()
				return
			}
		}
	}()
}

func (g *GossipProtocol) detectFailures() {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	timeout := 5 * time.Second

	for id, m := range g.members {
		if id == g.nodeID {
			continue
		}

		if m.Status == "alive" && now.Sub(m.LastSeen) > timeout {
			m.Status = "suspect"
			g.logger.Warn("node flagged as suspect", "id", id, "last_seen", m.LastSeen)
		} else if m.Status == "suspect" && now.Sub(m.LastSeen) > 3*timeout {
			m.Status = "dead"
			g.logger.Error("node flagged as dead", "id", id, "last_seen", m.LastSeen)
			// Reconfiguration is handled asynchronously by the Coordinator
			// seeing the updated status via GetClusterStatus.
		}
	}
}

func (g *GossipProtocol) Stop() {
	close(g.stopCh)
}

func (g *GossipProtocol) gossip() {
	g.mu.Lock()
	// Increment own heartbeat
	me := g.members[g.nodeID]
	me.Heartbeat++
	me.LastSeen = time.Now()

	// Prepare message
	msg := &pb.GossipMessage{
		SenderId:   g.nodeID,
		SenderAddr: g.address,
		Timestamp:  time.Now().Unix(),
		Members:    make(map[string]*pb.MemberState),
	}

	// Convert members to proto and select random peer
	var peers []string
	for id, m := range g.members {
		msg.Members[id] = &pb.MemberState{
			Addr:      m.Address,
			Status:    m.Status,
			LastSeen:  m.LastSeen.Unix(),
			Heartbeat: m.Heartbeat,
		}
		if id != g.nodeID && m.Status == "alive" {
			peers = append(peers, id)
		}
	}
	g.mu.Unlock()

	if len(peers) == 0 {
		return
	}

	// Pick random peer
	targetID := peers[rand.Intn(len(peers))]
	g.sendGossip(targetID, msg)
}

func (g *GossipProtocol) sendGossip(targetID string, msg *pb.GossipMessage) {
	g.mu.RLock()
	targetAddr := g.members[targetID].Address
	client, ok := g.peers[targetID]
	g.mu.RUnlock()

	if !ok {
		conn, err := grpc.NewClient(targetAddr, g.dialOpts...)
		if err != nil {
			g.logger.Error("failed to connect to peer", "peer", targetID, "error", err)
			return
		}
		client = pb.NewStorageNodeClient(conn)
		g.mu.Lock()
		g.peers[targetID] = client
		g.mu.Unlock()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Gossip(ctx, msg)
	if err != nil {
		g.logger.Warn("gossip failed", "target", targetID, "error", err)
		// Mark as suspect if needed (Phase 2 enhancement)
	}
}

func (g *GossipProtocol) OnGossip(msg *pb.GossipMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for id, remoteState := range msg.Members {
		localState, exists := g.members[id]
		if !exists {
			// New member discovered
			g.members[id] = &MemberState{
				Address:   remoteState.Addr,
				Status:    remoteState.Status,
				LastSeen:  time.Now(),
				Heartbeat: remoteState.Heartbeat,
			}
			g.logger.Info("discovered new member", "id", id, "addr", remoteState.Addr)
			continue
		}

		if remoteState.Heartbeat > localState.Heartbeat {
			localState.Heartbeat = remoteState.Heartbeat
			localState.LastSeen = time.Now()
			localState.Status = remoteState.Status
		}
	}
}
