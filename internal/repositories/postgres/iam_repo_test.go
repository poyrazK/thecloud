package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMRepository_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewIAMRepository(mock)
	ctx := context.Background()

	policyID := uuid.New()
	tenantID := uuid.New()
	policy := &domain.Policy{
		ID:       policyID,
		TenantID: tenantID,
		Name:     "TestPolicy",
		Statements: []domain.Statement{
			{Effect: domain.EffectAllow, Action: []string{"*"}},
		},
	}

	t.Run("CreatePolicy", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO policies").
			WithArgs(policy.ID, tenantID, policy.Name, policy.Description, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreatePolicy(ctx, tenantID, policy)
		assert.NoError(t, err)
	})

	t.Run("GetPolicyByID", func(t *testing.T) {
		statementsJSON, err := json.Marshal(policy.Statements)
		require.NoError(t, err)
		rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "statements"}).
			AddRow(policy.ID, policy.TenantID, policy.Name, policy.Description, statementsJSON)

		mock.ExpectQuery("SELECT id, tenant_id, name, description, statements FROM policies").
			WithArgs(policy.ID, policy.TenantID).
			WillReturnRows(rows)

		fetched, err := repo.GetPolicyByID(ctx, policy.TenantID, policy.ID)
		assert.NoError(t, err)
		assert.Equal(t, policy.ID, fetched.ID)
		assert.Equal(t, policy.Name, fetched.Name)
	})

	t.Run("AttachPolicyToUser", func(t *testing.T) {
		userID := uuid.New()
		
		// AttachPolicyToUser calls GetPolicyByID first
		statementsJSON, _ := json.Marshal(policy.Statements)
		policyRows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "statements"}).
			AddRow(policy.ID, policy.TenantID, policy.Name, policy.Description, statementsJSON)
		mock.ExpectQuery("SELECT id, tenant_id, name, description, statements FROM policies").
			WithArgs(policy.ID, tenantID).
			WillReturnRows(policyRows)

		mock.ExpectExec("INSERT INTO user_policies").
			WithArgs(userID, policy.ID, tenantID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.AttachPolicyToUser(ctx, tenantID, userID, policy.ID)
		assert.NoError(t, err)
	})

	t.Run("GetPoliciesForUser", func(t *testing.T) {
		userID := uuid.New()
		statementsJSON, _ := json.Marshal(policy.Statements)
		rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "statements"}).
			AddRow(policy.ID, policy.TenantID, policy.Name, policy.Description, statementsJSON)

		mock.ExpectQuery("SELECT p.id, p.tenant_id, p.name, p.description, p.statements").
			WithArgs(userID, tenantID).
			WillReturnRows(rows)

		userPolicies, err := repo.GetPoliciesForUser(ctx, tenantID, userID)
		assert.NoError(t, err)
		assert.Len(t, userPolicies, 1)
		assert.Equal(t, policy.ID, userPolicies[0].ID)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
