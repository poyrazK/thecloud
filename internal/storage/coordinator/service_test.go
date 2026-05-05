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

func (m *MockStoreClient) Context() context.Context     { return context.Background() }
func (m *MockStoreClient) Header() (metadata.MD, error) { return nil, nil }
func (m *MockStoreClient) Trailer() metadata.MD         { return nil }
func (m *MockStoreClient) CloseSend() error             { return nil }

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

func (m *MockRetrieveClient) Context() context.Context     { return context.Background() }
func (m *MockRetrieveClient) Header() (metadata.MD, error) { return nil, nil }
func (m *MockRetrieveClient) Trailer() metadata.MD         { return nil }
func (m *MockRetrieveClient) CloseSend() error             { return nil }

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

func TestCoordinatorWriteQuorum_TCs(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(m1, m2, m3 *MockStorageNodeClient)
		expectedError string
		expectedSize  int64
	}{
		{
			name: "Success_AllNodes",
			setupMocks: func(m1, m2, m3 *MockStorageNodeClient) {
				setupSuccess := func(m *MockStorageNodeClient) {
					sm := new(MockStoreClient)
					sm.On("Send", mock.Anything).Return(nil)
					sm.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
					m.On("Store", mock.Anything).Return(sm, nil)
				}
				setupSuccess(m1)
				setupSuccess(m2)
				setupSuccess(m3)
			},
			expectedSize: 5,
		},
		{
			name: "Failure_QuorumNotMet",
			setupMocks: func(m1, m2, m3 *MockStorageNodeClient) {
				// m1 fails init
				m1.On("Store", mock.Anything).Return(nil, errors.New("init failed"))
				// m2 fails Send
				sm2 := new(MockStoreClient)
				sm2.On("Send", mock.Anything).Return(errors.New("send failed"))
				m2.On("Store", mock.Anything).Return(sm2, nil)
				// m3 succeeds
				sm3 := new(MockStoreClient)
				sm3.On("Send", mock.Anything).Return(nil)
				sm3.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
				m3.On("Store", mock.Anything).Return(sm3, nil)
			},
			expectedError: "write quorum failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ring := NewConsistentHashRing(10)
			ring.AddNode(node1)
			ring.AddNode(node2)
			ring.AddNode(node3)

			c1, c2, c3 := new(MockStorageNodeClient), new(MockStorageNodeClient), new(MockStorageNodeClient)
			tt.setupMocks(c1, c2, c3)

			clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
			coord := NewCoordinator(context.Background(), ring, clients, 3)
			defer coord.Stop()

			n, err := coord.Write(context.Background(), "b", "k", bytes.NewReader([]byte("hello")))
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSize, n)
			}
		})
	}
}

func TestCoordinatorReadRepair(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1, c2, c3 := new(MockStorageNodeClient), new(MockStorageNodeClient), new(MockStorageNodeClient)
	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(context.Background(), ring, clients, 3)
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

	// Repairs expected for c2 and c3 with independent mocks
	smRepairC2 := new(MockStoreClient)
	smRepairC2.On("Send", mock.Anything).Return(nil)
	smRepairC2.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
	c2.On("Store", mock.Anything).Return(smRepairC2, nil)

	smRepairC3 := new(MockStoreClient)
	smRepairC3.On("Send", mock.Anything).Return(nil)
	smRepairC3.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
	c3.On("Store", mock.Anything).Return(smRepairC3, nil)

	r, err := coord.Read(context.Background(), "b", "k")
	require.NoError(t, err)

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))
	_ = r.Close()

	// Wait for async repair
	time.Sleep(200 * time.Millisecond)
	c2.AssertCalled(t, "Store", mock.Anything)
	c3.AssertCalled(t, "Store", mock.Anything)
}

func TestCoordinatorDelete(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	c1, c2, c3 := new(MockStorageNodeClient), new(MockStorageNodeClient), new(MockStorageNodeClient)
	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(context.Background(), ring, clients, 3)
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

	c1, c2, c3 := new(MockStorageNodeClient), new(MockStorageNodeClient), new(MockStorageNodeClient)
	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(context.Background(), ring, clients, 3)
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
	coord := NewCoordinator(context.Background(), ring, clients, 3)
	defer coord.Stop()

	status, err := coord.GetClusterStatus(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, status)
}

