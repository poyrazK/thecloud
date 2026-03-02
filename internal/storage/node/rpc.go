// Package node implements storage node services.
package node

import (
	"context"
	"io"
	"os"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
)

// RPCServer exposes storage-node RPC endpoints.
type RPCServer struct {
	pb.UnimplementedStorageNodeServer
	store    *LocalStore
	gossiper *GossipProtocol
}

// NewRPCServer constructs an RPCServer for storage operations.
func NewRPCServer(store *LocalStore, gossiper *GossipProtocol) *RPCServer {
	return &RPCServer{store: store, gossiper: gossiper}
}

func (s *RPCServer) Store(stream pb.StorageNode_StoreServer) error {
	// 1. Receive metadata
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	meta := req.GetMetadata()
	if meta == nil {
		return stream.SendAndClose(&pb.StoreResponse{Success: false, Error: "metadata expected as first message"})
	}

	// 2. Stream data to store
	reader := &grpcStoreReader{stream: stream}
	_, err = s.store.WriteStream(meta.Bucket, meta.Key, reader, meta.Timestamp)
	if err != nil {
		return stream.SendAndClose(&pb.StoreResponse{Success: false, Error: err.Error()})
	}

	return stream.SendAndClose(&pb.StoreResponse{Success: true})
}

type grpcStoreReader struct {
	stream pb.StorageNode_StoreServer
	buffer []byte
}

func (r *grpcStoreReader) Read(p []byte) (n int, err error) {
	if len(r.buffer) > 0 {
		n = copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		return n, nil
	}

	req, err := r.stream.Recv()
	if err == io.EOF {
		return 0, io.EOF
	}
	if err != nil {
		return 0, err
	}

	chunk := req.GetChunkData()
	if chunk == nil {
		return 0, nil // Skip non-data messages if any
	}

	n = copy(p, chunk)
	if n < len(chunk) {
		r.buffer = chunk[n:]
	}
	return n, nil
}

func (s *RPCServer) Retrieve(req *pb.RetrieveRequest, stream pb.StorageNode_RetrieveServer) error {
	rc, timestamp, err := s.store.ReadStream(req.Bucket, req.Key)
	if err != nil {
		if os.IsNotExist(err) {
			return stream.Send(&pb.RetrieveResponse{
				Payload: &pb.RetrieveResponse_Metadata{
					Metadata: &pb.RetrieveMetadata{Found: false},
				},
			})
		}
		return err
	}
	defer func() { _ = rc.Close() }()

	// Send metadata first
	err = stream.Send(&pb.RetrieveResponse{
		Payload: &pb.RetrieveResponse_Metadata{
			Metadata: &pb.RetrieveMetadata{
				Found:     true,
				Timestamp: timestamp,
			},
		},
	})
	if err != nil {
		return err
	}

	// Stream chunks
	buf := make([]byte, 1024*1024) // 1MB chunks
	for {
		n, err := rc.Read(buf)
		if n > 0 {
			errSend := stream.Send(&pb.RetrieveResponse{
				Payload: &pb.RetrieveResponse_ChunkData{
					ChunkData: buf[:n],
				},
			})
			if errSend != nil {
				return errSend
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.store.Delete(req.Bucket, req.Key)
	if err != nil && !os.IsNotExist(err) {
		return &pb.DeleteResponse{Success: false, Error: err.Error()}, err
	}
	return &pb.DeleteResponse{Success: true}, nil
}

func (s *RPCServer) Gossip(ctx context.Context, req *pb.GossipMessage) (*pb.GossipResponse, error) {
	if s.gossiper != nil {
		s.gossiper.OnGossip(req)
	}
	return &pb.GossipResponse{Success: true}, nil
}

func (s *RPCServer) GetClusterStatus(ctx context.Context, req *pb.Empty) (*pb.ClusterStatusResponse, error) {
	if s.gossiper == nil {
		return &pb.ClusterStatusResponse{}, nil
	}

	s.gossiper.mu.RLock()
	defer s.gossiper.mu.RUnlock()

	members := make(map[string]*pb.MemberState)
	for id, m := range s.gossiper.members {
		members[id] = &pb.MemberState{
			Addr:      m.Address,
			Status:    m.Status,
			LastSeen:  m.LastSeen.Unix(),
			Heartbeat: m.Heartbeat,
		}
	}

	return &pb.ClusterStatusResponse{Members: members}, nil
}

func (s *RPCServer) Assemble(ctx context.Context, req *pb.AssembleRequest) (*pb.AssembleResponse, error) {
	size, err := s.store.Assemble(req.Bucket, req.Key, req.Parts)
	if err != nil {
		return &pb.AssembleResponse{Error: err.Error()}, err
	}
	return &pb.AssembleResponse{Size: size}, nil
}
