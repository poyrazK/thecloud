package coordinator

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

const (
	node1 = "node-1"
	node2 = "node-2"
	node3 = "node-3"
)

// MockStorageNodeClient
type MockStorageNodeClient struct {
	mock.Mock
}

func (m *MockStorageNodeClient) Store(ctx context.Context, in *pb.StoreRequest, opts ...grpc.CallOption) (*pb.StoreResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.StoreResponse)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) Retrieve(ctx context.Context, in *pb.RetrieveRequest, opts ...grpc.CallOption) (*pb.RetrieveResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.RetrieveResponse)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) Delete(ctx context.Context, in *pb.DeleteRequest, opts ...grpc.CallOption) (*pb.DeleteResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.DeleteResponse)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) Gossip(ctx context.Context, in *pb.GossipMessage, opts ...grpc.CallOption) (*pb.GossipResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.GossipResponse)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) GetClusterStatus(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ClusterStatusResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.ClusterStatusResponse)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) Assemble(ctx context.Context, in *pb.AssembleRequest, opts ...grpc.CallOption) (*pb.AssembleResponse, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(*pb.AssembleResponse)
	return r0, args.Error(1)
}

func TestCoordinatorWriteQuorum(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	client1 := new(MockStorageNodeClient)
	client2 := new(MockStorageNodeClient)
	client3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{
		node1: client1,
		node2: client2,
		node3: client3,
	}

	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	// Expect Store calls on all nodes
	// Assume N=3, W=2.
	data := []byte("hello")

	// Setup expectations
	// Note: StoreRequest includes timestamp which changes, so use mock.MatchedBy or ignore it.
	client1.On("Store", mock.Anything, mock.MatchedBy(func(req *pb.StoreRequest) bool {
		return req.Bucket == "b" && req.Key == "k" && string(req.Data) == "hello"
	})).Return(&pb.StoreResponse{Success: true}, nil)

	client2.On("Store", mock.Anything, mock.Anything).Return(&pb.StoreResponse{Success: true}, nil)
	client3.On("Store", mock.Anything, mock.Anything).Return(&pb.StoreResponse{Success: true}, nil)

	n, err := coord.Write(context.Background(), "b", "k", bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
}

func TestCoordinatorWriteQuorumFailure(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	client1 := new(MockStorageNodeClient)
	client2 := new(MockStorageNodeClient)
	client3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{
		node1: client1,
		node2: client2,
		node3: client3,
	}

	coord := NewCoordinator(ring, clients, 3) // W=2
	defer coord.Stop()

	// 2 nodes fail
	client1.On("Store", mock.Anything, mock.Anything).Return(&pb.StoreResponse{Success: false}, errors.New("failed"))
	client2.On("Store", mock.Anything, mock.Anything).Return(&pb.StoreResponse{Success: false}, errors.New("failed"))
	client3.On("Store", mock.Anything, mock.Anything).Return(&pb.StoreResponse{Success: true}, nil)

	_, err := coord.Write(context.Background(), "b", "k", strings.NewReader("hello"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write quorum failed")
}

func TestCoordinatorReadRepair(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1 := new(MockStorageNodeClient)
	c2 := new(MockStorageNodeClient)
	c3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	tsNew := time.Now().UnixNano()
	tsOld := tsNew - 1000

	// Node 1: Latest data
	c1.On("Retrieve", mock.Anything, mock.Anything).Return(&pb.RetrieveResponse{
		Found: true, Data: []byte("new"), Timestamp: tsNew,
	}, nil)

	// Node 2: Old data (needs repair)
	c2.On("Retrieve", mock.Anything, mock.Anything).Return(&pb.RetrieveResponse{
		Found: true, Data: []byte("old"), Timestamp: tsOld,
	}, nil)

	// Node 3: Not found (needs repair)
	c3.On("Retrieve", mock.Anything, mock.Anything).Return(&pb.RetrieveResponse{
		Found: false,
	}, nil)

	// Expect repair writes to Node 2 and Node 3
	c2.On("Store", mock.Anything, mock.MatchedBy(func(req *pb.StoreRequest) bool {
		return string(req.Data) == "new" && req.Timestamp == tsNew
	})).Return(&pb.StoreResponse{Success: true}, nil)

	c3.On("Store", mock.Anything, mock.MatchedBy(func(req *pb.StoreRequest) bool {
		return string(req.Data) == "new" && req.Timestamp == tsNew
	})).Return(&pb.StoreResponse{Success: true}, nil)

	r, err := coord.Read(context.Background(), "b", "k")
	assert.NoError(t, err)

	data, _ := io.ReadAll(r)
	assert.Equal(t, "new", string(data))

	// Wait for async repair
	time.Sleep(100 * time.Millisecond)
	c2.AssertCalled(t, "Store", mock.Anything, mock.Anything)
	c3.AssertCalled(t, "Store", mock.Anything, mock.Anything)
}

func TestCoordinatorDelete(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1 := new(MockStorageNodeClient)
	c2 := new(MockStorageNodeClient)
	c3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	// All nodes success
	c1.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)
	c2.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)
	c3.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)

	err := coord.Delete(context.Background(), "b", "k")
	assert.NoError(t, err)
}

func TestCoordinatorAssemble(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1 := new(MockStorageNodeClient)
	c2 := new(MockStorageNodeClient)
	c3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	// Mock assembly on nodes
	c1.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)
	c2.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)
	c3.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)

	size, err := coord.Assemble(context.Background(), "b", "k", []string{"p1", "p2"})
	assert.NoError(t, err)
	assert.Equal(t, int64(100), size)
}

func TestCoordinatorGetClusterStatus(t *testing.T) {
	ring := NewConsistentHashRing(10)
	clients := make(map[string]pb.StorageNodeClient)
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	status, err := coord.GetClusterStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
}
