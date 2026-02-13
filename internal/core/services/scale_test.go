package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
)

// BenchmarkRealWorldLifecycle simulates a full user session resource cycle.
func BenchmarkRealWorldLifecycle(b *testing.B) {
	logger := slog.Default()

	// Components
	instRepo := &noop.NoopInstanceRepository{}
	vpcRepo := &noop.NoopVpcRepository{}
	subnetRepo := &noop.NoopSubnetRepository{}
	volumeRepo := &noop.NoopVolumeRepository{}
	compute := &noop.NoopComputeBackend{}
	network := noop.NewNoopNetworkAdapter(logger)
	eventSvc := &noop.NoopEventService{}
	auditSvc := &noop.NoopAuditService{}

	instSvc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             instRepo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        &services.TaskQueueStub{},
		Logger:           logger,
		TenantSvc:        &NoopTenantService{},
		InstanceTypeRepo: &noop.NoopInstanceTypeRepository{},
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := uuid.New()
		pCtx := appcontext.WithUserID(ctx, userID)

		vID := uuid.New()
		sID := uuid.New()

		inst, err := instSvc.LaunchInstance(pCtx, ports.LaunchParams{
			Name:         "test-server",
			Image:        "ubuntu",
			Ports:        "22:22",
			InstanceType: "basic-2",
			VpcID:        &vID,
			SubnetID:     &sID,
		})
		if err != nil {
			b.Fatal(err)
		}

		_, _ = instSvc.GetInstance(pCtx, inst.ID.String())
		_ = instSvc.TerminateInstance(pCtx, inst.ID.String())
	}
}

// BenchmarkRealWorldLifecycleParallel tests concurrent resource cycles.
func BenchmarkRealWorldLifecycleParallel(b *testing.B) {
	logger := slog.Default()
	instRepo := &noop.NoopInstanceRepository{}
	vpcRepo := &noop.NoopVpcRepository{}
	subnetRepo := &noop.NoopSubnetRepository{}
	volumeRepo := &noop.NoopVolumeRepository{}
	compute := &noop.NoopComputeBackend{}
	network := noop.NewNoopNetworkAdapter(logger)
	eventSvc := &noop.NoopEventService{}
	auditSvc := &noop.NoopAuditService{}

	instSvc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             instRepo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        &services.TaskQueueStub{},
		Logger:           logger,
		TenantSvc:        &NoopTenantService{},
		InstanceTypeRepo: &noop.NoopInstanceTypeRepository{},
	})

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			userID := uuid.New()
			pCtx := appcontext.WithUserID(ctx, userID)
			vID := uuid.New()
			sID := uuid.New()

			inst, _ := instSvc.LaunchInstance(pCtx, ports.LaunchParams{
				Name:         "test-server",
				Image:        "ubuntu",
				Ports:        "22:22",
				InstanceType: "basic-2",
				VpcID:        &vID,
				SubnetID:     &sID,
			})
			if inst != nil {
				_, _ = instSvc.GetInstance(pCtx, inst.ID.String())
				_ = instSvc.TerminateInstance(pCtx, inst.ID.String())
			}
		}
	})
}
