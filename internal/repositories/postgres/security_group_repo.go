package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type SecurityGroupRepository struct {
	db DB
}

func NewSecurityGroupRepository(db DB) *SecurityGroupRepository {
	return &SecurityGroupRepository{db: db}
}

func (r *SecurityGroupRepository) Create(ctx context.Context, sg *domain.SecurityGroup) error {
	query := `
		INSERT INTO security_groups (id, user_id, vpc_id, name, description, arn, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, sg.ID, sg.UserID, sg.VPCID, sg.Name, sg.Description, sg.ARN, sg.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create security group", err)
	}
	return nil
}

func (r *SecurityGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SecurityGroup, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, description, arn, created_at FROM security_groups WHERE id = $1 AND user_id = $2`

	sg, err := r.scanSecurityGroup(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		return nil, err
	}

	rules, err := r.getRulesForGroup(ctx, sg.ID)
	if err != nil {
		return nil, err
	}
	sg.Rules = rules

	return sg, nil
}

func (r *SecurityGroupRepository) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.SecurityGroup, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, description, arn, created_at FROM security_groups WHERE vpc_id = $1 AND name = $2 AND user_id = $3`

	sg, err := r.scanSecurityGroup(r.db.QueryRow(ctx, query, vpcID, name, userID))
	if err != nil {
		return nil, err
	}

	rules, err := r.getRulesForGroup(ctx, sg.ID)
	if err != nil {
		return nil, err
	}
	sg.Rules = rules

	return sg, nil
}

func (r *SecurityGroupRepository) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.SecurityGroup, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, description, arn, created_at FROM security_groups WHERE vpc_id = $1 AND user_id = $2 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, vpcID, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list security groups", err)
	}
	return r.scanSecurityGroups(rows)
}

func (r *SecurityGroupRepository) scanSecurityGroup(row pgx.Row) (*domain.SecurityGroup, error) {
	var sg domain.SecurityGroup
	err := row.Scan(&sg.ID, &sg.UserID, &sg.VPCID, &sg.Name, &sg.Description, &sg.ARN, &sg.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "security group not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan security group", err)
	}
	return &sg, nil
}

func (r *SecurityGroupRepository) scanSecurityGroups(rows pgx.Rows) ([]*domain.SecurityGroup, error) {
	defer rows.Close()
	var groups []*domain.SecurityGroup
	for rows.Next() {
		sg, err := r.scanSecurityGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, sg)
	}
	return groups, nil
}

func (r *SecurityGroupRepository) AddRule(ctx context.Context, rule *domain.SecurityRule) error {
	query := `
		INSERT INTO security_rules (id, group_id, direction, protocol, port_min, port_max, cidr, priority, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, rule.ID, rule.GroupID, rule.Direction, rule.Protocol, rule.PortMin, rule.PortMax, rule.CIDR, rule.Priority, rule.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to add security rule", err)
	}
	return nil
}

func (r *SecurityGroupRepository) GetRuleByID(ctx context.Context, ruleID uuid.UUID) (*domain.SecurityRule, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT sr.id, sr.group_id, sr.direction, sr.protocol, sr.port_min, sr.port_max, sr.cidr, sr.priority, sr.created_at 
		FROM security_rules sr
		JOIN security_groups sg ON sr.group_id = sg.id
		WHERE sr.id = $1 AND sg.user_id = $2
	`
	rule, err := r.scanSecurityRule(r.db.QueryRow(ctx, query, ruleID, userID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "security rule not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get security rule", err)
	}
	return &rule, nil
}

func (r *SecurityGroupRepository) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		DELETE FROM security_rules sr
		USING security_groups sg
		WHERE sr.group_id = sg.id AND sr.id = $1 AND sg.user_id = $2
	`
	res, err := r.db.Exec(ctx, query, ruleID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "security rule not found or you do not own it")
	}
	return nil
}

func (r *SecurityGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM security_groups WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, id, userID)
	return err
}

func (r *SecurityGroupRepository) AddInstanceToGroup(ctx context.Context, instanceID, groupID uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to start transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := r.verifyOwnership(ctx, tx, userID, instanceID, groupID); err != nil {
		return err
	}

	query := `INSERT INTO instance_security_groups (instance_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err = tx.Exec(ctx, query, instanceID, groupID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to add instance to group", err)
	}

	return tx.Commit(ctx)
}

func (r *SecurityGroupRepository) RemoveInstanceFromGroup(ctx context.Context, instanceID, groupID uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to start transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := r.verifyOwnership(ctx, tx, userID, instanceID, groupID); err != nil {
		return err
	}

	query := `DELETE FROM instance_security_groups WHERE instance_id = $1 AND group_id = $2`
	_, err = tx.Exec(ctx, query, instanceID, groupID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to remove instance from group", err)
	}

	return tx.Commit(ctx)
}

func (r *SecurityGroupRepository) verifyOwnership(ctx context.Context, tx pgx.Tx, userID, instanceID, groupID uuid.UUID) error {
	var instanceOwner uuid.UUID
	err := tx.QueryRow(ctx, "SELECT user_id FROM instances WHERE id = $1", instanceID).Scan(&instanceOwner)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.New(errors.NotFound, "instance not found")
		}
		return errors.Wrap(errors.Internal, "failed to verify instance ownership", err)
	}
	if instanceOwner != userID {
		return errors.New(errors.Forbidden, "you do not own this instance")
	}

	var groupOwner uuid.UUID
	err = tx.QueryRow(ctx, "SELECT user_id FROM security_groups WHERE id = $1", groupID).Scan(&groupOwner)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.New(errors.NotFound, "security group not found")
		}
		return errors.Wrap(errors.Internal, "failed to verify security group ownership", err)
	}
	if groupOwner != userID {
		return errors.New(errors.Forbidden, "you do not own this security group")
	}
	return nil
}

func (r *SecurityGroupRepository) ListInstanceGroups(ctx context.Context, instanceID uuid.UUID) ([]*domain.SecurityGroup, error) {
	query := `
		SELECT sg.id, sg.user_id, sg.vpc_id, sg.name, sg.description, sg.arn, sg.created_at 
		FROM security_groups sg
		JOIN instance_security_groups isg ON sg.id = isg.group_id
		WHERE isg.instance_id = $1
	`
	rows, err := r.db.Query(ctx, query, instanceID)
	if err != nil {
		return nil, err
	}
	return r.scanSecurityGroups(rows)
}

func (r *SecurityGroupRepository) getRulesForGroup(ctx context.Context, groupID uuid.UUID) ([]domain.SecurityRule, error) {
	query := `SELECT id, group_id, direction, protocol, port_min, port_max, cidr, priority, created_at FROM security_rules WHERE group_id = $1 ORDER BY priority DESC`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	return r.scanSecurityRules(rows)
}

func (r *SecurityGroupRepository) scanSecurityRule(row pgx.Row) (domain.SecurityRule, error) {
	var rule domain.SecurityRule
	var direction string
	if err := row.Scan(&rule.ID, &rule.GroupID, &direction, &rule.Protocol, &rule.PortMin, &rule.PortMax, &rule.CIDR, &rule.Priority, &rule.CreatedAt); err != nil {
		return domain.SecurityRule{}, err
	}
	rule.Direction = domain.RuleDirection(direction)
	return rule, nil
}

func (r *SecurityGroupRepository) scanSecurityRules(rows pgx.Rows) ([]domain.SecurityRule, error) {
	defer rows.Close()
	var rules []domain.SecurityRule
	for rows.Next() {
		rule, err := r.scanSecurityRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
