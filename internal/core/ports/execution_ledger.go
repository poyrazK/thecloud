// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"time"
)

// ExecutionLedger provides idempotent job execution tracking.
// Before processing a job, a worker calls TryAcquire with a unique job key.
// If TryAcquire returns true, the caller owns the execution and must
// eventually call MarkComplete or MarkFailed.
// If TryAcquire returns false, another worker already processed (or is
// processing) the job and the caller should skip it.
type ExecutionLedger interface {
	// TryAcquire attempts to claim ownership of a job execution.
	// Returns true if the caller now owns the execution (inserted a new row
	// with status='running'). Returns false if the job was already acquired
	// by another worker (row exists with status='completed' or a recent
	// 'running' entry within staleThreshold).
	//
	// If a previous 'running' entry is older than staleThreshold, it is
	// considered abandoned and the caller can reclaim it.
	TryAcquire(ctx context.Context, jobKey string, staleThreshold time.Duration) (bool, error)

	// MarkComplete marks a job execution as successfully completed.
	MarkComplete(ctx context.Context, jobKey string, result string) error

	// MarkFailed marks a job execution as failed, allowing future retries.
	MarkFailed(ctx context.Context, jobKey string, reason string) error

	// GetStatus returns the current status, result and start time of a job.
	// Returns status="", nil error if the job does not exist.
	GetStatus(ctx context.Context, jobKey string) (status string, result string, startedAt time.Time, err error)
}
