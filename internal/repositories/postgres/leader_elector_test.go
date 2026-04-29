package postgres

import (
	"testing"
)

func TestKeyToLockIDDeterministic(t *testing.T) {
	key := "singleton:lb"
	id1 := keyToLockID(key)
	id2 := keyToLockID(key)
	if id1 != id2 {
		t.Fatalf("expected same lock ID for same key, got %d and %d", id1, id2)
	}
}

func TestKeyToLockIDUnique(t *testing.T) {
	keys := []string{
		"singleton:lb",
		"singleton:cron",
		"singleton:autoscaling",
		"singleton:container",
		"singleton:healing",
		"singleton:db-failover",
		"singleton:cluster-reconciler",
		"singleton:replica-monitor",
		"singleton:lifecycle",
		"singleton:log",
		"singleton:accounting",
	}

	seen := make(map[int64]string)
	for _, k := range keys {
		id := keyToLockID(k)
		if id <= 0 {
			t.Fatalf("expected positive lock ID for key %q, got %d", k, id)
		}
		if existing, ok := seen[id]; ok {
			t.Fatalf("lock ID collision: key %q and %q both map to %d", k, existing, id)
		}
		seen[id] = k
	}
}

func TestKeyToLockIDPositive(t *testing.T) {
	// Ensure the masking produces positive values
	testKeys := []string{"a", "b", "test", "singleton:anything", ""}
	for _, k := range testKeys {
		id := keyToLockID(k)
		if id < 0 {
			t.Fatalf("expected non-negative lock ID for key %q, got %d", k, id)
		}
	}
}
