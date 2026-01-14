// Package domain defines core business entities.
package domain

import (
	"encoding/json"
	"time"
)

// WSEventType defines the types of real-time events sent via WebSocket.
type WSEventType string

const (
	// WSEventInstanceCreated is triggered when a new compute instance is successfully provisioned.
	WSEventInstanceCreated WSEventType = "INSTANCE_CREATED"
	// WSEventInstanceStarted is triggered when an instance enters the running state.
	WSEventInstanceStarted WSEventType = "INSTANCE_STARTED"
	// WSEventInstanceStopped is triggered when an instance enters the stopped state.
	WSEventInstanceStopped WSEventType = "INSTANCE_STOPPED"
	// WSEventInstanceTerminated is triggered when an instance is permanently removed.
	WSEventInstanceTerminated WSEventType = "INSTANCE_TERMINATED"

	// WSEventVolumeCreated is triggered after a storage volume is provisioned.
	WSEventVolumeCreated WSEventType = "VOLUME_CREATED"
	// WSEventVolumeAttached is triggered when a volume is successfully mounted to an instance.
	WSEventVolumeAttached WSEventType = "VOLUME_ATTACHED"
	// WSEventVPCCreated is triggered after a Virtual Private Cloud is established.
	WSEventVPCCreated WSEventType = "VPC_CREATED"

	// WSEventMetricUpdate is triggered when new performance metrics (CPU, RAM) are available.
	WSEventMetricUpdate WSEventType = "METRIC_UPDATE"

	// WSEventAuditLog is triggered whenever a new audit record is generated.
	WSEventAuditLog WSEventType = "AUDIT_LOG"
)

// WSEvent represents a single WebSocket message broadcasted to authenticated clients for UI updates.
type WSEvent struct {
	Type      WSEventType     `json:"type"`      // The classification of the event
	Payload   json.RawMessage `json:"payload"`   // Context-specific JSON data (e.g., an Instance or AuditLog object)
	Timestamp time.Time       `json:"timestamp"` // Actual time when the event was generated
}

// NewWSEvent wraps a payload into a WSEvent structure with the current system timestamp.
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
