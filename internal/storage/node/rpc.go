// Package node implements storage node services.
package node

import (
	"context"
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

func (s *RPCServer) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	err := s.store.Write(req.Bucket, req.Key, req.Data, req.Timestamp)
	if err != nil {
		return &pb.StoreResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.StoreResponse{Success: true}, nil
}

func (s *RPCServer) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	data, timestamp, err := s.store.Read(req.Bucket, req.Key)
	if err != nil {
		if os.IsNotExist(err) {
			return &pb.RetrieveResponse{Found: false}, nil
		}
		return &pb.RetrieveResponse{Found: false, Error: err.Error()}, nil
	}
	return &pb.RetrieveResponse{Data: data, Found: true, Timestamp: timestamp}, nil
}

func (s *RPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.store.Delete(req.Bucket, req.Key)
	if err != nil && !os.IsNotExist(err) {
		return &pb.DeleteResponse{Success: false, Error: err.Error()}, nil
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
		return &pb.AssembleResponse{Error: err.Error()}, nil
	}
	return &pb.AssembleResponse{Size: size}, nil
}
