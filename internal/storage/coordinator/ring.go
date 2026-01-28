// Package coordinator manages distributed storage coordination.
package coordinator

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHashRing maps keys to storage nodes using consistent hashing.
type ConsistentHashRing struct {
	ring         []uint32
	nodes        map[uint32]string
	virtualNodes int
	mu           sync.RWMutex
}

// NewConsistentHashRing constructs a ring with the given virtual node count.
func NewConsistentHashRing(virtualNodes int) *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes:        make(map[uint32]string),
		virtualNodes: virtualNodes,
	}
}

func (r *ConsistentHashRing) AddNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.virtualNodes; i++ {
		key := nodeID + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(key))
		r.nodes[hash] = nodeID
		r.ring = append(r.ring, hash)
	}
	sort.Slice(r.ring, func(i, j int) bool { return r.ring[i] < r.ring[j] })
}

func (r *ConsistentHashRing) RemoveNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	newRing := []uint32{}
	for _, hash := range r.ring {
		if r.nodes[hash] != nodeID {
			newRing = append(newRing, hash)
		} else {
			delete(r.nodes, hash)
		}
	}
	r.ring = newRing
}

func (r *ConsistentHashRing) GetNodes(key string, count int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ring) == 0 {
		return nil
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(r.ring), func(i int) bool { return r.ring[i] >= hash })

	if idx == len(r.ring) {
		idx = 0
	}

	result := make([]string, 0, count)
	seen := make(map[string]bool)

	// Traverse the ring clockwise
	iterations := 0
	maxIter := len(r.ring)

	for len(result) < count && iterations < maxIter {
		nodeID := r.nodes[r.ring[idx]]
		if !seen[nodeID] {
			result = append(result, nodeID)
			seen[nodeID] = true
		}
		idx = (idx + 1) % len(r.ring)
		iterations++
	}

	return result
}
