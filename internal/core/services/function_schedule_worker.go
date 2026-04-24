// Package services implements core business workflows.
package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/robfig/cron/v3"
)

const FunctionScheduleClaimTimeout = 5 * time.Minute

// FunctionScheduleWorker polls for due function schedules and invokes them.
type FunctionScheduleWorker struct {
	repo   ports.FunctionScheduleRepository
	fnSvc  ports.FunctionService
	parser cron.Parser
}

// NewFunctionScheduleWorker constructs a FunctionScheduleWorker.
func NewFunctionScheduleWorker(repo ports.FunctionScheduleRepository, fnSvc ports.FunctionService) *FunctionScheduleWorker {
	return &FunctionScheduleWorker{
		repo:   repo,
		fnSvc:  fnSvc,
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// Run starts the worker loop.
func (w *FunctionScheduleWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	reaperTicker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	defer reaperTicker.Stop()

	log.Println("FunctionSchedule Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("FunctionSchedule Worker stopping")
			return
		case <-ticker.C:
			w.ProcessSchedules(ctx)
		case <-reaperTicker.C:
			w.reapStaleClaims(ctx)
		}
	}
}

func (w *FunctionScheduleWorker) ProcessSchedules(ctx context.Context) {
	schedules, err := w.repo.ClaimNextSchedulesToRun(ctx, FunctionScheduleClaimTimeout)
	if err != nil {
		log.Printf("FunctionScheduleWorker: failed to claim schedules: %v", err)
		return
	}

	for _, sched := range schedules {
		w.runSchedule(ctx, sched)
	}
}

func (w *FunctionScheduleWorker) runSchedule(ctx context.Context, sched *domain.FunctionSchedule) {
	start := time.Now()

	// Invoke function asynchronously - does not block the worker
	invocation, err := w.fnSvc.InvokeFunction(ctx, sched.FunctionID, sched.Payload, true)

	duration := time.Since(start)
	status := "PENDING"
	statusCode := 0
	errorMsg := ""

	if err != nil {
		status = "FAILED"
		errorMsg = err.Error()
	} else if invocation != nil {
		status = invocation.Status
		statusCode = invocation.StatusCode
		if invocation.Status == "FAILED" && len(invocation.Logs) > 0 {
			errorMsg = invocation.Logs
		}
	}

	run := &domain.FunctionScheduleRun{
		ID:           uuid.New(),
		ScheduleID:   sched.ID,
		InvocationID: func() *uuid.UUID { if invocation != nil { return &invocation.ID }; return nil }(),
		Status:       status,
		StatusCode:   statusCode,
		DurationMs:   duration.Milliseconds(),
		ErrorMessage: errorMsg,
		StartedAt:    start,
	}

	// Calculate next run
	now := time.Now()
	nextRun := now.Add(24 * time.Hour)
	if parsed, err := w.parser.Parse(sched.Schedule); err == nil {
		nextRun = parsed.Next(now)
	}

	if err := w.repo.CompleteScheduleRun(ctx, run, sched, nextRun); err != nil {
		log.Printf("FunctionScheduleWorker: failed to complete schedule run, retrying: %v", err)
		for i := 1; i <= 3; i++ {
			time.Sleep(time.Duration(i) * time.Second)
			if err := w.repo.CompleteScheduleRun(ctx, run, sched, nextRun); err == nil {
				return
			}
		}
		log.Printf("FunctionScheduleWorker: failed to complete schedule run after 3 retries, schedule %s may need manual intervention", sched.ID)
	}
}

func (w *FunctionScheduleWorker) reapStaleClaims(ctx context.Context) {
	count, err := w.repo.ReapStaleClaims(ctx)
	if err != nil {
		log.Printf("FunctionScheduleWorker: failed to reap stale claims: %v", err)
	} else if count > 0 {
		log.Printf("FunctionScheduleWorker: reclaimed %d stale claims", count)
	}
}