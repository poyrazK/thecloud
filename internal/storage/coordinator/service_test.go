package coordinator

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	node1 = "node-1"
	node2 = "node-2"
	node3 = "node-3"
)

// MockStoreClient implements pb.StorageNode_StoreClient
type MockStoreClient struct {
	mock.Mock
	grpc.ClientStream
}

func (m *MockStoreClient) Send(req *pb.StoreRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockStoreClient) CloseAndRecv() (*pb.StoreResponse, error) {
	args := m.Called()
	r0, _ := args.Get(0).(*pb.StoreResponse)
	return r0, args.Error(1)
}

func (m *MockStoreClient) Context() context.Context { return context.Background() }
func (m *MockStoreClient) Header() (metadata.MD, error) { return nil, nil }
func (m *MockStoreClient) Trailer() metadata.MD { return nil }
func (m *MockStoreClient) CloseSend() error { return nil }

// MockRetrieveClient implements pb.StorageNode_RetrieveClient
type MockRetrieveClient struct {
	mock.Mock
	grpc.ClientStream
	resps []*pb.RetrieveResponse
	idx   int
}

func (m *MockRetrieveClient) Recv() (*pb.RetrieveResponse, error) {
	if m.idx >= len(m.resps) {
		return nil, io.EOF
	}
	r := m.resps[m.idx]
	m.idx++
	return r, nil
}

func (m *MockRetrieveClient) Context() context.Context { return context.Background() }
func (m *MockRetrieveClient) Header() (metadata.MD, error) { return nil, nil }
func (m *MockRetrieveClient) Trailer() metadata.MD { return nil }
func (m *MockRetrieveClient) CloseSend() error { return nil }

// MockStorageNodeClient
type MockStorageNodeClient struct {
	mock.Mock
}

func (m *MockStorageNodeClient) Store(ctx context.Context, opts ...grpc.CallOption) (pb.StorageNode_StoreClient, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).(pb.StorageNode_StoreClient)
	return r0, args.Error(1)
}

func (m *MockStorageNodeClient) Retrieve(ctx context.Context, in *pb.RetrieveRequest, opts ...grpc.CallOption) (pb.StorageNode_RetrieveClient, error) {
	args := m.Called(ctx, in)
	r0, _ := args.Get(0).(pb.StorageNode_RetrieveClient)
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

	c1 := new(MockStorageNodeClient)
	c2 := new(MockStorageNodeClient)
	c3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	setupMockStore := func(m *MockStorageNodeClient) {
		sm := new(MockStoreClient)
		sm.On("Send", mock.Anything).Return(nil)
		sm.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
		m.On("Store", mock.Anything).Return(sm, nil)
	}

	setupMockStore(c1)
	setupMockStore(c2)
	setupMockStore(c3)

	data := []byte("hello")
	n, err := coord.Write(context.Background(), "b", "k", bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
}

func TestCoordinatorWriteQuorumFailure(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1 := new(MockStorageNodeClient)
	c2 := new(MockStorageNodeClient)
	c3 := new(MockStorageNodeClient)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(ring, clients, 3) // W=2
	defer coord.Stop()

	// Node 1 fails init
	c1.On("Store", mock.Anything).Return(nil, errors.New("init failed"))

	// Node 2 fails Send
	sm2 := new(MockStoreClient)
	sm2.On("Send", mock.Anything).Return(errors.New("send failed"))
	c2.On("Store", mock.Anything).Return(sm2, nil)

	// Node 3 succeeds
	sm3 := new(MockStoreClient)
	sm3.On("Send", mock.Anything).Return(nil)
	sm3.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
	c3.On("Store", mock.Anything).Return(sm3, nil)

	_, err := coord.Write(context.Background(), "b", "k", bytes.NewReader([]byte("hello")))
	require.Error(t, err)
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
	rm1 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: true, Timestamp: tsNew}}},
		{Payload: &pb.RetrieveResponse_ChunkData{ChunkData: []byte("new")}},
	}}
	c1.On("Retrieve", mock.Anything, mock.Anything).Return(rm1, nil)

	// Node 2: Old data
	rm2 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: true, Timestamp: tsOld}}},
		{Payload: &pb.RetrieveResponse_ChunkData{ChunkData: []byte("old")}},
	}}
	c2.On("Retrieve", mock.Anything, mock.Anything).Return(rm2, nil)

	// Node 3: Not found
	rm3 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: false}}},
	}}
	c3.On("Retrieve", mock.Anything, mock.Anything).Return(rm3, nil)

	// Repairs expected for c2 and c3
	smRepair := new(MockStoreClient)
	smRepair.On("Send", mock.Anything).Return(nil)
	smRepair.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
	
	c2.On("Store", mock.Anything).Return(smRepair, nil)
	c3.On("Store", mock.Anything).Return(smRepair, nil)

	r, err := coord.Read(context.Background(), "b", "k")
	require.NoError(t, err)

	data, _ := io.ReadAll(r)
	assert.Equal(t, "new", string(data))

	time.Sleep(100 * time.Millisecond)
	c2.AssertCalled(t, "Store", mock.Anything)
	c3.AssertCalled(t, "Store", mock.Anything)
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

	c1.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)
	c2.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)
	c3.On("Delete", mock.Anything, mock.Anything).Return(&pb.DeleteResponse{Success: true}, nil)

	err := coord.Delete(context.Background(), "b", "k")
	require.NoError(t, err)
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

	c1.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)
	c2.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)
	c3.On("Assemble", mock.Anything, mock.Anything).Return(&pb.AssembleResponse{Size: 100}, nil)

	size, err := coord.Assemble(context.Background(), "b", "k", []string{"p1", "p2"})
	require.NoError(t, err)
	assert.Equal(t, int64(100), size)
}

func TestCoordinatorGetClusterStatus(t *testing.T) {
	ring := NewConsistentHashRing(10)
	clients := make(map[string]pb.StorageNodeClient)
	coord := NewCoordinator(ring, clients, 3)
	defer coord.Stop()

	status, err := coord.GetClusterStatus(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, status)
}
