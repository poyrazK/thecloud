// Package services contains the core business logic of the application.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// VpcServiceParams defines dependencies for VpcService creation.
type VpcServiceParams struct {
	Repo           ports.VpcRepository
	LBRepo         ports.LBRepository
	PeeringRepo    ports.VPCPeeringRepository
	RouteTableRepo ports.RouteTableRepository
	AsRepo         ports.AutoScalingRepository
	RBACSvc        ports.RBACService
	Network        ports.NetworkBackend
	AuditSvc       ports.AuditService
	Logger         *slog.Logger
	DefaultCIDR    string
}

// VpcService handles the lifecycle of Virtual Private Clouds (VPCs),
// including network isolation through OVS bridges and backend persistence.
type VpcService struct {
	repo           ports.VpcRepository
	lbRepo         ports.LBRepository
	peeringRepo    ports.VPCPeeringRepository
	routeTableRepo ports.RouteTableRepository
	asRepo         ports.AutoScalingRepository
	rbacSvc        ports.RBACService
	network        ports.NetworkBackend
	auditSvc       ports.AuditService
	logger         *slog.Logger
	defaultCIDR    string
}

// NewVpcService creates a new instance of VpcService.
// If defaultCIDR is empty, it defaults to "10.0.0.0/16".
func NewVpcService(params VpcServiceParams) *VpcService {
	defaultCIDR := params.DefaultCIDR
	if defaultCIDR == "" {
		defaultCIDR = "10.0.0.0/16" // Fallback if not provided
	}
	logger := params.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &VpcService{
		repo:           params.Repo,
		lbRepo:         params.LBRepo,
		peeringRepo:    params.PeeringRepo,
		routeTableRepo: params.RouteTableRepo,
		asRepo:         params.AsRepo,
		rbacSvc:        params.RBACSvc,
		network:        params.Network,
		auditSvc:       params.AuditSvc,
		logger:         logger,
		defaultCIDR:    defaultCIDR,
	}
}

// CreateVPC provisions a new VPC with an associated OVS bridge for network isolation.
// It generates a unique VXLAN ID and persists the VPC metadata to the database.
func (s *VpcService) CreateVPC(ctx context.Context, name, cidrBlock string) (*domain.VPC, error) {
	ctx, span := otel.Tracer("vpc-service").Start(ctx, "CreateVPC")
	defer span.End()

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionVpcCreate, "*"); err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("vpc.name", name),
		attribute.String("vpc.cidr", cidrBlock),
	)

	if cidrBlock == "" {
		cidrBlock = s.defaultCIDR
	}

	// Validate CIDR format
	if _, _, err := net.ParseCIDR(cidrBlock); err != nil {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("invalid CIDR block format: %s", cidrBlock))
	}

	vpcID := uuid.New()
	// 1. Generate unique VNI (for demo purposes we use a hash based int)
	vxlanID := int(vpcID[0]) + 100

	// 2. Create OVS bridge
	bridgeName := fmt.Sprintf("br-vpc-%s", vpcID.String()[:8])
	if err := s.network.CreateBridge(ctx, bridgeName, vxlanID); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create OVS bridge", err)
	}

	// 3. Construct ARN
	arn := fmt.Sprintf("arn:thecloud:vpc:local:%s:vpc/%s", userID.String(), vpcID.String())

	// 4. Persist to DB
	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      name,
		CIDRBlock: cidrBlock,
		NetworkID: bridgeName,
		VXLANID:   vxlanID,
		Status:    "active",
		ARN:       arn,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, vpc); err != nil {
		// Cleanup OVS bridge if DB fails
		s.logger.Error("failed to create VPC in DB, rolling back bridge", "name", name, "error", err)
		if rbErr := s.network.DeleteBridge(ctx, bridgeName); rbErr != nil {
			s.logger.Error("failed to rollback bridge", "bridge", bridgeName, "error", rbErr)
		}
		return nil, errors.Wrap(errors.Internal, "failed to create VPC in database", err)
	}

	// 5. Create main route table with local route
	if s.routeTableRepo != nil {
		mainRT := &domain.RouteTable{
			ID:        uuid.New(),
			VPCID:     vpc.ID,
			Name:      "main",
			IsMain:    true,
			Routes:    []domain.Route{},
		}
		mainRT.Routes = append(mainRT.Routes, domain.Route{
			ID:              uuid.New(),
			RouteTableID:    mainRT.ID,
			DestinationCIDR: vpc.CIDRBlock,
			TargetType:     domain.RouteTargetLocal,
		})
		if err := s.routeTableRepo.Create(ctx, mainRT); err != nil {
			// Rollback: delete VPC
			s.logger.Error("failed to create main route table, rolling back VPC", "error", err)
			_ = s.repo.Delete(ctx, vpc.ID)
			_ = s.network.DeleteBridge(ctx, bridgeName)
			return nil, errors.Wrap(errors.Internal, "failed to create main route table", err)
		}
	}

	if err := s.auditSvc.Log(ctx, vpc.UserID, "vpc.create", "vpc", vpc.ID.String(), map[string]interface{}{
		"name":       vpc.Name,
		"cidr_block": vpc.CIDRBlock,
		"arn":        vpc.ARN,
	}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "vpc.create", "vpc_id", vpc.ID, "error", err)
	}

	return vpc, nil
}

