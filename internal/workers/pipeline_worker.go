// Package workers provides background worker implementations.
package workers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const pipelineQueueName = "pipeline_build_queue"

// PipelineWorker processes queued pipeline builds.
type PipelineWorker struct {
	repo      ports.PipelineRepository
	taskQueue ports.TaskQueue
	compute   ports.ComputeBackend
	logger    *slog.Logger
}

// NewPipelineWorker creates a new PipelineWorker.
func NewPipelineWorker(repo ports.PipelineRepository, taskQueue ports.TaskQueue, compute ports.ComputeBackend, logger *slog.Logger) *PipelineWorker {
	return &PipelineWorker{
		repo:      repo,
		taskQueue: taskQueue,
		compute:   compute,
		logger:    logger,
	}
}

func (w *PipelineWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting pipeline worker")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping pipeline worker")
			return
		default:
			msg, err := w.taskQueue.Dequeue(ctx, pipelineQueueName)
			if err != nil {
				w.logger.Error("failed to dequeue pipeline job", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if msg == "" {
				continue
			}

			var job domain.BuildJob
			if err := json.Unmarshal([]byte(msg), &job); err != nil {
				w.logger.Error("failed to unmarshal build job", "error", err)
				continue
			}

			w.processJob(job)
		}
	}
}

func (w *PipelineWorker) processJob(job domain.BuildJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	ctx = appcontext.WithUserID(ctx, job.UserID)

	build, pipeline := w.loadBuildAndPipeline(ctx, job)
	if build == nil || pipeline == nil {
		return
	}

	if !w.markBuildRunning(ctx, build) {
		return
	}

	if len(pipeline.Config.Stages) == 0 {
		w.failBuild(ctx, build, "pipeline has no stages")
		return
	}

	if !w.executePipeline(ctx, build, pipeline) {
		return
	}

	w.markBuildSucceeded(ctx, build)
}

func (w *PipelineWorker) loadBuildAndPipeline(ctx context.Context, job domain.BuildJob) (*domain.Build, *domain.Pipeline) {
	build, err := w.repo.GetBuild(ctx, job.BuildID, job.UserID)
	if err != nil || build == nil {
		w.logger.Error("failed to load build", "build_id", job.BuildID, "error", err)
		return nil, nil
	}

	pipeline, err := w.repo.GetPipeline(ctx, job.PipelineID, job.UserID)
	if err != nil || pipeline == nil {
		w.logger.Error("failed to load pipeline", "pipeline_id", job.PipelineID, "error", err)
		w.failBuild(ctx, build, "pipeline not found")
		return nil, nil
	}

	return build, pipeline
}

func (w *PipelineWorker) markBuildRunning(ctx context.Context, build *domain.Build) bool {
	now := time.Now()
	build.Status = domain.BuildStatusRunning
	build.StartedAt = &now
	build.UpdatedAt = now
	if err := w.repo.UpdateBuild(ctx, build); err != nil {
		w.logger.Error("failed to mark build running", "build_id", build.ID, "error", err)
		return false
	}
	return true
}

func (w *PipelineWorker) executePipeline(ctx context.Context, build *domain.Build, pipeline *domain.Pipeline) bool {
	for _, stage := range pipeline.Config.Stages {
		for _, stageStep := range stage.Steps {
			if !w.executeStageStep(ctx, build, stageStep) {
				return false
			}
		}
	}
	return true
}

func (w *PipelineWorker) executeStageStep(ctx context.Context, build *domain.Build, stageStep domain.PipelineStep) bool {
	step := w.newRunningStep(build.ID, stageStep)
	if err := w.repo.CreateBuildStep(ctx, step); err != nil {
		w.failBuild(ctx, build, "failed creating build step: "+err.Error())
		return false
	}

	if stageStep.Image == "" {
		w.failStepAndBuild(ctx, step, build, "step image is required")
		return false
	}

	if len(stageStep.Commands) == 0 {
		w.markStepSucceeded(ctx, step)
		return true
	}

	exitCode, logs, execErr := w.runTaskForStep(ctx, build.ID, stageStep)
	if execErr != nil {
		w.failStepAndBuild(ctx, step, build, execErr.Error())
		return false
	}

	if logs != "" {
		_ = w.repo.AppendBuildLog(ctx, &domain.BuildLog{ID: uuid.New(), BuildID: build.ID, StepID: step.ID, Content: logs, CreatedAt: time.Now()})
	}

	if exitCode != 0 {
		w.markStepFinished(ctx, step, domain.BuildStatusFailed, int(exitCode))
		w.failBuild(ctx, build, "step failed with non-zero exit code")
		return false
	}

	w.markStepFinished(ctx, step, domain.BuildStatusSucceeded, int(exitCode))
	return true
}

