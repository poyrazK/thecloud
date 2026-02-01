// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// GatewayRoute defines an ingress rule for mapping external HTTP traffic to internal resources.
type GatewayRoute struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	PathPrefix  string    `json:"path_prefix"`  // Legacy: Request path to match (e.g., "/api/v1")
	PathPattern string    `json:"path_pattern"` // New: Pattern with {params}
	PatternType string    `json:"pattern_type"` // "prefix" or "pattern"
	ParamNames  []string  `json:"param_names"`  // Extracted parameter names
	TargetURL   string    `json:"target_url"`   // Internal destination (e.g., "http://service-a:8080")
	Methods     []string  `json:"methods"`      // New: HTTP methods to match (empty = all)
	StripPrefix bool      `json:"strip_prefix"` // If true, removes path_prefix from request before forwarding
	RateLimit   int       `json:"rate_limit"`   // Maximum allowed requests per second per IP
	Priority    int       `json:"priority"`     // Manual priority for tie-breaking
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RouteMatch represents a successful route pattern match.
type RouteMatch struct {
	Route      *GatewayRoute
	Params     map[string]string // Extracted path parameters
	MatchScore int               // Specificity score
}
