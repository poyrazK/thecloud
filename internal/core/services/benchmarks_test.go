package services_test

import (
	"context"
	"testing"

	"log/slog"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
)

func BenchmarkInstanceService_List(b *testing.B) {
	// Setup
	repo := &noop.NoopInstanceRepository{}
	compute := &noop.NoopComputeBackend{}
	eventSvc := &noop.NoopEventService{}
	auditSvc := &noop.NoopAuditService{}
	logger := slog.Default()
	svc := services.NewInstanceService(repo, compute, eventSvc, auditSvc, logger)

	ctx := context.Background()
	userID := uuid.New()

	// Seed some instances in repo if needed, but noop doesn't care
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ListInstances(ctx, userID)
	}
}

func BenchmarkVPCService_Get(b *testing.B) {
	repo := &noop.NoopVpcRepository{}
	eventSvc := &noop.NoopEventService{}
	auditSvc := &noop.NoopAuditService{}
	logger := slog.Default()
	svc := services.NewVpcService(repo, eventSvc, auditSvc, logger)

	ctx := context.Background()
	id := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetVPC(ctx, id)
	}
}
