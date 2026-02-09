// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ContainerWorker reconciles container deployments and instances.
type ContainerWorker struct {
	repo        ports.ContainerRepository
	instanceSvc ports.InstanceService
	eventSvc    ports.EventService
}

// NewContainerWorker constructs a ContainerWorker with its dependencies.
func NewContainerWorker(repo ports.ContainerRepository, instanceSvc ports.InstanceService, eventSvc ports.EventService) *ContainerWorker {
	return &ContainerWorker{
		repo:        repo,
		instanceSvc: instanceSvc,
		eventSvc:    eventSvc,
	}
}

func (w *ContainerWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Println("CloudContainers Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("CloudContainers Worker stopping")
			return
		case <-ticker.C:
			w.Reconcile(ctx)
		}
	}
}

func (w *ContainerWorker) Reconcile(ctx context.Context) {
	deployments, err := w.repo.ListAllDeployments(ctx)
	if err != nil {
		log.Printf("ContainerWorker: failed to list deployments: %v", err)
		return
	}

	for _, dep := range deployments {
		w.reconcileDeployment(ctx, dep)
	}
}

func (w *ContainerWorker) reconcileDeployment(ctx context.Context, dep *domain.Deployment) {
	// Wrap context with user ID
	uCtx := appcontext.WithUserID(ctx, dep.UserID)

	containerIDs, err := w.repo.GetContainers(uCtx, dep.ID)
	if err != nil {
		log.Printf("ContainerWorker: failed to get containers for %s: %v", dep.Name, err)
		return
	}

	// Filter out unhealthy or missing instances
	var healthyContainerIDs []uuid.UUID
	for _, id := range containerIDs {
		inst, err := w.instanceSvc.GetInstance(uCtx, id.String())
		if err != nil || inst.Status == domain.StatusError || inst.Status == domain.StatusDeleted {
			log.Printf("ContainerWorker: instance %s for deployment %s is unhealthy or missing, removing from group", id, dep.Name)
			_ = w.repo.RemoveContainer(uCtx, dep.ID, id)
			continue
		}
		healthyContainerIDs = append(healthyContainerIDs, id)
	}

	if w.handleDeletingDeployment(uCtx, dep, healthyContainerIDs) {
		return
	}

	current := len(healthyContainerIDs)
	w.scaleDeployment(uCtx, dep, healthyContainerIDs, current)
	w.updateDeploymentStatus(uCtx, dep, current)
}

func (w *ContainerWorker) handleDeletingDeployment(ctx context.Context, dep *domain.Deployment, containerIDs []uuid.UUID) bool {
	if dep.Status != domain.DeploymentStatusDeleting {
		return false
	}
	if len(containerIDs) == 0 {
		_ = w.repo.DeleteDeployment(ctx, dep.ID)
		return true
	}
	for _, id := range containerIDs {
		_ = w.terminateContainer(ctx, dep, id)
	}
	return true
}

func (w *ContainerWorker) scaleDeployment(ctx context.Context, dep *domain.Deployment, containerIDs []uuid.UUID, current int) {
	if current < dep.Replicas {
		w.launchMissingContainers(ctx, dep, dep.Replicas-current)
		return
	}
	if current > dep.Replicas {
		w.terminateExcessContainers(ctx, dep, containerIDs, current-dep.Replicas)
	}
}

func (w *ContainerWorker) launchMissingContainers(ctx context.Context, dep *domain.Deployment, count int) {
	for i := 0; i < count; i++ {
		if err := w.launchContainer(ctx, dep); err != nil {
			log.Printf("ContainerWorker: failed to launch container for %s: %v", dep.Name, err)
			return
		}
	}
}

func (w *ContainerWorker) terminateExcessContainers(ctx context.Context, dep *domain.Deployment, containerIDs []uuid.UUID, count int) {
	for i := 0; i < count; i++ {
		if err := w.terminateContainer(ctx, dep, containerIDs[i]); err != nil {
			log.Printf("ContainerWorker: failed to terminate container for %s: %v", dep.Name, err)
			return
		}
	}
}

func (w *ContainerWorker) updateDeploymentStatus(ctx context.Context, dep *domain.Deployment, current int) {
	newStatus := domain.DeploymentStatusReady
	if current != dep.Replicas {
		newStatus = domain.DeploymentStatusScaling
	}

	if dep.Status != newStatus || dep.CurrentCount != current {
		dep.Status = newStatus
		dep.CurrentCount = current
		_ = w.repo.UpdateDeployment(ctx, dep)
	}
}

func (w *ContainerWorker) launchContainer(ctx context.Context, dep *domain.Deployment) error {
	name := fmt.Sprintf("dep-%s-%d", dep.Name, time.Now().UnixNano())

	// Deployments usually run in a default VPC or we could add VPC support to deployments
	// For now using nil VPC (default network)
	inst, err := w.instanceSvc.LaunchInstance(ctx, ports.LaunchParams{
		Name:         name,
		Image:        dep.Image,
		Ports:        dep.Ports,
		InstanceType: dep.InstanceType,
	})
	if err != nil {
		return err
	}

	if err := w.repo.AddContainer(ctx, dep.ID, inst.ID); err != nil {
		// Cleanup instance if association fails
		_ = w.instanceSvc.TerminateInstance(ctx, inst.ID.String())
		return err
	}

	return nil
}

func (w *ContainerWorker) terminateContainer(ctx context.Context, dep *domain.Deployment, instanceID uuid.UUID) error {
	if err := w.repo.RemoveContainer(ctx, dep.ID, instanceID); err != nil {
		return err
	}

	return w.instanceSvc.TerminateInstance(ctx, instanceID.String())
}
