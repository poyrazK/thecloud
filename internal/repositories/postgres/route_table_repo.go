// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	stdlib_errors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// RouteTableRepository provides PostgreSQL-backed persistence for route tables.
type RouteTableRepository struct {
	db DB
}

// NewRouteTableRepository creates a RouteTableRepository using the provided DB.
func NewRouteTableRepository(db DB) *RouteTableRepository {
	return &RouteTableRepository{db: db}
}

// Create inserts a new route table record into the database.
func (r *RouteTableRepository) Create(ctx context.Context, rt *domain.RouteTable) error {
	query := `
		INSERT INTO route_tables (id, vpc_id, name, is_main, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, rt.ID, rt.VPCID, rt.Name, rt.IsMain, rt.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create route table", err)
	}
	return nil
}

// GetByID retrieves a single route table by its UUID.
func (r *RouteTableRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RouteTable, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT rt.id, rt.vpc_id, rt.name, rt.is_main, rt.created_at
		FROM route_tables rt
		JOIN vpcs v ON rt.vpc_id = v.id
		WHERE rt.id = $1 AND v.tenant_id = $2
	`
	rt, err := r.scanRouteTable(r.db.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		return nil, err
	}

	routes, err := r.ListRoutes(ctx, id)
	if err != nil {
		return nil, err
	}
	rt.Routes = routes

	associations, err := r.ListAssociatedSubnets(ctx, id)
	if err != nil {
		return nil, err
	}
	rt.Associations = make([]domain.RouteTableAssociation, len(associations))
	for i, subnetID := range associations {
		rt.Associations[i] = domain.RouteTableAssociation{
			ID:           uuid.New(),
			RouteTableID: id,
			SubnetID:     subnetID,
		}
	}

	return rt, nil
}

// GetByVPC returns all route tables for a given VPC.
func (r *RouteTableRepository) GetByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.RouteTable, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT rt.id, rt.vpc_id, rt.name, rt.is_main, rt.created_at
		FROM route_tables rt
		JOIN vpcs v ON rt.vpc_id = v.id
		WHERE rt.vpc_id = $1 AND v.tenant_id = $2
		ORDER BY rt.is_main DESC, rt.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, vpcID, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list route tables", err)
	}
	return r.scanRouteTables(rows)
}

// GetMainByVPC returns the main route table for a given VPC.
func (r *RouteTableRepository) GetMainByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.RouteTable, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT rt.id, rt.vpc_id, rt.name, rt.is_main, rt.created_at
		FROM route_tables rt
		JOIN vpcs v ON rt.vpc_id = v.id
		WHERE rt.vpc_id = $1 AND rt.is_main = TRUE AND v.tenant_id = $2
	`
	return r.scanRouteTable(r.db.QueryRow(ctx, query, vpcID, tenantID))
}

// Update modifies an existing route table.
func (r *RouteTableRepository) Update(ctx context.Context, rt *domain.RouteTable) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		UPDATE route_tables SET name = $1, is_main = $2
		WHERE id = $3 AND vpc_id IN (SELECT id FROM vpcs WHERE tenant_id = $4)
	`
	cmd, err := r.db.Exec(ctx, query, rt.Name, rt.IsMain, rt.ID, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update route table", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "route table not found")
	}
	return nil
}

// Delete removes a route table from the database.
func (r *RouteTableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		DELETE FROM route_tables rt
		USING vpcs v
		WHERE rt.vpc_id = v.id AND rt.id = $1 AND v.tenant_id = $2 AND rt.is_main = FALSE
	`
	cmd, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete route table", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "route table not found or is a main table")
	}
	return nil
}

// AddRoute adds a new route to a route table.
func (r *RouteTableRepository) AddRoute(ctx context.Context, rtID uuid.UUID, route *domain.Route) error {
	query := `
		INSERT INTO routes (id, route_table_id, destination_cidr, target_type, target_id, target_name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, route.ID, rtID, route.DestinationCIDR, route.TargetType, route.TargetID, route.TargetName, route.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to add route", err)
	}
	return nil
}

