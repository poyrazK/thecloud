package ws

import (
	"io"
	"log/slog"
	"testing"
	"time"
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
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
