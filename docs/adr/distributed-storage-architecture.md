# TheCloud Distributed Storage - Architecture Document

## Overview

A self-built distributed object storage system that spreads data across multiple servers, survives failures, and scales horizontally.

---

## System Architecture

```
                              ┌──────────────────────┐
                              │   TheCloud API       │
                              │  /storage/buckets/*  │
                              └──────────┬───────────┘
                                         │
                              ┌──────────▼───────────┐
                              │    Coordinator       │
                              │  • Hash ring         │
                              │  • Request routing   │
                              │  • Replication mgmt  │
                              └──────────┬───────────┘
                                         │
            ┌────────────────────────────┼────────────────────────────┐
            │                            │                            │
   ┌────────▼────────┐         ┌────────▼────────┐         ┌────────▼────────┐
   │  Storage Node   │         │  Storage Node   │         │  Storage Node   │
   │     (node-1)    │◄───────►│     (node-2)    │◄───────►│     (node-3)    │
   │   /data/chunks  │ gossip  │   /data/chunks  │ gossip  │   /data/chunks  │
   └─────────────────┘         └─────────────────┘         └─────────────────┘
```

---

## Core Components

### 1. Coordinator Service
Manages the cluster and routes requests.

```go
type Coordinator struct {
    ring       *ConsistentHashRing
    nodes      map[string]*StorageNode
    replicaCount int  // default: 3
}

// Route a PUT request
func (c *Coordinator) Put(bucket, key string, data io.Reader) error {
    nodes := c.ring.GetNodes(bucket + "/" + key, c.replicaCount)
    return c.replicateToNodes(nodes, bucket, key, data)
}
```

### 2. Consistent Hash Ring
Determines data placement with minimal reshuffling.

```go
type ConsistentHashRing struct {
    ring       []uint32              // sorted hash positions
    nodes      map[uint32]string     // hash → node ID
    virtualNodes int                 // typically 100-200 per physical node
}

func (r *ConsistentHashRing) GetNodes(key string, count int) []string {
    hash := fnv32(key)
    pos := r.findPosition(hash)
    return r.getNextN(pos, count)  // clockwise traversal
}
```

### 3. Storage Node
Stores data locally and participates in cluster.

```go
type StorageNode struct {
    id         string
    dataDir    string
    peers      map[string]*NodeClient
    gossiper   *GossipProtocol
}

// Local storage operations
func (n *StorageNode) Store(bucket, key string, data []byte) error
func (n *StorageNode) Retrieve(bucket, key string) ([]byte, error)
func (n *StorageNode) Delete(bucket, key string) error
```

### 4. Gossip Protocol
Nodes discover each other and detect failures.

```go
type GossipMessage struct {
    Sender     string
    Members    map[string]MemberInfo
    Timestamp  time.Time
}

type MemberInfo struct {
    Address   string
    Status    NodeStatus  // Alive, Suspect, Dead
    LastSeen  time.Time
    Heartbeat uint64
}
```

---

## Data Flow

### Write Path
```
1. Client → PUT /buckets/photos/cat.jpg
2. API → Coordinator.Put("photos", "cat.jpg", data)
3. Coordinator → hash("photos/cat.jpg") = 0x4F2A...
4. Ring lookup → [node-2, node-3, node-1] (primary + replicas)
5. Coordinator → parallel write to all 3 nodes
6. Wait for W confirmations (W=2 for quorum)
7. Return success
```

### Read Path
```
1. Client → GET /buckets/photos/cat.jpg
2. API → Coordinator.Get("photos", "cat.jpg")
3. Ring lookup → [node-2, node-3, node-1]
4. Try primary (node-2)
5. If failed → try node-3, then node-1
6. Return data
```

### Failure Recovery
```
1. Gossip detects node-2 is dead
2. Coordinator updates ring (node-2 removed)
3. Data that was on node-2 now routes to node-3
4. Background: replicate under-replicated chunks
5. When node-2 returns → anti-entropy sync
```

---

## Directory Structure

```
internal/
├── storage/
│   ├── coordinator/
│   │   ├── coordinator.go    # Main coordinator logic
│   │   ├── ring.go           # Consistent hash ring
│   │   ├── router.go         # Request routing
│   │   └── replicator.go     # Replication management
│   │
│   ├── node/
│   │   ├── node.go           # Storage node main
│   │   ├── store.go          # Local file storage
│   │   ├── rpc.go            # gRPC server
│   │   └── gossip.go         # Gossip protocol
│   │
│   └── protocol/
│       ├── storage.proto     # gRPC definitions
│       ├── messages.go       # Wire format
│       └── client.go         # Node client
│
├── core/
│   ├── domain/
│   │   └── bucket.go         # Bucket, Object models
│   └── ports/
│       └── distributed_storage.go  # Interface
│
└── handlers/
    └── storage_handler.go    # HTTP API
```

---

## Configuration

```yaml
# config/storage.yaml
storage:
  mode: distributed          # "local" or "distributed"
  
  coordinator:
    address: ":9100"
    
  cluster:
    replica_count: 3         # copies per object
    write_quorum: 2          # W for success
    read_quorum: 1           # R for consistency
    virtual_nodes: 150       # per physical node
    
  gossip:
    interval: 1s
    suspect_timeout: 5s
    dead_timeout: 30s
    
  nodes:
    - id: node-1
      address: "storage-1:9101"
      data_dir: "/data/storage"
    - id: node-2
      address: "storage-2:9101"
      data_dir: "/data/storage"
    - id: node-3
      address: "storage-3:9101"
      data_dir: "/data/storage"
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/storage/buckets/{name}` | Create bucket |
| `GET` | `/storage/buckets` | List buckets |
| `DELETE` | `/storage/buckets/{name}` | Delete bucket |
| `PUT` | `/storage/buckets/{bucket}/objects/{key}` | Upload object |
| `GET` | `/storage/buckets/{bucket}/objects/{key}` | Download object |
| `DELETE` | `/storage/buckets/{bucket}/objects/{key}` | Delete object |
| `GET` | `/storage/buckets/{bucket}/objects` | List objects |
| `GET` | `/storage/cluster/status` | Cluster health |

---

## Success Criteria

### Phase 1 Complete When:
- [ ] 3-node cluster running
- [ ] Objects distributed across nodes
- [ ] Reads work after 1 node failure
- [ ] Basic CLI: `cloud storage put/get/ls`

### Phase 2 Complete When:
- [ ] Automatic failure detection
- [ ] Self-healing replication
- [ ] Graceful node join/leave