// GetVPC retrieves a VPC by its unique identifier (UUID) or its name.
func (s *VpcService) GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionVpcRead, idOrName); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idOrName)
	if err == nil {
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.GetByName(ctx, idOrName)
}

// ListVPCs returns a list of all VPCs accessible by the current user.
func (s *VpcService) ListVPCs(ctx context.Context) ([]*domain.VPC, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionVpcRead, "*"); err != nil {
		return nil, err
	}

	return s.repo.List(ctx)
}

// DeleteVPC removes a VPC, its associated OVS bridge, and all related database records.
func (s *VpcService) DeleteVPC(ctx context.Context, idOrName string) error {
	s.logger.Info("DeleteVPC called", "idOrName", idOrName)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionVpcDelete, idOrName); err != nil {
		s.logger.Info("DeleteVPC: authorize failed", "idOrName", idOrName, "error", err)
		return err
	}

	vpc, err := s.GetVPC(ctx, idOrName)
	if err != nil {
		s.logger.Info("DeleteVPC: GetVPC failed", "idOrName", idOrName, "error", err)
		return err
	}

	s.logger.Info("DeleteVPC: starting dependency check", "vpcID", vpc.ID)
	// 1. Check for dependent resources
	if err := s.checkDeleteDependencies(ctx, vpc.ID); err != nil {
		s.logger.Info("DeleteVPC: dependency check failed", "vpcID", vpc.ID, "error", err)
		return err
	}

	// 2. Remove OVS bridge
	if err := s.network.DeleteBridge(ctx, vpc.NetworkID); err != nil {
		s.logger.Error("failed to remove OVS bridge", "bridge", vpc.NetworkID, "error", err)
		return errors.Wrap(errors.Internal, "failed to remove OVS bridge", err)
	}
	s.logger.Info("vpc bridge removed", "bridge", vpc.NetworkID)

	// 3. Delete from DB
	if err := s.repo.Delete(ctx, vpc.ID); err != nil {
		return errors.Wrap(errors.Internal, "failed to delete VPC from database", err)
	}

	if err := s.auditSvc.Log(ctx, vpc.UserID, "vpc.delete", "vpc", vpc.ID.String(), map[string]interface{}{
		"name": vpc.Name,
	}); err != nil {
		s.logger.Warn("failed to log audit event", "action", "vpc.delete", "vpc_id", vpc.ID, "error", err)
	}

	return nil
}

func (s *VpcService) checkDeleteDependencies(ctx context.Context, vpcID uuid.UUID) error {
	// Check for Load Balancers
	lbs, err := s.lbRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("check load balancers: %w", err)
	}
	for _, lb := range lbs {
		if lb == nil {
			s.logger.Warn("checkDeleteDependencies: nil LB in list, skipping")
			continue
		}
		if lb.VpcID != vpcID {
			continue
		}
		// Allow deletion if LB is being cleaned up (DELETED status).
		// The LBWorker runs every 5s to delete DELETED LBs from DB.
		// If LBWorker isn't running (ROLE=api), the LB record stays in DELETED
		// but won't block VPC deletion since we skip DELETED status here.
		if lb.Status == domain.LBStatusDeleted {
			continue
		}
		s.logger.Info("checkDeleteDependencies: blocking VPC delete due to LB", "lb_id", lb.ID, "lb_status", lb.Status)
		return errors.New(errors.Conflict, "cannot delete VPC: load balancers still exist")
	}

	// Check for active VPC peering connections
	if s.peeringRepo != nil {
		peerings, peerErr := s.peeringRepo.ListByVPC(ctx, vpcID)
		if peerErr != nil {
			return fmt.Errorf("check VPC peerings: %w", peerErr)
		}
		for _, p := range peerings {
			if p.Status == domain.PeeringStatusActive || p.Status == domain.PeeringStatusPendingAcceptance {
				return errors.New(errors.Conflict, "cannot delete VPC: active peering connections exist")
			}
		}
	}

	// Check for scaling groups linked to this VPC
	if s.asRepo != nil {
		count, asErr := s.asRepo.CountGroupsByVPC(ctx, vpcID)
		if asErr != nil {
			s.logger.Error("checkDeleteDependencies: scaling groups count failed", "error", asErr)
			return fmt.Errorf("check scaling groups: %w", asErr)
		}
		if count > 0 {
			return errors.New(errors.Conflict, "cannot delete VPC: scaling groups still exist")
		}
	}

	return nil
}
