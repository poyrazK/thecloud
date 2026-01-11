package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type accountingService struct {
	repo         ports.AccountingRepository
	instanceRepo ports.InstanceRepository
	// In a real system, we'd have price configuration here
}

func NewAccountingService(repo ports.AccountingRepository, instanceRepo ports.InstanceRepository) ports.AccountingService {
	return &accountingService{
		repo:         repo,
		instanceRepo: instanceRepo,
	}
}

func (s *accountingService) TrackUsage(ctx context.Context, record domain.UsageRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	return s.repo.CreateRecord(ctx, record)
}

func (s *accountingService) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.BillSummary, error) {
	usage, err := s.repo.GetUsageSummary(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	// Simple simulated billing: $0.01 per minute for instances, etc.
	total := 0.0
	for resType, quantity := range usage {
		switch resType {
		case domain.ResourceInstance:
			total += quantity * 0.01
		case domain.ResourceStorage:
			total += quantity * 0.005
		}
	}

	return &domain.BillSummary{
		UserID:      userID,
		TotalAmount: total,
		Currency:    "USD",
		UsageByType: usage,
		PeriodStart: start,
		PeriodEnd:   end,
	}, nil
}

func (s *accountingService) ListUsage(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) {
	return s.repo.ListRecords(ctx, userID, start, end)
}

func (s *accountingService) ProcessHourlyBilling(ctx context.Context) error {
	// 1. Get all running instances
	// For simplicity in this demo, let's assume we collect usage for the last hour
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	instances, err := s.instanceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances for billing: %w", err)
	}

	for _, inst := range instances {
		if inst.Status == domain.StatusRunning {
			// Record 60 minutes of usage
			record := domain.UsageRecord{
				ID:           uuid.New(),
				UserID:       inst.UserID,
				ResourceID:   inst.ID,
				ResourceType: domain.ResourceInstance,
				Quantity:     60, // 60 minutes
				Unit:         "minute",
				StartTime:    startTime,
				EndTime:      now,
			}
			if err := s.repo.CreateRecord(ctx, record); err != nil {
				// Log error but continue
				fmt.Printf("Error recording usage for instance %s: %v\n", inst.ID, err)
			}
		}
	}

	return nil
}