func (w *PipelineWorker) runTaskForStep(ctx context.Context, buildID uuid.UUID, stageStep domain.PipelineStep) (int64, string, error) {
	containerID, _, runErr := w.compute.RunTask(ctx, ports.RunTaskOptions{
		Name:    "pipeline-" + buildID.String(),
		Image:   stageStep.Image,
		Command: []string{"/bin/sh", "-lc", strings.Join(stageStep.Commands, "\n")},
	})
	if runErr != nil {
		return 0, "", runErr
	}
	defer func() { _ = w.compute.DeleteInstance(context.Background(), containerID) }()

	exitCode, waitErr := w.compute.WaitTask(ctx, containerID)
	if waitErr != nil {
		return 0, "", waitErr
	}

	logs, logsErr := w.collectTaskLogs(ctx, containerID)
	if logsErr != nil {
		w.logger.Warn("failed to collect task logs", "container_id", containerID, "error", logsErr)
	}

	return exitCode, logs, nil
}

func (w *PipelineWorker) newRunningStep(buildID uuid.UUID, stageStep domain.PipelineStep) *domain.BuildStep {
	now := time.Now()
	return &domain.BuildStep{
		ID:        uuid.New(),
		BuildID:   buildID,
		Name:      stageStep.Name,
		Image:     stageStep.Image,
		Commands:  stageStep.Commands,
		Status:    domain.BuildStatusRunning,
		StartedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (w *PipelineWorker) markStepSucceeded(ctx context.Context, step *domain.BuildStep) {
	w.markStepFinished(ctx, step, domain.BuildStatusSucceeded, 0)
}

func (w *PipelineWorker) markStepFinished(ctx context.Context, step *domain.BuildStep, status domain.BuildStatus, exitCode int) {
	finished := time.Now()
	step.Status = status
	step.FinishedAt = &finished
	step.UpdatedAt = finished
	step.ExitCode = &exitCode
	_ = w.repo.UpdateBuildStep(ctx, step)
}

func (w *PipelineWorker) markBuildSucceeded(ctx context.Context, build *domain.Build) {
	finish := time.Now()
	build.Status = domain.BuildStatusSucceeded
	build.FinishedAt = &finish
	build.UpdatedAt = finish
	_ = w.repo.UpdateBuild(ctx, build)
}

func (w *PipelineWorker) failStepAndBuild(ctx context.Context, step *domain.BuildStep, build *domain.Build, message string) {
	end := time.Now()
	step.Status = domain.BuildStatusFailed
	step.FinishedAt = &end
	step.UpdatedAt = end
	_ = w.repo.UpdateBuildStep(ctx, step)
	_ = w.repo.AppendBuildLog(ctx, &domain.BuildLog{ID: uuid.New(), BuildID: build.ID, StepID: step.ID, Content: message, CreatedAt: end})
	w.failBuild(ctx, build, message)
}

func (w *PipelineWorker) failBuild(ctx context.Context, build *domain.Build, message string) {
	finish := time.Now()
	build.Status = domain.BuildStatusFailed
	build.FinishedAt = &finish
	build.UpdatedAt = finish
	_ = w.repo.UpdateBuild(ctx, build)
	_ = w.repo.AppendBuildLog(ctx, &domain.BuildLog{ID: uuid.New(), BuildID: build.ID, Content: message, CreatedAt: finish})
}

func (w *PipelineWorker) collectTaskLogs(ctx context.Context, taskID string) (string, error) {
	reader, err := w.compute.GetInstanceLogs(ctx, taskID)
	if err != nil {
		return "", err
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
