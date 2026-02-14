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
	policy := &domain.Policy{
		ID:   policyID,
		Name: "TestPolicy",
		Statements: []domain.Statement{
			{Effect: domain.EffectAllow, Action: []string{"*"}},
		},
	}

	t.Run("CreatePolicy", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO policies").
			WithArgs(policy.ID, policy.Name, policy.Description, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreatePolicy(ctx, policy)
		assert.NoError(t, err)
	})

	t.Run("GetPolicyByID", func(t *testing.T) {
		statementsJSON, _ := json.Marshal(policy.Statements)
		rows := pgxmock.NewRows([]string{"id", "name", "description", "statements"}).
			AddRow(policy.ID, policy.Name, policy.Description, statementsJSON)

		mock.ExpectQuery("SELECT id, name, description, statements FROM policies").
			WithArgs(policy.ID).
			WillReturnRows(rows)

		fetched, err := repo.GetPolicyByID(ctx, policy.ID)
		assert.NoError(t, err)
		assert.Equal(t, policy.ID, fetched.ID)
		assert.Equal(t, policy.Name, fetched.Name)
	})

	t.Run("AttachPolicyToUser", func(t *testing.T) {
		userID := uuid.New()
		mock.ExpectExec("INSERT INTO user_policies").
			WithArgs(userID, policy.ID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.AttachPolicyToUser(ctx, userID, policy.ID)
		assert.NoError(t, err)
	})

	t.Run("GetPoliciesForUser", func(t *testing.T) {
		userID := uuid.New()
		statementsJSON, _ := json.Marshal(policy.Statements)
		rows := pgxmock.NewRows([]string{"id", "name", "description", "statements"}).
			AddRow(policy.ID, policy.Name, policy.Description, statementsJSON)

		mock.ExpectQuery("SELECT p.id, p.name, p.description, p.statements").
			WithArgs(userID).
			WillReturnRows(rows)

		userPolicies, err := repo.GetPoliciesForUser(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userPolicies, 1)
		assert.Equal(t, policy.ID, userPolicies[0].ID)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
