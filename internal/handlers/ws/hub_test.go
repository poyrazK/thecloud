package ws

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

func TestHubRegisterUnregister(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 1)}

	hub.Register(client)
	waitForCondition(t, func() bool { return hub.ClientCount() == 1 })

	hub.Unregister(client)
	waitForCondition(t, func() bool { return hub.ClientCount() == 0 })

	select {
	case _, ok := <-client.send:
		if ok {
			t.Fatal("expected client send channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected client send channel to be closed")
	}
}

func TestHubBroadcast(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 1)}
	hub.Register(client)
	waitForCondition(t, func() bool { return hub.ClientCount() == 1 })

	payload := []byte("hello")
	hub.broadcast <- payload

	select {
	case msg := <-client.send:
		if string(msg) != string(payload) {
			t.Fatalf("unexpected broadcast payload: got %s want %s", string(msg), string(payload))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected to receive broadcast message")
	}
}

func waitForCondition(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestHubConcurrentBroadcastAndUnregister(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	go hub.Run()
	defer hub.Stop()

	client := &Client{
		hub:      hub,
		send:     make(chan []byte, 1),
		tenantID: uuid.New(),
		userID:   uuid.New().String(),
	}
	hub.Register(client)
	waitForCondition(t, func() bool { return hub.ClientCount() == 1 })

	// Signal all goroutines to start broadcasting at the same time
	startCh := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startCh
			for j := 0; j < 50; j++ {
				hub.BroadcastEvent(&domain.WSEvent{Type: "test"})
			}
		}()
	}

	close(startCh)
	// Small delay before unregister to allow goroutines to be in broadcast
	time.Sleep(5 * time.Millisecond)
	hub.Unregister(client)

	wg.Wait()
}
