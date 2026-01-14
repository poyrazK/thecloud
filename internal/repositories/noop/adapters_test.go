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
)

func TestNoopInstanceAndVolumeRepositories(t *testing.T) {
	ctx := context.Background()

	instRepo := &NoopInstanceRepository{}
	instID := uuid.New()
	if _, err := instRepo.GetByID(ctx, instID); err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if _, err := instRepo.GetByName(ctx, "name"); err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}
	if err := instRepo.Create(ctx, &domain.Instance{ID: instID}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := instRepo.Delete(ctx, instID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	volRepo := &NoopVolumeRepository{}
	volID := uuid.New()
	if _, err := volRepo.GetByID(ctx, volID); err != nil {
		t.Fatalf("volume GetByID returned error: %v", err)
	}
	if err := volRepo.Delete(ctx, volID); err != nil {
		t.Fatalf("volume Delete returned error: %v", err)
	}
}

func TestNoopSubnetRepository(t *testing.T) {
	ctx := context.Background()
	repo := &NoopSubnetRepository{}
	id := uuid.New()
	if _, err := repo.GetByID(ctx, id); err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if _, err := repo.GetByName(ctx, uuid.New(), "s"); err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

func TestNoopComputeBackend(t *testing.T) {
	ctx := context.Background()
	backend := NewNoopComputeBackend()

	if backend.Type() != "noop" {
		t.Fatalf("expected type noop")
	}
	id, err := backend.CreateInstance(ctx, ports.CreateInstanceOptions{})
	if err != nil || id == "" {
		t.Fatalf("CreateInstance returned err=%v id=%q", err, id)
	}
	if err := backend.StopInstance(ctx, id); err != nil {
		t.Fatalf("StopInstance returned error: %v", err)
	}
	if err := backend.DeleteInstance(ctx, id); err != nil {
		t.Fatalf("DeleteInstance returned error: %v", err)
	}
	if _, err := backend.GetInstanceLogs(ctx, id); err != nil {
		t.Fatalf("GetInstanceLogs returned error: %v", err)
	}
	if _, err := backend.GetInstanceStats(ctx, id); err != nil {
		t.Fatalf("GetInstanceStats returned error: %v", err)
	}
	if _, err := backend.GetInstancePort(ctx, id, "8080"); err != nil {
		t.Fatalf("GetInstancePort returned error: %v", err)
	}
	if ip, err := backend.GetInstanceIP(ctx, id); err != nil || ip == "" {
		t.Fatalf("GetInstanceIP returned err=%v ip=%q", err, ip)
	}
	if url, err := backend.GetConsoleURL(ctx, id); err != nil || url == "" {
		t.Fatalf("GetConsoleURL returned err=%v url=%q", err, url)
	}
	if _, err := backend.Exec(ctx, id, []string{"echo", "hi"}); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	if taskID, err := backend.RunTask(ctx, ports.RunTaskOptions{}); err != nil || taskID == "" {
		t.Fatalf("RunTask returned err=%v id=%q", err, taskID)
	}
	if _, err := backend.WaitTask(ctx, "tid"); err != nil {
		t.Fatalf("WaitTask returned error: %v", err)
	}
	if _, err := backend.CreateNetwork(ctx, "net"); err != nil {
		t.Fatalf("CreateNetwork returned error: %v", err)
	}
	if err := backend.DeleteNetwork(ctx, "net"); err != nil {
		t.Fatalf("DeleteNetwork returned error: %v", err)
	}
	if err := backend.AttachVolume(ctx, id, "vol"); err != nil {
		t.Fatalf("AttachVolume returned error: %v", err)
	}
	if err := backend.DetachVolume(ctx, id, "vol"); err != nil {
		t.Fatalf("DetachVolume returned error: %v", err)
	}
	if err := backend.Ping(ctx); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
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
	fRepo := &NoopFunctionRepository{}
	fID := uuid.New()
	if _, err := fRepo.GetByID(ctx, fID); err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if err := fRepo.CreateInvocation(ctx, &domain.Invocation{}); err != nil {
		t.Fatalf("CreateInvocation returned error: %v", err)
	}
	if list, err := fRepo.GetInvocations(ctx, fID, 1); err != nil || len(list) != 0 {
		t.Fatalf("GetInvocations err=%v len=%d", err, len(list))
	}

	identSvc := &NoopIdentityService{}
	userID := uuid.New()
	if key, err := identSvc.CreateKey(ctx, userID, "name"); err != nil || key == nil {
		t.Fatalf("CreateKey err=%v key=%v", err, key)
	}
	if _, err := identSvc.ValidateAPIKey(ctx, "k"); err != nil {
		t.Fatalf("ValidateAPIKey returned error: %v", err)
	}
	if list, err := identSvc.ListKeys(ctx, userID); err != nil || len(list) != 0 {
		t.Fatalf("ListKeys err=%v len=%d", err, len(list))
	}
	if err := identSvc.RevokeKey(ctx, userID, uuid.New()); err != nil {
		t.Fatalf("RevokeKey error: %v", err)
	}
	if key, err := identSvc.RotateKey(ctx, userID, uuid.New()); err != nil || key == nil {
		t.Fatalf("RotateKey err=%v key=%v", err, key)
	}

	identRepo := &NoopIdentityRepository{}
	if err := identRepo.CreateAPIKey(ctx, &domain.APIKey{}); err != nil {
		t.Fatalf("CreateAPIKey error: %v", err)
	}
	if _, err := identRepo.GetAPIKeyByKey(ctx, "k"); err != nil {
		t.Fatalf("GetAPIKeyByKey error: %v", err)
	}
	if _, err := identRepo.GetAPIKeyByID(ctx, uuid.New()); err != nil {
		t.Fatalf("GetAPIKeyByID error: %v", err)
	}
	if list, err := identRepo.ListAPIKeysByUserID(ctx, userID); err != nil || len(list) != 0 {
		t.Fatalf("ListAPIKeysByUserID err=%v len=%d", err, len(list))
	}
	if err := identRepo.DeleteAPIKey(ctx, uuid.New()); err != nil {
		t.Fatalf("DeleteAPIKey error: %v", err)
	}
}

func TestNoopNetworkAdapter(t *testing.T) {
	ctx := context.Background()
	adapter := NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(io.Discard, nil)))

	if adapter.Type() != "noop" {
		t.Fatalf("expected noop type")
	}
	if err := adapter.CreateBridge(ctx, "br0", 1); err != nil {
		t.Fatalf("CreateBridge error: %v", err)
	}
	if err := adapter.DeleteBridge(ctx, "br0"); err != nil {
		t.Fatalf("DeleteBridge error: %v", err)
	}
	if list, err := adapter.ListBridges(ctx); err != nil || len(list) != 0 {
		t.Fatalf("ListBridges err=%v len=%d", err, len(list))
	}
	if err := adapter.AddPort(ctx, "br0", "p0"); err != nil {
		t.Fatalf("AddPort error: %v", err)
	}
	if err := adapter.DeletePort(ctx, "br0", "p0"); err != nil {
		t.Fatalf("DeletePort error: %v", err)
	}
	if err := adapter.CreateVXLANTunnel(ctx, "br0", 1, "1.1.1.1"); err != nil {
		t.Fatalf("CreateVXLANTunnel error: %v", err)
	}
	if err := adapter.DeleteVXLANTunnel(ctx, "br0", "1.1.1.1"); err != nil {
		t.Fatalf("DeleteVXLANTunnel error: %v", err)
	}
	if err := adapter.AddFlowRule(ctx, "br0", ports.FlowRule{}); err != nil {
		t.Fatalf("AddFlowRule error: %v", err)
	}
	if err := adapter.DeleteFlowRule(ctx, "br0", "match"); err != nil {
		t.Fatalf("DeleteFlowRule error: %v", err)
	}
	if flows, err := adapter.ListFlowRules(ctx, "br0"); err != nil || len(flows) != 0 {
		t.Fatalf("ListFlowRules err=%v len=%d", err, len(flows))
	}
	if err := adapter.CreateVethPair(ctx, "h", "c"); err != nil {
		t.Fatalf("CreateVethPair error: %v", err)
	}
	if err := adapter.AttachVethToBridge(ctx, "br0", "h"); err != nil {
		t.Fatalf("AttachVethToBridge error: %v", err)
	}
	if err := adapter.DeleteVethPair(ctx, "h"); err != nil {
		t.Fatalf("DeleteVethPair error: %v", err)
	}
	if err := adapter.SetVethIP(ctx, "h", "1.1.1.2", "24"); err != nil {
		t.Fatalf("SetVethIP error: %v", err)
	}
	if err := adapter.Ping(ctx); err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}
