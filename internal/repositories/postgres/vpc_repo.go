package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type VpcRepository struct {
	db *pgxpool.Pool
}

func NewVpcRepository(db *pgxpool.Pool) *VpcRepository {
	return &VpcRepository{db: db}
}

func (r *VpcRepository) Create(ctx context.Context, vpc *domain.VPC) error {
	query := `
		INSERT INTO vpcs (id, user_id, name, cidr_block, network_id, vxlan_id, status, arn, created_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::cidr, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, vpc.ID, vpc.UserID, vpc.Name, vpc.CIDRBlock, vpc.NetworkID, vpc.VXLANID, vpc.Status, vpc.ARN, vpc.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create vpc", err)
	}
	return nil
}

func (r *VpcRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, COALESCE(cidr_block::text, ''), network_id, vxlan_id, status, arn, created_at FROM vpcs WHERE id = $1 AND user_id = $2`
	var vpc domain.VPC
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&vpc.ID, &vpc.UserID, &vpc.Name, &vpc.CIDRBlock, &vpc.NetworkID, &vpc.VXLANID, &vpc.Status, &vpc.ARN, &vpc.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, fmt.Sprintf("vpc %s not found", id))
		}
		return nil, errors.Wrap(errors.Internal, "failed to get vpc", err)
	}
	return &vpc, nil
}

func (r *VpcRepository) GetByName(ctx context.Context, name string) (*domain.VPC, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, COALESCE(cidr_block::text, ''), network_id, vxlan_id, status, arn, created_at FROM vpcs WHERE name = $1 AND user_id = $2`
	var vpc domain.VPC
	err := r.db.QueryRow(ctx, query, name, userID).Scan(
		&vpc.ID, &vpc.UserID, &vpc.Name, &vpc.CIDRBlock, &vpc.NetworkID, &vpc.VXLANID, &vpc.Status, &vpc.ARN, &vpc.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, fmt.Sprintf("vpc name %s not found", name))
		}
		return nil, errors.Wrap(errors.Internal, "failed to get vpc by name", err)
	}
	return &vpc, nil
}

func (r *VpcRepository) List(ctx context.Context) ([]*domain.VPC, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, COALESCE(cidr_block::text, ''), network_id, vxlan_id, status, arn, created_at FROM vpcs WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list vpcs", err)
	}
	defer rows.Close()

	var vpcs []*domain.VPC
	for rows.Next() {
		var vpc domain.VPC
		err := rows.Scan(
			&vpc.ID, &vpc.UserID, &vpc.Name, &vpc.CIDRBlock, &vpc.NetworkID, &vpc.VXLANID, &vpc.Status, &vpc.ARN, &vpc.CreatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan vpc", err)
		}
		vpcs = append(vpcs, &vpc)
	}
	return vpcs, nil
}

func (r *VpcRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM vpcs WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete vpc", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "vpc not found")
	}
	return nil
}
