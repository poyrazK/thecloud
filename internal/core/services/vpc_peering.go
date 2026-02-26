// Package services implements core business workflows.
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

const (
	vpcPeeringTracer  = "vpc-peering-service"
	peeringFlowFormat = "ip,nw_dst=%s"
)

// VPCPeeringService manages VPC peering connection lifecycle,
// including CIDR validation and OVS flow rule programming.
type VPCPeeringService struct {
	repo     ports.VPCPeeringRepository
	vpcRepo  ports.VpcRepository
	network  ports.NetworkBackend
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// VPCPeeringServiceParams holds dependencies for VPCPeeringService.
type VPCPeeringServiceParams struct {
	Repo     ports.VPCPeeringRepository
	VpcRepo  ports.VpcRepository
	Network  ports.NetworkBackend
	AuditSvc ports.AuditService
	Logger   *slog.Logger
}

// NewVPCPeeringService constructs a VPCPeeringService with its dependencies.
func NewVPCPeeringService(params VPCPeeringServiceParams) *VPCPeeringService {
	return &VPCPeeringService{
		repo:     params.Repo,
		vpcRepo:  params.VpcRepo,
		network:  params.Network,
		auditSvc: params.AuditSvc,
		logger:   params.Logger,
	}
}

// CreatePeering initiates a peering connection request between two VPCs.
// Validates that both VPCs belong to the same tenant and have non-overlapping CIDRs.
func (s *VPCPeeringService) CreatePeering(ctx context.Context, requesterVPCID, accepterVPCID uuid.UUID) (*domain.VPCPeering, error) {
	ctx, span := otel.Tracer(vpcPeeringTracer).Start(ctx, "CreatePeering")
	defer span.End()

	span.SetAttributes(
		attribute.String("requester_vpc_id", requesterVPCID.String()),
		attribute.String("accepter_vpc_id", accepterVPCID.String()),
	)

	// 1. Self-peering guard
	if requesterVPCID == accepterVPCID {
		return nil, errors.New(errors.InvalidInput, "cannot peer a VPC with itself")
	}

	tenantID := appcontext.TenantIDFromContext(ctx)

	// 2. Get both VPCs and verify ownership
	requesterVPC, err := s.vpcRepo.GetByID(ctx, requesterVPCID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "requester VPC not found", err)
	}

	accepterVPC, err := s.vpcRepo.GetByID(ctx, accepterVPCID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "accepter VPC not found", err)
	}

	// 3. CIDR overlap check
	if err := validateNonOverlappingCIDRs(requesterVPC.CIDRBlock, accepterVPC.CIDRBlock); err != nil {
		return nil, err
	}

	// 4. Duplicate check
	existing, _ := s.repo.GetActiveByVPCPair(ctx, requesterVPCID, accepterVPCID)
	if existing != nil {
		return nil, errors.New(errors.Conflict, "an active or pending peering already exists between these VPCs")
	}

	// 5. Create peering record
	peeringID := uuid.New()
	arn := fmt.Sprintf("arn:thecloud:vpc-peering:local:%s:peering/%s", tenantID.String(), peeringID.String())

	peering := &domain.VPCPeering{
		ID:             peeringID,
		RequesterVPCID: requesterVPCID,
		AccepterVPCID:  accepterVPCID,
		TenantID:       tenantID,
		Status:         domain.PeeringStatusPendingAcceptance,
		ARN:            arn,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, peering); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create vpc peering", err)
	}

	userID := appcontext.UserIDFromContext(ctx)
	_ = s.auditSvc.Log(ctx, userID, "vpc_peering.create", "vpc_peering", peeringID.String(), map[string]interface{}{
		"requester_vpc_id": requesterVPCID.String(),
		"accepter_vpc_id":  accepterVPCID.String(),
	})

	s.logger.Info("vpc peering created",
		"peering_id", peeringID,
		"requester_vpc", requesterVPC.Name,
		"accepter_vpc", accepterVPC.Name,
	)

	return peering, nil
}

