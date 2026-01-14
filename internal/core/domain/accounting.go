// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ResourceType defines the category of a cloud resource for billing purposes.
type ResourceType string

const (
	// ResourceInstance represents compute instances (VMs or containers).
	ResourceInstance ResourceType = "INSTANCE"
	// ResourceStorage represents block storage volumes or object storage.
	ResourceStorage ResourceType = "STORAGE"
	// ResourceNetwork represents networking resources like IPs or bandwidth.
	ResourceNetwork ResourceType = "NETWORK"
)

// UsageRecord represents a single unit of resource consumption.
// It tracks how much of a specific resource was used over a time period.
type UsageRecord struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	ResourceID   uuid.UUID    `json:"resource_id"`
	ResourceType ResourceType `json:"resource_type"`
	Quantity     float64      `json:"quantity"` // Amount consumed (e.g. 60)
	Unit         string       `json:"unit"`     // Unit of measure (e.g. "minutes")
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
}

// BillSummary aggregates usage costs for a user over a specific billing period.
type BillSummary struct {
	UserID      uuid.UUID                `json:"user_id"`
	TotalAmount float64                  `json:"total_amount"`
	Currency    string                   `json:"currency"` // ISO 4217 code (e.g. "USD")
	UsageByType map[ResourceType]float64 `json:"usage_by_type"`
	PeriodStart time.Time                `json:"period_start"`
	PeriodEnd   time.Time                `json:"period_end"`
}
