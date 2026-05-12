package services

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMEvaluator_Evaluate_Conditions(t *testing.T) {
	evaluator := NewIAMEvaluator()
	ctx := context.Background()

	tests := []struct {
		name     string
		policies []*domain.Policy
		action   string
		resource string
		evalCtx  map[string]interface{}
		want     domain.PolicyEffect
	}{
		{
			name: "IpAddress condition - matching IP",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:  domain.EffectAllow,
							Action:  []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp": "192.168.1.50",
			},
			want: domain.EffectAllow,
		},
		{
			name: "IpAddress condition - non-matching IP",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:  domain.EffectAllow,
							Action:  []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp": "10.0.0.1",
			},
			want: "",
		},
		{
			name: "NotIpAddress condition - IP not in CIDR",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:  domain.EffectAllow,
							Action:  []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"NotIpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp": "10.0.0.1",
			},
			want: domain.EffectAllow,
		},
		{
			name: "StringEquals condition - matching tenant",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"StringEquals": {
									"thecloud:TenantId": "tenant-123",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"thecloud:TenantId": "tenant-123",
			},
			want: domain.EffectAllow,
		},
		{
			name: "StringEquals condition - non-matching tenant",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"StringEquals": {
									"thecloud:TenantId": "tenant-123",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"thecloud:TenantId": "tenant-456",
			},
			want: "",
		},
		{
			name: "StringLike condition - wildcard matching",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"StringLike": {
									"aws:UserId": "user-*",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:UserId": "user-abc-123",
			},
			want: domain.EffectAllow,
		},
		{
			name: "DateGreaterThan condition - current time after threshold",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"DateGreaterThan": {
									"aws:CurrentTime": "2020-01-01T00:00:00Z",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:CurrentTime": "2030-06-15T10:00:00Z",
			},
			want: domain.EffectAllow,
		},
		{
			name: "DateLessThan condition - current time before threshold",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"DateLessThan": {
									"aws:CurrentTime": "2099-01-01T00:00:00Z",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:CurrentTime": "2025-01-01T00:00:00Z",
			},
			want: domain.EffectAllow,
		},
		{
			name: "Bool condition - matching true",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"Bool": {
									"thecloud:IsAdmin": true,
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"thecloud:IsAdmin": true,
			},
			want: domain.EffectAllow,
		},
		{
			name: "Null condition - key does not exist",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"Null": {
									"thecloud:SomeKey": "true",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"otherKey": "value",
			},
			want: domain.EffectAllow,
		},
		{
			name: "Deny with condition - condition met",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectDeny,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp": "192.168.1.50",
			},
			want: domain.EffectDeny,
		},
		{
			name: "Multiple conditions - all must be met",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
								"StringEquals": {
									"thecloud:TenantId": "tenant-123",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp":      "192.168.1.50",
				"thecloud:TenantId": "tenant-123",
			},
			want: domain.EffectAllow,
		},
		{
			name: "Multiple conditions - one not met",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
								"StringEquals": {
									"thecloud:TenantId": "tenant-123",
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx: map[string]interface{}{
				"aws:SourceIp":      "192.168.1.50",
				"thecloud:TenantId": "tenant-456",
			},
			want: "",
		},
		{
			name: "No condition - works without evalCtx",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:   domain.EffectAllow,
							Action:   []string{"instance:*"},
							Resource: []string{"*"},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx:  nil,
			want:     domain.EffectAllow,
		},
		{
			name: "Condition with nil evalCtx - fails",
			policies: []*domain.Policy{
				{
					Statements: []domain.Statement{
						{
							Effect:  domain.EffectAllow,
							Action:  []string{"instance:*"},
							Resource: []string{"*"},
							Condition: domain.Condition{
								"IpAddress": {
									"aws:SourceIp": []interface{}{"192.168.1.0/24"},
								},
							},
						},
					},
				},
			},
			action:   "instance:launch",
			resource: "instance:123",
			evalCtx:  nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.Evaluate(ctx, tt.policies, tt.action, tt.resource, tt.evalCtx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIAMEvaluator_MatchPattern(t *testing.T) {
	evaluator := NewIAMEvaluator()

	tests := []struct {
		pattern string
		target  string
		want    bool
	}{
		{"*", "anything", true},
		{"instance:*", "instance:launch", true},
		{"instance:*", "instance:stop", true},
		{"instance:*", "other:launch", false},
		{"instance:launch", "instance:launch", true},
		{"instance:launch", "instance:stop", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.target, func(t *testing.T) {
			got := evaluator.matchPattern(tt.pattern, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}
