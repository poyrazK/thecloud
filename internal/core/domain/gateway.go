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
	PathPrefix  string    `json:"path_prefix"`  // Request path to match (e.g., "/api/v1")
	TargetURL   string    `json:"target_url"`   // Internal destination (e.g., "http://service-a:8080")
	StripPrefix bool      `json:"strip_prefix"` // If true, removes path_prefix from request before forwarding
	RateLimit   int       `json:"rate_limit"`   // Maximum allowed requests per second per IP
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
