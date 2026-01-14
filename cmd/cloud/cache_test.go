package main

import "testing"

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(0); got != "0 B" {
		t.Fatalf("expected 0 B, got %q", got)
	}
	if got := formatBytes(1024); got != "1.0 KB" {
		t.Fatalf("expected 1.0 KB, got %q", got)
	}
	if got := formatBytes(1024 * 1024); got != "1.0 MB" {
		t.Fatalf("expected 1.0 MB, got %q", got)
	}
}