func TestCoordinatorWriteStreamFailureMidChunk(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	// Node A (node1): succeeds on all sends
	sm1 := new(MockStoreClient)
	sm1.On("Send", mock.Anything).Return(nil).Once() // metadata
	sm1.On("Send", mock.Anything).Return(nil).Once() // chunk
	sm1.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)

	// Node B (node2): fails on first chunk Send (after metadata Send succeeds)
	sm2 := new(MockStoreClient)
	sm2.On("Send", mock.Anything).Return(nil).Once() // metadata succeeds
	sm2.On("Send", mock.Anything).Return(errors.New("mid-stream failure")).Once()
	sm2.On("CloseAndRecv").Return(&pb.StoreResponse{Success: false, Error: "stream error"}, nil)

	// Node C (node3): succeeds on all sends
	sm3 := new(MockStoreClient)
	sm3.On("Send", mock.Anything).Return(nil).Maybe()
	sm3.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)

	c1, c2, c3 := new(MockStorageNodeClient), new(MockStorageNodeClient), new(MockStorageNodeClient)
	c1.On("Store", mock.Anything).Return(sm1, nil)
	c2.On("Store", mock.Anything).Return(sm2, nil)
	c3.On("Store", mock.Anything).Return(sm3, nil)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(context.Background(), ring, clients, 3)
	defer coord.Stop()

	// One chunk of data "hello" — 3 nodes, writeQuorum=2 (3/2+1), B fails mid-stream
	n, err := coord.Write(context.Background(), "b", "k", bytes.NewReader([]byte("hello")))
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)

	// B's stream had 1 successful Send (metadata) then 1 failing Send (chunk) — verified by mock
	sm2.AssertNumberOfCalls(t, "Send", 2)

	// A and C should have received the chunk (at least 2 Sends each: metadata + chunk)
	sm1.AssertNumberOfCalls(t, "Send", 2)
	sm3.AssertNumberOfCalls(t, "Send", 2)

	// A and C must have called CloseAndRecv (stream stayed live)
	sm1.AssertCalled(t, "CloseAndRecv")
	sm3.AssertCalled(t, "CloseAndRecv")

	// B's stream should also have called CloseAndRecv (to clean up the failed stream)
	sm2.AssertCalled(t, "CloseAndRecv")
}

func TestCoordinatorRepairStreamFailureContinues(t *testing.T) {
	ring := NewConsistentHashRing(10)
	ring.AddNode(node1)
	ring.AddNode(node2)
	ring.AddNode(node3)

	// Setup read: node1 is the winner (newest), node2 and node3 are stale
	tsNew := time.Now().UnixNano()
	tsOld := tsNew - 1000

	rm1 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: true, Timestamp: tsNew}}},
		{Payload: &pb.RetrieveResponse_ChunkData{ChunkData: []byte("newdata")}},
	}}
	c1 := new(MockStorageNodeClient)
	c1.On("Retrieve", mock.Anything, mock.Anything).Return(rm1, nil)

	// node2: stale, repair will fail on first chunk
	rm2 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: true, Timestamp: tsOld}}},
		{Payload: &pb.RetrieveResponse_ChunkData{ChunkData: []byte("olddata")}},
	}}
	c2 := new(MockStorageNodeClient)
	c2.On("Retrieve", mock.Anything, mock.Anything).Return(rm2, nil)

	// node3: stale, repair should succeed
	rm3 := &MockRetrieveClient{resps: []*pb.RetrieveResponse{
		{Payload: &pb.RetrieveResponse_Metadata{Metadata: &pb.RetrieveMetadata{Found: true, Timestamp: tsOld}}},
		{Payload: &pb.RetrieveResponse_ChunkData{ChunkData: []byte("olddata")}},
	}}
	c3 := new(MockStorageNodeClient)
	c3.On("Retrieve", mock.Anything, mock.Anything).Return(rm3, nil)

	// Repair streams: node2 fails on first chunk, node3 succeeds
	smRepair2 := new(MockStoreClient)
	smRepair2.On("Send", mock.Anything).Return(nil).Once()  // metadata
	smRepair2.On("Send", mock.Anything).Return(errors.New("repair node2 failure")).Once()
	smRepair2.On("CloseAndRecv").Return(&pb.StoreResponse{Success: false, Error: "repair failed"}, nil)
	c2.On("Store", mock.Anything).Return(smRepair2, nil)

	smRepair3 := new(MockStoreClient)
	smRepair3.On("Send", mock.Anything).Return(nil).Maybe()
	smRepair3.On("CloseAndRecv").Return(&pb.StoreResponse{Success: true}, nil)
	c3.On("Store", mock.Anything).Return(smRepair3, nil)

	clients := map[string]pb.StorageNodeClient{node1: c1, node2: c2, node3: c3}
	coord := NewCoordinator(context.Background(), ring, clients, 3)
	defer coord.Stop()

	r, err := coord.Read(context.Background(), "b", "k")
	require.NoError(t, err)

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "newdata", string(data))
	_ = r.Close()

	// Wait for async repair goroutines
	time.Sleep(200 * time.Millisecond)

	// node2: metadata succeeded, chunk failed, CloseAndRecv called to clean up
	smRepair2.AssertNumberOfCalls(t, "Send", 2)
	smRepair2.AssertCalled(t, "CloseAndRecv")

	// node3: both metadata and chunk sent, CloseAndRecv called
	smRepair3.AssertNumberOfCalls(t, "Send", 2)
	smRepair3.AssertCalled(t, "CloseAndRecv")
}
