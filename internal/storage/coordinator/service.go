package coordinator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"math/rand"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
)

// Coordinator implements ports.FileStore to manage distributed storage.
type Coordinator struct {
	ring         *ConsistentHashRing
	clients      map[string]pb.StorageNodeClient
	replicaCount int
	writeQuorum  int
	stopCh       chan struct{}
}

// NewCoordinator creates a new distributed storage coordinator.
func NewCoordinator(ring *ConsistentHashRing, clients map[string]pb.StorageNodeClient, replicaCount int) *Coordinator {
	if replicaCount < 1 {
		replicaCount = 1
	}
	c := &Coordinator{
		ring:         ring,
		clients:      clients,
		replicaCount: replicaCount,
		writeQuorum:  (replicaCount / 2) + 1,
		stopCh:       make(chan struct{}),
	}
	go c.startSyncLoop()
	return c
}

func (c *Coordinator) startSyncLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.SyncClusterState()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Coordinator) SyncClusterState() {
	// Pick random node to query
	var client pb.StorageNodeClient
	for _, cl := range c.clients {
		client = cl
		if rand.Float32() < 0.5 {
			break
		}
	}

	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.GetClusterStatus(ctx, &pb.Empty{})
	if err != nil {
		// Try another one next time
		return
	}

	// Update Ring based on status
	for id, m := range resp.Members {
		if m.Status == "dead" {
			c.ring.RemoveNode(id)
			// Idea: If we implement dynamic client pool, we would remove from c.clients too.
		} else if m.Status == "alive" {
			// Ensure it's in the ring.
			// Currently our AddNode just appends, so we don't want to add if already there.
			// Since we don't have HasNode efficient check exposed, let's assume static membership for now
			// except for removing dead nodes.
			// Getting this right requires better Ring implementation (idempotent Add).
		}
	}
}

func (c *Coordinator) Stop() {
	close(c.stopCh)
}

// Write saves data to the cluster with replication.
func (c *Coordinator) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	// 1. Read all data (Phase 1 simplification)
	b, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	size := int64(len(b))

	// 2. Get target nodes
	nodes := c.ring.GetNodes(bucket+"/"+key, c.replicaCount)
	if len(nodes) == 0 {
		return 0, fmt.Errorf("no storage nodes available")
	}

	// 3. Parallel Write
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	var lastErr error

	for _, nodeID := range nodes {
		client, ok := c.clients[nodeID]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(id string, cl pb.StorageNodeClient) {
			defer wg.Done()
			_, err := cl.Store(ctx, &pb.StoreRequest{Bucket: bucket, Key: key, Data: b})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				lastErr = err
			} else {
				successCount++
			}
		}(nodeID, client)
	}
	wg.Wait()

	// 4. Check Quorum
	if successCount < c.writeQuorum {
		return 0, fmt.Errorf("write quorum failed (%d/%d): %v", successCount, c.writeQuorum, lastErr)
	}

	return size, nil
}

// Read retrieves data from the cluster.
func (c *Coordinator) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	nodes := c.ring.GetNodes(bucket+"/"+key, c.replicaCount)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no storage nodes available")
	}

	var lastErr error
	for _, nodeID := range nodes {
		client, ok := c.clients[nodeID]
		if !ok {
			continue
		}

		resp, err := client.Retrieve(ctx, &pb.RetrieveRequest{Bucket: bucket, Key: key})
		if err != nil {
			lastErr = err
			continue
		}
		if !resp.Found {
			continue
		}

		// Found it
		return io.NopCloser(bytes.NewReader(resp.Data)), nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("object not found")
}

// Delete removes data from the cluster.
func (c *Coordinator) Delete(ctx context.Context, bucket, key string) error {
	nodes := c.ring.GetNodes(bucket+"/"+key, c.replicaCount)

	// Best effort delete from all replicas
	// We don't necessarily fail if one is down, but we should report if all fail.

	successCount := 0
	for _, nodeID := range nodes {
		client, ok := c.clients[nodeID]
		if !ok {
			continue
		}

		_, err := client.Delete(ctx, &pb.DeleteRequest{Bucket: bucket, Key: key})
		if err == nil {
			successCount++
		}
	}

	if successCount == 0 && len(nodes) > 0 {
		return fmt.Errorf("failed to delete from any node")
	}

	return nil
}
