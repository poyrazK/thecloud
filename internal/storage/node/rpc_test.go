package node

import (
	"context"
	"testing"
	"time"

	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCServer(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewLocalStore(tmpDir)
	// Pass nil gossiper for now
	server := NewRPCServer(store, nil)

	ctx := context.Background()

	// 1. Store
	_, err := server.Store(ctx, &pb.StoreRequest{
		Bucket:    "bucket1",
		Key:       "key1",
		Data:      []byte("value1"),
		Timestamp: time.Now().UnixNano(),
	})
	require.NoError(t, err)

	// 2. Retrieve
	resp, err := server.Retrieve(ctx, &pb.RetrieveRequest{
		Bucket: "bucket1",
		Key:    "key1",
	})
	require.NoError(t, err)
	assert.True(t, resp.Found)
	assert.Equal(t, []byte("value1"), resp.Data)

	// 3. Delete
	_, err = server.Delete(ctx, &pb.DeleteRequest{
		Bucket: "bucket1",
		Key:    "key1",
	})
	require.NoError(t, err)

	// Verify Retrieve fails
	resp, err = server.Retrieve(ctx, &pb.RetrieveRequest{Bucket: "bucket1", Key: "key1"})
	require.NoError(t, err)
	assert.False(t, resp.Found)

	// 4. Assemble
	// Create parts manually in store first
	require.NoError(t, store.Write("bucket1", "parts/1", []byte("A"), 0))
	require.NoError(t, store.Write("bucket1", "parts/2", []byte("B"), 0))

	asmResp, err := server.Assemble(ctx, &pb.AssembleRequest{
		Bucket: "bucket1",
		Key:    "final",
		Parts:  []string{"parts/1", "parts/2"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), asmResp.Size)
}

func TestRPCServer_GetClusterStatus(t *testing.T) {
	// With nil gossiper
	server := NewRPCServer(nil, nil)
	resp, err := server.GetClusterStatus(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	assert.Empty(t, resp.Members)

	// TODO: Test with Gossiper populated
}
