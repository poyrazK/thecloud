package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"gopkg.in/yaml.v3"
)

type stackService struct {
	repo        ports.StackRepository
	instanceSvc ports.InstanceService
	vpcSvc      ports.VpcService
	logger      *slog.Logger
}

func NewStackService(
	repo ports.StackRepository,
	instanceSvc ports.InstanceService,
	vpcSvc ports.VpcService,
	logger *slog.Logger,
) *stackService {
	return &stackService{
		repo:        repo,
		instanceSvc: instanceSvc,
		vpcSvc:      vpcSvc,
		logger:      logger,
	}
}

type Template struct {
	Resources map[string]ResourceDefinition `yaml:"Resources"`
}

type ResourceDefinition struct {
	Type       string                 `yaml:"Type"`
	Properties map[string]interface{} `yaml:"Properties"`
}

func (s *stackService) CreateStack(ctx context.Context, name, templateStr string, parameters map[string]string) (*domain.Stack, error) {
	userID := appcontext.UserIDFromContext(ctx)

	paramsJSON, _ := json.Marshal(parameters)
	stack := &domain.Stack{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       name,
		Template:   templateStr,
		Parameters: paramsJSON,
		Status:     domain.StackStatusCreateInProgress,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, stack); err != nil {
		return nil, err
	}

	// Process in background
	go s.processStack(stack)

	return stack, nil
}

func (s *stackService) processStack(stack *domain.Stack) {
	ctx := context.Background()
	ctx = appcontext.WithUserID(ctx, stack.UserID)

	var t Template
	if err := yaml.Unmarshal([]byte(stack.Template), &t); err != nil {
		s.updateStackStatus(ctx, stack, domain.StackStatusCreateFailed, "Invalid template YAML")
		return
	}

	logicalToPhysical := make(map[string]uuid.UUID)

	// Simple non-topological sort for now: Create VPCs first, then everything else
	// A real implementation would build a dependency graph

	// Pass 1: VPCs
	for logicalID, res := range t.Resources {
		if res.Type == "VPC" {
			id, err := s.createVPC(ctx, stack.ID, logicalID, res.Properties)
			if err != nil {
				s.updateStackStatus(ctx, stack, domain.StackStatusCreateFailed, fmt.Sprintf("Failed to create VPC %s: %v", logicalID, err))
				return
			}
			logicalToPhysical[logicalID] = id
		}
	}

	// Pass 2: Everything else
	for logicalID, res := range t.Resources {
		if res.Type == "VPC" {
			continue
		}

		// Resolve references (Ref: LogicalID)
		props := s.resolveRefs(res.Properties, logicalToPhysical)

		if res.Type == "Instance" {
			_, err := s.createInstance(ctx, stack.ID, logicalID, props)
			if err != nil {
				s.updateStackStatus(ctx, stack, domain.StackStatusCreateFailed, fmt.Sprintf("Failed to create Instance %s: %v", logicalID, err))
				return
			}
		}
	}

	s.updateStackStatus(ctx, stack, domain.StackStatusCreateComplete, "")
}

func (s *stackService) resolveRefs(props map[string]interface{}, refs map[string]uuid.UUID) map[string]interface{} {
	newProps := make(map[string]interface{})
	for k, v := range props {
		if m, ok := v.(map[string]interface{}); ok {
			if ref, exists := m["Ref"]; exists {
				if refID, ok := ref.(string); ok {
					if physicalID, found := refs[refID]; found {
						newProps[k] = physicalID
						continue
					}
				}
			}
		}
		newProps[k] = v
	}
	return newProps
}

func (s *stackService) createVPC(ctx context.Context, stackID uuid.UUID, logicalID string, props map[string]interface{}) (uuid.UUID, error) {
	name, _ := props["Name"].(string)
	if name == "" {
		name = fmt.Sprintf("%s-%s", logicalID, stackID.String()[:8])
	}
	cidr, _ := props["CIDR"].(string)

	vpc, err := s.vpcSvc.Create(ctx, name, cidr)
	if err != nil {
		return uuid.Nil, err
	}

	_ = s.repo.AddResource(ctx, &domain.StackResource{
		ID:           uuid.New(),
		StackID:      stackID,
		LogicalID:    logicalID,
		PhysicalID:   vpc.ID.String(),
		ResourceType: "VPC",
		Status:       "CREATE_COMPLETE",
		CreatedAt:    time.Now(),
	})

	return vpc.ID, nil
}

func (s *stackService) createInstance(ctx context.Context, stackID uuid.UUID, logicalID string, props map[string]interface{}) (uuid.UUID, error) {
	name, _ := props["Name"].(string)
	image, _ := props["Image"].(string)
	cpu, _ := props["CPU"].(int)
	mem, _ := props["Memory"].(int)
	vpcID, _ := props["VpcID"].(uuid.UUID)

	if name == "" {
		name = fmt.Sprintf("%s-%s", logicalID, stackID.String()[:8])
	}

	inst, err := s.instanceSvc.Launch(ctx, name, image, cpu, mem, vpcID, nil)
	if err != nil {
		return uuid.Nil, err
	}

	_ = s.repo.AddResource(ctx, &domain.StackResource{
		ID:           uuid.New(),
		StackID:      stackID,
		LogicalID:    logicalID,
		PhysicalID:   inst.ID.String(),
		ResourceType: "Instance",
		Status:       "CREATE_COMPLETE",
		CreatedAt:    time.Now(),
	})

	return inst.ID, nil
}

func (s *stackService) updateStackStatus(ctx context.Context, stack *domain.Stack, status domain.StackStatus, reason string) {
	stack.Status = status
	stack.StatusReason = reason
	stack.UpdatedAt = time.Now()
	_ = s.repo.Update(ctx, stack)
}

func (s *stackService) GetStack(ctx context.Context, id uuid.UUID) (*domain.Stack, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *stackService) ListStacks(ctx context.Context) ([]*domain.Stack, error) {
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.ListByUserID(ctx, userID)
}

func (s *stackService) DeleteStack(ctx context.Context, id uuid.UUID) error {
	// 1. Get stack
	stack, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Perform background deletion
	go func() {
		bgCtx := context.Background()
		bgCtx = appcontext.WithUserID(bgCtx, stack.UserID)

		resources, _ := s.repo.ListResources(bgCtx, id)

		// Delete resources in reverse order (naive)
		for i := len(resources) - 1; i >= 0; i-- {
			res := resources[i]
			physID, err := uuid.Parse(res.PhysicalID)
			if err != nil {
				continue
			}

			switch res.ResourceType {
			case "Instance":
				_ = s.instanceSvc.Terminate(bgCtx, physID)
			case "VPC":
				_ = s.vpcSvc.Delete(bgCtx, physID)
			}
		}

		_ = s.repo.Delete(bgCtx, id)
	}()

	return nil
}

func (s *stackService) ValidateTemplate(ctx context.Context, template string) (*domain.TemplateValidateResponse, error) {
	var t Template
	if err := yaml.Unmarshal([]byte(template), &t); err != nil {
		return &domain.TemplateValidateResponse{
			Valid:  false,
			Errors: []string{fmt.Sprintf("YAML parse error: %v", err)},
		}, nil
	}

	if len(t.Resources) == 0 {
		return &domain.TemplateValidateResponse{
			Valid:  false,
			Errors: []string{"Template must contain at least one resource"},
		}, nil
	}

	return &domain.TemplateValidateResponse{Valid: true}, nil
}
