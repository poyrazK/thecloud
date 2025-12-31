package domain

import (
	"time"

	"github.com/google/uuid"
)

type InstanceStatus string

const (
	StatusStarting InstanceStatus = "STARTING"
	StatusRunning  InstanceStatus = "RUNNING"
	StatusStopped  InstanceStatus = "STOPPED"
	StatusError    InstanceStatus = "ERROR"
	StatusDeleted  InstanceStatus = "DELETED"
)

type Instance struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Image     string         `json:"image"`
	Status    InstanceStatus `json:"status"`
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
