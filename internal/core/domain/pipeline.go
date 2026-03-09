// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// PipelineStatus represents the current state of a pipeline.
type PipelineStatus string

const (
	// PipelineStatusActive indicates the pipeline can accept triggers and execute builds.
	PipelineStatusActive PipelineStatus = "ACTIVE"
	// PipelineStatusPaused indicates triggers are ignored until resumed.
	PipelineStatusPaused PipelineStatus = "PAUSED"
	// PipelineStatusDeleted indicates the pipeline has been soft-deleted.
	PipelineStatusDeleted PipelineStatus = "DELETED"
)

// BuildStatus represents the current state of a pipeline execution.
type BuildStatus string

const (
	// BuildStatusQueued indicates the build has been accepted and is waiting for a worker.
	BuildStatusQueued BuildStatus = "QUEUED"
	// BuildStatusRunning indicates the build is currently executing.
	BuildStatusRunning BuildStatus = "RUNNING"
	// BuildStatusSucceeded indicates all stages/steps completed successfully.
	BuildStatusSucceeded BuildStatus = "SUCCEEDED"
	// BuildStatusFailed indicates at least one stage/step failed.
	BuildStatusFailed BuildStatus = "FAILED"
	// BuildStatusCanceled indicates execution was stopped by user/system action.
	BuildStatusCanceled BuildStatus = "CANCELED"
)

// BuildTriggerType indicates how a build was started.
type BuildTriggerType string

const (
	// BuildTriggerManual indicates a user-triggered execution from API/CLI.
	BuildTriggerManual BuildTriggerType = "MANUAL"
	// BuildTriggerWebhook indicates an inbound webhook-triggered execution.
	BuildTriggerWebhook BuildTriggerType = "WEBHOOK"
)

// Pipeline defines CI/CD configuration and source repository metadata.
type Pipeline struct {
	ID            uuid.UUID      `json:"id"`
	UserID        uuid.UUID      `json:"user_id"`
	Name          string         `json:"name"`
	RepositoryURL string         `json:"repository_url"`
	Branch        string         `json:"branch"`
	WebhookSecret string         `json:"-"`
	Config        PipelineConfig `json:"config"`
	Status        PipelineStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Build defines one execution run of a pipeline.
type Build struct {
	ID          uuid.UUID        `json:"id"`
	PipelineID  uuid.UUID        `json:"pipeline_id"`
	UserID      uuid.UUID        `json:"user_id"`
	CommitHash  string           `json:"commit_hash"`
	TriggerType BuildTriggerType `json:"trigger_type"`
	Status      BuildStatus      `json:"status"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	FinishedAt  *time.Time       `json:"finished_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// BuildStep records execution state of a resolved step in a build.
type BuildStep struct {
	ID         uuid.UUID   `json:"id"`
	BuildID    uuid.UUID   `json:"build_id"`
	Name       string      `json:"name"`
	Image      string      `json:"image"`
	Commands   []string    `json:"commands"`
	Status     BuildStatus `json:"status"`
	ExitCode   *int        `json:"exit_code,omitempty"`
	StartedAt  *time.Time  `json:"started_at,omitempty"`
	FinishedAt *time.Time  `json:"finished_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// BuildLog contains one persisted log chunk produced by build execution.
type BuildLog struct {
	ID        uuid.UUID `json:"id"`
	BuildID   uuid.UUID `json:"build_id"`
	StepID    uuid.UUID `json:"step_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// PipelineConfig is the declarative execution plan for a pipeline.
type PipelineConfig struct {
	Stages      []PipelineStage   `json:"stages"`
	Environment map[string]string `json:"environment,omitempty"`
}

// PipelineStage groups steps executed sequentially.
type PipelineStage struct {
	Name  string         `json:"name"`
	Steps []PipelineStep `json:"steps"`
}

// PipelineStep defines an executable unit in a stage.
type PipelineStep struct {
	Name     string   `json:"name"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}
