package noop

import (
	"context"
	"io"
	"strings"
	"testing"

	"log/slog"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

const (
	errGetByID = "GetByID returned error: %v"
	errDelete  = "Delete returned error: %v"
)

func TestNoopInstanceAndVolumeRepositories(t *testing.T) {
	ctx := context.Background()

	instRepo := &NoopInstanceRepository{}
	instID := uuid.New()
	_, err := instRepo.GetByID(ctx, instID)
	require.NoError(t, err, errGetByID, err)

	_, err = instRepo.GetByName(ctx, "name")
	require.NoError(t, err, "GetByName returned error: %v", err)

	err = instRepo.Create(ctx, &domain.Instance{ID: instID})
	require.NoError(t, err, "Create returned error: %v", err)

	err = instRepo.Delete(ctx, instID)
	require.NoError(t, err, errDelete, err)

	volRepo := &NoopVolumeRepository{}
	volID := uuid.New()
	_, err = volRepo.GetByID(ctx, volID)
	require.NoError(t, err, "volume GetByID returned error: %v", err)

	err = volRepo.Delete(ctx, volID)
	require.NoError(t, err, "volume Delete returned error: %v", err)
}

func TestNoopSubnetRepository(t *testing.T) {
	ctx := context.Background()
	repo := &NoopSubnetRepository{}
	id := uuid.New()
	_, err := repo.GetByID(ctx, id)
	require.NoError(t, err, errGetByID, err)

	_, err = repo.GetByName(ctx, uuid.New(), "s")
	require.NoError(t, err, "GetByName returned error: %v", err)

	err = repo.Delete(ctx, id)
	require.NoError(t, err, errDelete, err)
}

func TestNoopComputeBackend(t *testing.T) {
	ctx := context.Background()
	backend := NewNoopComputeBackend()
	require.Equal(t, "noop", backend.Type())

	t.Run("InstanceLifecycle", func(t *testing.T) {
		id, _, err := backend.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, id)

		require.NoError(t, backend.StopInstance(ctx, id))
		require.NoError(t, backend.DeleteInstance(ctx, id))
	})

	t.Run("InspectAndLogs", func(t *testing.T) {
		id := "test-id"
		_, err := backend.GetInstanceLogs(ctx, id)
		require.NoError(t, err)

		_, err = backend.GetInstanceStats(ctx, id)
		require.NoError(t, err)

		_, err = backend.GetInstancePort(ctx, id, "8080")
		require.NoError(t, err)

		ip, err := backend.GetInstanceIP(ctx, id)
		require.NoError(t, err)
		require.NotEmpty(t, ip)

		url, err := backend.GetConsoleURL(ctx, id)
		require.NoError(t, err)
		require.NotEmpty(t, url)
	})

	t.Run("ExecAndTasks", func(t *testing.T) {
		_, err := backend.Exec(ctx, "id", []string{"echo", "hi"})
		require.NoError(t, err)

		taskID, _, err := backend.RunTask(ctx, ports.RunTaskOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, taskID)

		_, err = backend.WaitTask(ctx, "tid")
		require.NoError(t, err)
	})

	t.Run("NetworkingAndVolumes", func(t *testing.T) {
		_, err := backend.CreateNetwork(ctx, "net")
		require.NoError(t, err)
		require.NoError(t, backend.DeleteNetwork(ctx, "net"))

		require.NoError(t, backend.AttachVolume(ctx, "id", "vol"))
		require.NoError(t, backend.DetachVolume(ctx, "id", "vol"))
		require.NoError(t, backend.Ping(ctx))
	})
}

func TestNoopEventAndAuditServices(t *testing.T) {
	ctx := context.Background()
	ev := &NoopEventService{}
	if err := ev.RecordEvent(ctx, "type", "rid", "rtype", nil); err != nil {
		t.Fatalf("RecordEvent returned error: %v", err)
	}
	if events, err := ev.ListEvents(ctx, 1); err != nil || len(events) != 0 {
		t.Fatalf("ListEvents returned err=%v len=%d", err, len(events))
	}

	aud := &NoopAuditService{}
	if err := aud.Log(ctx, uuid.New(), "action", "resource", "rid", nil); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}
	if logs, err := aud.ListLogs(ctx, uuid.New(), 1); err != nil || len(logs) != 0 {
		t.Fatalf("ListLogs returned err=%v len=%d", err, len(logs))
	}
}

