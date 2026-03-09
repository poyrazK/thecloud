package services

import (
	"context"
	"strings"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

type iamEvaluator struct{}

// NewIAMEvaluator creates a new instance of the policy evaluator.
func NewIAMEvaluator() *iamEvaluator {
	return &iamEvaluator{}
}

func (e *iamEvaluator) Evaluate(ctx context.Context, policies []*domain.Policy, action string, resource string, evalCtx map[string]interface{}) (bool, error) {
	allowFound := false

	for _, policy := range policies {
		for _, statement := range policy.Statements {
			// Skip statements with conditions for now
			if len(statement.Condition) > 0 {
				continue
			}

			if e.matches(statement, action, resource) {
				if statement.Effect == domain.EffectDeny {
					// Explicit Deny always wins
					return false, nil
				}
				if statement.Effect == domain.EffectAllow {
					allowFound = true
				}
			}
		}
	}

	return allowFound, nil
}

func (e *iamEvaluator) matches(statement domain.Statement, action string, resource string) bool {
	actionMatched := false
	for _, a := range statement.Action {
		if e.matchPattern(a, action) {
			actionMatched = true
			break
		}
	}

	if !actionMatched {
		return false
	}

	resourceMatched := false
	for _, r := range statement.Resource {
		if e.matchPattern(r, resource) {
			resourceMatched = true
			break
		}
	}

	return resourceMatched
}

func (e *iamEvaluator) matchPattern(pattern, target string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching: "instance:*" matches "instance:launch"
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(target, prefix)
	}

	return pattern == target
}
