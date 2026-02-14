package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type iamRepository struct {
	db DB
}

// NewIAMRepository creates a new postgres-backed IAM repository.
func NewIAMRepository(db DB) *iamRepository {
	return &iamRepository{db: db}
}

func (r *iamRepository) CreatePolicy(ctx context.Context, policy *domain.Policy) error {
	statementsJSON, err := json.Marshal(policy.Statements)
	if err != nil {
		return fmt.Errorf("failed to marshal statements: %w", err)
	}

	query := `
		INSERT INTO policies (id, name, description, statements)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.Exec(ctx, query, policy.ID, policy.Name, policy.Description, statementsJSON)
	return err
}

func (r *iamRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	query := `SELECT id, name, description, statements FROM policies WHERE id = $1`
	var p domain.Policy
	var statementsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &statementsJSON)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(statementsJSON, &p.Statements); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statements: %w", err)
	}

	return &p, nil
}

func (r *iamRepository) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	query := `SELECT id, name, description, statements FROM policies`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		var statementsJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &statementsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(statementsJSON, &p.Statements); err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, nil
}

func (r *iamRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, "DELETE FROM policies WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}
	return nil
}

func (r *iamRepository) AttachPolicyToUser(ctx context.Context, userID uuid.UUID, policyID uuid.UUID) error {
	query := `INSERT INTO user_policies (user_id, policy_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID, policyID)
	return err
}

func (r *iamRepository) DetachPolicyFromUser(ctx context.Context, userID uuid.UUID, policyID uuid.UUID) error {
	query := `DELETE FROM user_policies WHERE user_id = $1 AND policy_id = $2`
	result, err := r.db.Exec(ctx, query, userID, policyID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy assignment not found")
	}
	return nil
}

func (r *iamRepository) GetPoliciesForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Policy, error) {
	query := `
		SELECT p.id, p.name, p.description, p.statements
		FROM policies p
		JOIN user_policies up ON p.id = up.policy_id
		WHERE up.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		var statementsJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &statementsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(statementsJSON, &p.Statements); err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, nil
}