// AcceptPeering accepts a pending peering connection and programs OVS flow rules.
func (s *VPCPeeringService) AcceptPeering(ctx context.Context, peeringID uuid.UUID) (*domain.VPCPeering, error) {
	ctx, span := otel.Tracer(vpcPeeringTracer).Start(ctx, "AcceptPeering")
	defer span.End()

	span.SetAttributes(attribute.String("peering_id", peeringID.String()))

	peering, err := s.repo.GetByID(ctx, peeringID)
	if err != nil {
		return nil, err
	}

	if peering.Status != domain.PeeringStatusPendingAcceptance {
		return nil, errors.New(errors.InvalidInput, "only pending peering connections can be accepted")
	}

	// Get both VPCs for OVS flow programming
	requesterVPC, err := s.vpcRepo.GetByID(ctx, peering.RequesterVPCID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get requester VPC", err)
	}

	accepterVPC, err := s.vpcRepo.GetByID(ctx, peering.AccepterVPCID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get accepter VPC", err)
	}

	// Program OVS flow rules for cross-bridge routing
	if err := s.addPeeringFlows(ctx, requesterVPC, accepterVPC); err != nil {
		s.logger.Error("failed to program peering OVS flows", "peering_id", peeringID, "error", err)
		_ = s.repo.UpdateStatus(ctx, peeringID, domain.PeeringStatusFailed)
		return nil, errors.Wrap(errors.Internal, "failed to establish network peering", err)
	}

	// Update status to active
	if err := s.repo.UpdateStatus(ctx, peeringID, domain.PeeringStatusActive); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to activate peering", err)
	}

	peering.Status = domain.PeeringStatusActive

	userID := appcontext.UserIDFromContext(ctx)
	_ = s.auditSvc.Log(ctx, userID, "vpc_peering.accept", "vpc_peering", peeringID.String(), nil)

	s.logger.Info("vpc peering accepted",
		"peering_id", peeringID,
		"requester_vpc", requesterVPC.Name,
		"accepter_vpc", accepterVPC.Name,
	)

	return peering, nil
}

// RejectPeering rejects a pending peering connection request.
func (s *VPCPeeringService) RejectPeering(ctx context.Context, peeringID uuid.UUID) error {
	ctx, span := otel.Tracer(vpcPeeringTracer).Start(ctx, "RejectPeering")
	defer span.End()

	peering, err := s.repo.GetByID(ctx, peeringID)
	if err != nil {
		return err
	}

	if peering.Status != domain.PeeringStatusPendingAcceptance {
		return errors.New(errors.InvalidInput, "only pending peering connections can be rejected")
	}

	if err := s.repo.UpdateStatus(ctx, peeringID, domain.PeeringStatusRejected); err != nil {
		return errors.Wrap(errors.Internal, "failed to reject peering", err)
	}

	userID := appcontext.UserIDFromContext(ctx)
	_ = s.auditSvc.Log(ctx, userID, "vpc_peering.reject", "vpc_peering", peeringID.String(), nil)

	return nil
}

// DeletePeering tears down a peering connection and removes OVS flow rules.
func (s *VPCPeeringService) DeletePeering(ctx context.Context, peeringID uuid.UUID) error {
	ctx, span := otel.Tracer(vpcPeeringTracer).Start(ctx, "DeletePeering")
	defer span.End()

	peering, err := s.repo.GetByID(ctx, peeringID)
	if err != nil {
		return err
	}

	// If peering was active, clean up OVS flow rules
	if peering.Status == domain.PeeringStatusActive {
		requesterVPC, err := s.vpcRepo.GetByID(ctx, peering.RequesterVPCID)
		if err != nil {
			s.logger.Error("failed to get requester VPC during peering cleanup", "error", err)
		}

		accepterVPC, err := s.vpcRepo.GetByID(ctx, peering.AccepterVPCID)
		if err != nil {
			s.logger.Error("failed to get accepter VPC during peering cleanup", "error", err)
		}

		if requesterVPC != nil && accepterVPC != nil {
			if err := s.removePeeringFlows(ctx, requesterVPC, accepterVPC); err != nil {
				s.logger.Error("failed to remove peering OVS flows", "peering_id", peeringID, "error", err)
			}
		}
	}

	if err := s.repo.Delete(ctx, peeringID); err != nil {
		return errors.Wrap(errors.Internal, "failed to delete peering", err)
	}

	userID := appcontext.UserIDFromContext(ctx)
	_ = s.auditSvc.Log(ctx, userID, "vpc_peering.delete", "vpc_peering", peeringID.String(), nil)

	s.logger.Info("vpc peering deleted", "peering_id", peeringID)
	return nil
}

