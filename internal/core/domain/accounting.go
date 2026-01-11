package domain

import (
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	ResourceInstance ResourceType = "INSTANCE"
	ResourceStorage  ResourceType = "STORAGE"
	ResourceNetwork  ResourceType = "NETWORK"
)

type UsageRecord struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	ResourceID   uuid.UUID    `json:"resource_id"`
	ResourceType ResourceType `json:"resource_type"`
	Quantity     float64      `json:"quantity"` // e.g., minutes, GB-hours
	Unit         string       `json:"unit"`     // e.g., "minute", "gb-hour"
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
}

type BillSummary struct {
	UserID      uuid.UUID                `json:"user_id"`
	TotalAmount float64                  `json:"total_amount"`
	Currency    string                   `json:"currency"`
	UsageByType map[ResourceType]float64 `json:"usage_by_type"`
	PeriodStart time.Time                `json:"period_start"`
	PeriodEnd   time.Time                `json:"period_end"`
}