func TestNoopStorage(t *testing.T) {
	ctx := context.Background()
	store := &NoopStorageRepository{}
	if err := store.SaveMeta(ctx, &domain.Object{Bucket: "b", Key: "k"}); err != nil {
		t.Fatalf("SaveMeta error: %v", err)
	}
	if obj, err := store.GetMeta(ctx, "b", "k"); err != nil || obj == nil {
		t.Fatalf("GetMeta err=%v obj=%v", err, obj)
	}
	if list, err := store.List(ctx, "b"); err != nil || len(list) != 0 {
		t.Fatalf("List err=%v len=%d", err, len(list))
	}
	if err := store.SoftDelete(ctx, "b", "k"); err != nil {
		t.Fatalf("SoftDelete error: %v", err)
	}

	backend := NewNoopStorageBackend()
	if backend.Type() != "noop" {
		t.Fatalf("expected storage backend type noop")
	}
	if _, err := backend.CreateVolume(ctx, "vol", 1); err != nil {
		t.Fatalf("CreateVolume error: %v", err)
	}
	if err := backend.AttachVolume(ctx, "vol", "inst"); err != nil {
		t.Fatalf("AttachVolume error: %v", err)
	}
	if err := backend.DetachVolume(ctx, "vol", "inst"); err != nil {
		t.Fatalf("DetachVolume error: %v", err)
	}
	if err := backend.CreateSnapshot(ctx, "vol", "snap"); err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}
	if err := backend.RestoreSnapshot(ctx, "vol", "snap"); err != nil {
		t.Fatalf("RestoreSnapshot error: %v", err)
	}
	if err := backend.DeleteSnapshot(ctx, "snap"); err != nil {
		t.Fatalf("DeleteSnapshot error: %v", err)
	}
	if err := backend.DeleteVolume(ctx, "vol"); err != nil {
		t.Fatalf("DeleteVolume error: %v", err)
	}
	if err := backend.Ping(ctx); err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestNoopFileStore(t *testing.T) {
	ctx := context.Background()
	store := &NoopFileStore{}
	if _, err := store.Write(ctx, "b", "k", strings.NewReader("")); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if rc, err := store.Read(ctx, "b", "k"); err != nil {
		t.Fatalf("Read returned error: %v", err)
	} else {
		_ = rc.Close()
	}
	if err := store.Delete(ctx, "b", "k"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

func TestNoopFunctionAndIdentity(t *testing.T) {
	ctx := context.Background()

	t.Run("FunctionRepository", func(t *testing.T) {
		fRepo := &NoopFunctionRepository{}
		fID := uuid.New()
		_, err := fRepo.GetByID(ctx, fID)
		require.NoError(t, err, errGetByID, err)

		require.NoError(t, fRepo.CreateInvocation(ctx, &domain.Invocation{}))
		list, err := fRepo.GetInvocations(ctx, fID, 1)
		require.NoError(t, err)
		require.Empty(t, list)
	})

	t.Run("IdentityService", func(t *testing.T) {
		identSvc := &NoopIdentityService{}
		userID := uuid.New()
		key, err := identSvc.CreateKey(ctx, userID, "name")
		require.NoError(t, err)
		require.NotNil(t, key)

		_, err = identSvc.ValidateAPIKey(ctx, "k")
		require.NoError(t, err)

		list, err := identSvc.ListKeys(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, list)

		require.NoError(t, identSvc.RevokeKey(ctx, userID, uuid.New()))
		rotKey, err := identSvc.RotateKey(ctx, userID, uuid.New())
		require.NoError(t, err)
		require.NotNil(t, rotKey)
	})

	t.Run("IdentityRepository", func(t *testing.T) {
		identRepo := &NoopIdentityRepository{}
		userID := uuid.New()
		require.NoError(t, identRepo.CreateAPIKey(ctx, &domain.APIKey{}))
		_, err := identRepo.GetAPIKeyByKey(ctx, "k")
		require.NoError(t, err)
		_, err = identRepo.GetAPIKeyByID(ctx, uuid.New())
		require.NoError(t, err)
		list, err := identRepo.ListAPIKeysByUserID(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, list)
		require.NoError(t, identRepo.DeleteAPIKey(ctx, uuid.New()))
	})
}

func TestNoopNetworkAdapter(t *testing.T) {
	ctx := context.Background()
	adapter := NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.Equal(t, "noop", adapter.Type())

	t.Run("Bridges", func(t *testing.T) {
		require.NoError(t, adapter.CreateBridge(ctx, "br0", 1))
		require.NoError(t, adapter.DeleteBridge(ctx, "br0"))
		list, err := adapter.ListBridges(ctx)
		require.NoError(t, err)
		require.Empty(t, list)
	})

	t.Run("PortsAndTunnels", func(t *testing.T) {
		require.NoError(t, adapter.AddPort(ctx, "br0", "p0"))
		require.NoError(t, adapter.DeletePort(ctx, "br0", "p0"))
		require.NoError(t, adapter.CreateVXLANTunnel(ctx, "br0", 1, testutil.TestNoopIP1))
		require.NoError(t, adapter.DeleteVXLANTunnel(ctx, "br0", testutil.TestNoopIP1))
	})

	t.Run("FlowRules", func(t *testing.T) {
		require.NoError(t, adapter.AddFlowRule(ctx, "br0", ports.FlowRule{}))
		require.NoError(t, adapter.DeleteFlowRule(ctx, "br0", "match"))
		flows, err := adapter.ListFlowRules(ctx, "br0")
		require.NoError(t, err)
		require.Empty(t, flows)
	})

	t.Run("VethAndL3", func(t *testing.T) {
		require.NoError(t, adapter.CreateVethPair(ctx, "h", "c"))
		require.NoError(t, adapter.AttachVethToBridge(ctx, "br0", "h"))
		require.NoError(t, adapter.DeleteVethPair(ctx, "h"))
		require.NoError(t, adapter.SetVethIP(ctx, "h", testutil.TestNoopIP2, "24"))
		require.NoError(t, adapter.Ping(ctx))
	})
}