// GetPeering retrieves details of a specific peering connection.
func (s *VPCPeeringService) GetPeering(ctx context.Context, peeringID uuid.UUID) (*domain.VPCPeering, error) {
	return s.repo.GetByID(ctx, peeringID)
}

// ListPeerings returns all peering connections for the current tenant.
func (s *VPCPeeringService) ListPeerings(ctx context.Context) ([]*domain.VPCPeering, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	return s.repo.List(ctx, tenantID)
}

// addPeeringFlows programs OVS flow rules to allow traffic between two VPC bridges.
func (s *VPCPeeringService) addPeeringFlows(ctx context.Context, requesterVPC, accepterVPC *domain.VPC) error {
	// Flow on requester bridge: route traffic destined for accepter's CIDR
	requesterFlow := ports.FlowRule{
		Priority: 500,
		Match:    fmt.Sprintf(peeringFlowFormat, accepterVPC.CIDRBlock),
		Actions:  "NORMAL",
	}
	if err := s.network.AddFlowRule(ctx, requesterVPC.NetworkID, requesterFlow); err != nil {
		return fmt.Errorf("failed to add flow on requester bridge %s: %w", requesterVPC.NetworkID, err)
	}

	// Flow on accepter bridge: route traffic destined for requester's CIDR
	accepterFlow := ports.FlowRule{
		Priority: 500,
		Match:    fmt.Sprintf(peeringFlowFormat, requesterVPC.CIDRBlock),
		Actions:  "NORMAL",
	}
	if err := s.network.AddFlowRule(ctx, accepterVPC.NetworkID, accepterFlow); err != nil {
		// Rollback first flow
		_ = s.network.DeleteFlowRule(ctx, requesterVPC.NetworkID, requesterFlow.Match)
		return fmt.Errorf("failed to add flow on accepter bridge %s: %w", accepterVPC.NetworkID, err)
	}

	s.logger.Info("peering OVS flows programmed",
		"requester_bridge", requesterVPC.NetworkID,
		"accepter_bridge", accepterVPC.NetworkID,
	)

	return nil
}

// removePeeringFlows removes OVS flow rules for a peering connection.
func (s *VPCPeeringService) removePeeringFlows(ctx context.Context, requesterVPC, accepterVPC *domain.VPC) error {
	// Remove flow from requester bridge
	requesterMatch := fmt.Sprintf(peeringFlowFormat, accepterVPC.CIDRBlock)
	if err := s.network.DeleteFlowRule(ctx, requesterVPC.NetworkID, requesterMatch); err != nil {
		s.logger.Error("failed to remove flow from requester bridge", "bridge", requesterVPC.NetworkID, "error", err)
	}

	// Remove flow from accepter bridge
	accepterMatch := fmt.Sprintf(peeringFlowFormat, requesterVPC.CIDRBlock)
	if err := s.network.DeleteFlowRule(ctx, accepterVPC.NetworkID, accepterMatch); err != nil {
		s.logger.Error("failed to remove flow from accepter bridge", "bridge", accepterVPC.NetworkID, "error", err)
	}

	s.logger.Info("peering OVS flows removed",
		"requester_bridge", requesterVPC.NetworkID,
		"accepter_bridge", accepterVPC.NetworkID,
	)

	return nil
}

// validateNonOverlappingCIDRs checks that two CIDR blocks do not overlap.
func validateNonOverlappingCIDRs(cidr1, cidr2 string) error {
	_, net1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return errors.New(errors.InvalidInput, "invalid requester VPC CIDR block")
	}

	_, net2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return errors.New(errors.InvalidInput, "invalid accepter VPC CIDR block")
	}

	// Check if either network contains the other's start address
	if net1.Contains(net2.IP) || net2.Contains(net1.IP) {
		return errors.New(errors.InvalidInput, "VPC CIDR blocks overlap; peering requires non-overlapping address spaces")
	}

	return nil
}
