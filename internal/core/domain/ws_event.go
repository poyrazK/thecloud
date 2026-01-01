package domain

import (
	"encoding/json"
	"time"
)

// WSEventType defines the types of real-time events sent via WebSocket.
type WSEventType string

const (
	// Instance lifecycle events
	WSEventInstanceCreated    WSEventType = "INSTANCE_CREATED"
	WSEventInstanceStarted    WSEventType = "INSTANCE_STARTED"
	WSEventInstanceStopped    WSEventType = "INSTANCE_STOPPED"
	WSEventInstanceTerminated WSEventType = "INSTANCE_TERMINATED"

	// Resource events
	WSEventVolumeCreated  WSEventType = "VOLUME_CREATED"
	WSEventVolumeAttached WSEventType = "VOLUME_ATTACHED"
	WSEventVPCCreated     WSEventType = "VPC_CREATED"

	// Metrics events
	WSEventMetricUpdate WSEventType = "METRIC_UPDATE"

	// Audit events
	WSEventAuditLog WSEventType = "AUDIT_LOG"
)

// WSEvent represents a single WebSocket message sent to connected clients.
type WSEvent struct {
	Type      WSEventType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewWSEvent creates a new WebSocket event with the current timestamp.
func NewWSEvent(eventType WSEventType, payload interface{}) (*WSEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &WSEvent{
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now(),
	}, nil
}