// RemoveRoute removes a route from a route table.
func (r *RouteTableRepository) RemoveRoute(ctx context.Context, rtID, routeID uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		DELETE FROM routes r
		USING route_tables rt
		JOIN vpcs v ON rt.vpc_id = v.id
		WHERE r.route_table_id = rt.id AND r.id = $1 AND rt.id = $2 AND v.tenant_id = $3
	`
	cmd, err := r.db.Exec(ctx, query, routeID, rtID, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to remove route", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "route not found")
	}
	return nil
}

// ListRoutes returns all routes for a given route table.
func (r *RouteTableRepository) ListRoutes(ctx context.Context, rtID uuid.UUID) ([]domain.Route, error) {
	query := `SELECT id, route_table_id, destination_cidr, target_type, COALESCE(target_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(target_name, ''), created_at FROM routes WHERE route_table_id = $1 ORDER BY destination_cidr`
	rows, err := r.db.Query(ctx, query, rtID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list routes", err)
	}
	return r.scanRoutes(rows)
}

// AssociateSubnet links a subnet to a route table.
func (r *RouteTableRepository) AssociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error {
	query := `
		INSERT INTO route_table_associations (id, route_table_id, subnet_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (route_table_id, subnet_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, uuid.New(), rtID, subnetID, "now")
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to associate subnet", err)
	}
	return nil
}

// DisassociateSubnet removes a subnet's association with a route table.
func (r *RouteTableRepository) DisassociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		DELETE FROM route_table_associations rta
		USING route_tables rt
		JOIN vpcs v ON rt.vpc_id = v.id
		WHERE rta.route_table_id = rt.id AND rta.subnet_id = $1 AND rt.id = $2 AND v.tenant_id = $3
	`
	_, err := r.db.Exec(ctx, query, subnetID, rtID, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to disassociate subnet", err)
	}
	return nil
}

// ListAssociatedSubnets returns all subnet IDs associated with a route table.
func (r *RouteTableRepository) ListAssociatedSubnets(ctx context.Context, rtID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT subnet_id FROM route_table_associations WHERE route_table_id = $1`
	rows, err := r.db.Query(ctx, query, rtID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list associated subnets", err)
	}
	defer rows.Close()

	var subnetIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan subnet id", err)
		}
		subnetIDs = append(subnetIDs, id)
	}
	return subnetIDs, nil
}

func (r *RouteTableRepository) scanRouteTable(row pgx.Row) (*domain.RouteTable, error) {
	var rt domain.RouteTable
	err := row.Scan(&rt.ID, &rt.VPCID, &rt.Name, &rt.IsMain, &rt.CreatedAt)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "route table not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan route table", err)
	}
	return &rt, nil
}

func (r *RouteTableRepository) scanRouteTables(rows pgx.Rows) ([]*domain.RouteTable, error) {
	defer rows.Close()
	var tables []*domain.RouteTable
	for rows.Next() {
		rt, err := r.scanRouteTable(rows)
		if err != nil {
			return nil, err
		}
		tables = append(tables, rt)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to iterate route tables", err)
	}
	return tables, nil
}

func (r *RouteTableRepository) scanRoutes(rows pgx.Rows) ([]domain.Route, error) {
	defer rows.Close()
	var routes []domain.Route
	for rows.Next() {
		var route domain.Route
		var targetID *uuid.UUID
		if err := rows.Scan(&route.ID, &route.RouteTableID, &route.DestinationCIDR, &route.TargetType, &targetID, &route.TargetName, &route.CreatedAt); err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan route", err)
		}
		// Convert zero UUID back to nil
		if targetID != nil && *targetID == uuid.Nil {
			targetID = nil
		}
		route.TargetID = targetID
		routes = append(routes, route)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to iterate routes", err)
	}
	return routes, nil
}
