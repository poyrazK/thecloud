package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewWSEvent_Success(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	event, err := NewWSEvent(WSEventInstanceCreated, map[string]string{"id": "i-1"}, tenantID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Type != WSEventInstanceCreated {
		t.Fatalf("unexpected event type: %s", event.Type)
	}
	if event.TenantID != tenantID {
		t.Fatalf("unexpected tenant id: got %s want %s", event.TenantID, tenantID)
	}
	if len(event.Payload) == 0 {
		t.Fatalf("expected payload to be populated")
	}
}

func TestNewWSEvent_InvalidPayload(t *testing.T) {
	t.Parallel()
	_, err := NewWSEvent(WSEventMetricUpdate, make(chan int), uuid.New())
	if err == nil {
		t.Fatalf("expected error for invalid payload")
	}
}
