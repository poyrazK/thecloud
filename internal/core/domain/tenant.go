package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents an organization, team or project.
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"` // URL-friendly identifier
	OwnerID   uuid.UUID `json:"owner_id"`
	Plan      string    `json:"plan"`   // "free", "pro", "enterprise"
	Status    string    `json:"status"` // "active", "suspended"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantMember represents a user's membership in a tenant.
type TenantMember struct {
	TenantID uuid.UUID `json:"tenant_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"` // "owner", "admin", "member"
	JoinedAt time.Time `json:"joined_at"`
}

// TenantQuota defines resource limits for a tenant.
type TenantQuota struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	MaxInstances  int       `json:"max_instances"`
	UsedInstances int       `json:"used_instances"`
	MaxVPCs       int       `json:"max_vpcs"`
	UsedVPCs      int       `json:"used_vpcs"`
	MaxStorageGB  int       `json:"max_storage_gb"`
	UsedStorageGB int       `json:"used_storage_gb"`
	MaxMemoryGB   int       `json:"max_memory_gb"`
	UsedMemoryGB  int       `json:"used_memory_gb"`
	MaxVCPUs      int       `json:"max_vcpus"`
	UsedVCPUs     int       `json:"used_vcpus"`
}
