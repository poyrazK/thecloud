package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMEvaluator_Evaluate(t *testing.T) {
	evaluator := NewIAMEvaluator()
	ctx := context.Background()

	policy1 := &domain.Policy{
		ID:   uuid.New(),
		Name: "AllowCompute",
		Statements: []domain.Statement{
			{
				Effect:   domain.EffectAllow,
				Action:   []string{"instance:*"},
				Resource: []string{"*"},
			},
		},
	}

	policy2 := &domain.Policy{
		ID:   uuid.New(),
		Name: "DenySpecificInstance",
		Statements: []domain.Statement{
			{
				Effect:   domain.EffectDeny,
				Action:   []string{"instance:terminate"},
				Resource: []string{"arn:thecloud:compute:instance:prod-123"},
			},
		},
	}

	policies := []*domain.Policy{policy1, policy2}

	t.Run("AllowByWildcard", func(t *testing.T) {
		effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "any", nil)
		require.NoError(t, err)
		assert.Equal(t, domain.EffectAllow, effect)
	})

	t.Run("ExplicitDenyWins", func(t *testing.T) {
		effect, err := evaluator.Evaluate(ctx, policies, "instance:terminate", "arn:thecloud:compute:instance:prod-123", nil)
		require.NoError(t, err)
		assert.Equal(t, domain.EffectDeny, effect)
	})

	t.Run("AllowOtherInstanceTerminate", func(t *testing.T) {
		effect, err := evaluator.Evaluate(ctx, policies, "instance:terminate", "arn:thecloud:compute:instance:dev-456", nil)
		require.NoError(t, err)
		assert.Equal(t, domain.EffectAllow, effect)
	})

	t.Run("DefaultDeny", func(t *testing.T) {
		effect, err := evaluator.Evaluate(ctx, policies, "vpc:create", "*", nil)
		require.NoError(t, err)
		assert.Equal(t, domain.PolicyEffect(""), effect)
	})
}
