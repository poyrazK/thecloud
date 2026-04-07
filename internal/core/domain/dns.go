package domain

import (
	"time"

	"github.com/google/uuid"
)

// DNSZone represents a private DNS zone linked to a VPC.
type DNSZone struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	VpcID       uuid.UUID `json:"vpc_id"` // Required for private zones
	Name        string    `json:"name"`   // e.g., "myapp.internal"
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`      // ACTIVE, CREATING, DELETING, FAILED
	DefaultTTL  int       `json:"default_ttl"` // Default: 300
	PowerDNSID  string    `json:"powerdns_id"` // Zone ID in PowerDNS (e.g., "myapp.internal.")
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Zone status constants
const (
	ZoneStatusCreating = "CREATING"
	ZoneStatusActive   = "ACTIVE"
	ZoneStatusDeleting = "DELETING"
	ZoneStatusFailed   = "FAILED"
)

// DNSRecord represents a DNS record within a zone.
type DNSRecord struct {
	ID          uuid.UUID  `json:"id"`
	ZoneID      uuid.UUID  `json:"zone_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`    // e.g., "www", "api", "@"
	Type        RecordType `json:"type"`    // A, AAAA, CNAME, MX, TXT, SRV
	Content     string     `json:"content"` // IP, hostname, or value
	TTL         int        `json:"ttl"`
	Priority    *int       `json:"priority,omitempty"` // For MX, SRV
	Disabled    bool       `json:"disabled"`
	AutoManaged bool       `json:"auto_managed"`          // True for instance auto-registration
	InstanceID  *uuid.UUID `json:"instance_id,omitempty"` // Link to instance if auto-managed
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RecordType defines supported DNS record types.
type RecordType string

const (
	RecordTypeA     RecordType = "A"
	RecordTypeAAAA  RecordType = "AAAA"
	RecordTypeCNAME RecordType = "CNAME"
	RecordTypeMX    RecordType = "MX"
	RecordTypeTXT   RecordType = "TXT"
	RecordTypeSRV   RecordType = "SRV"
)

// ValidRecordTypes returns all valid record types.
func ValidRecordTypes() []RecordType {
	return []RecordType{RecordTypeA, RecordTypeAAAA, RecordTypeCNAME, RecordTypeMX, RecordTypeTXT, RecordTypeSRV}
}

// IsValidRecordType checks if a record type is supported.
func IsValidRecordType(t RecordType) bool {
	for _, valid := range ValidRecordTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
