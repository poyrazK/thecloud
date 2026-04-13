package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMEvaluator_Unit(t *testing.T) {
	t.Run("Evaluate_Allow", TestIAMEvaluator_Evaluate_Allow)
	t.Run("Evaluate_Deny", TestIAMEvaluator_Evaluate_Deny)
	t.Run("Evaluate_NoMatch", TestIAMEvaluator_Evaluate_NoMatch)
	t.Run("Evaluate_ExplicitDenyWins", TestIAMEvaluator_Evaluate_ExplicitDenyWins)
	t.Run("Evaluate_WildcardAction", TestIAMEvaluator_Evaluate_WildcardAction)
	t.Run("Evaluate_WildcardResource", TestIAMEvaluator_Evaluate_WildcardResource)
}

func TestIAMEvaluator_Evaluate_Allow(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectAllow,
					Action:   []string{"instance:launch", "instance:stop"},
					Resource: []string{"*"},
				},
			},
		},
	}

	effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "resource-1", nil)
	require.NoError(t, err)
	assert.Equal(t, domain.EffectAllow, effect)
}

func TestIAMEvaluator_Evaluate_Deny(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectDeny,
					Action:   []string{"instance:launch"},
					Resource: []string{"*"},
				},
			},
		},
	}

	effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "resource-1", nil)
	require.NoError(t, err)
	assert.Equal(t, domain.EffectDeny, effect)
}

func TestIAMEvaluator_Evaluate_NoMatch(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectAllow,
					Action:   []string{"volume:create"},
					Resource: []string{"*"},
				},
			},
		},
	}

	effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "resource-1", nil)
	require.NoError(t, err)
	assert.Equal(t, domain.PolicyEffect(""), effect)
}

func TestIAMEvaluator_Evaluate_ExplicitDenyWins(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectAllow,
					Action:   []string{"instance:launch"},
					Resource: []string{"*"},
				},
				{
					Effect:   domain.EffectDeny,
					Action:   []string{"instance:launch"},
					Resource: []string{"*"},
				},
			},
		},
	}

	effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "resource-1", nil)
	require.NoError(t, err)
	assert.Equal(t, domain.EffectDeny, effect)
}

func TestIAMEvaluator_Evaluate_WildcardAction(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectAllow,
					Action:   []string{"instance:*"},
					Resource: []string{"*"},
				},
			},
		},
	}

	tests := []struct {
		action   string
		expected domain.PolicyEffect
	}{
		{"instance:launch", domain.EffectAllow},
		{"instance:stop", domain.EffectAllow},
		{"instance:terminate", domain.EffectAllow},
	}

	for _, tt := range tests {
		effect, err := evaluator.Evaluate(ctx, policies, tt.action, "resource-1", nil)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, effect, "action %s should match", tt.action)
	}
}

func TestIAMEvaluator_Evaluate_WildcardResource(t *testing.T) {
	evaluator := services.NewIAMEvaluator()
	ctx := context.Background()

	policies := []*domain.Policy{
		{
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectAllow,
					Action:   []string{"instance:launch"},
					Resource: []string{"*"},
				},
			},
		},
	}

	effect, err := evaluator.Evaluate(ctx, policies, "instance:launch", "any-resource", nil)
	require.NoError(t, err)
	assert.Equal(t, domain.EffectAllow, effect)
}