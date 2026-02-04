package domain

import "testing"

func TestNewWSEvent_Success(t *testing.T) {
	t.Parallel()
	event, err := NewWSEvent(WSEventInstanceCreated, map[string]string{"id": "i-1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Type != WSEventInstanceCreated {
		t.Fatalf("unexpected event type: %s", event.Type)
	}
	if len(event.Payload) == 0 {
		t.Fatalf("expected payload to be populated")
	}
}

func TestNewWSEvent_InvalidPayload(t *testing.T) {
	t.Parallel()
	_, err := NewWSEvent(WSEventMetricUpdate, make(chan int))
	if err == nil {
		t.Fatalf("expected error for invalid payload")
	}
}
