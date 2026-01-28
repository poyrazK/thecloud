// Package helpers provides shared test helpers for E2E tests.
package helpers

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RunConcurrently runs the given function n times in parallel
func RunConcurrently(n int, fn func(i int) error) []error {
	var wg sync.WaitGroup
	results := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = fn(idx)
		}(i)
	}

	wg.Wait()
	return results
}

// AssertOnlyOneSucceeds verifies that exactly one operation in the set succeeded
func AssertOnlyOneSucceeds(t *testing.T, results []error) {
	t.Helper()
	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount, "Expected exactly one operation to succeed, got %d", successCount)
}

// AssertAllFail verifies that all operations in the set failed
func AssertAllFail(t *testing.T, results []error) {
	t.Helper()
	for i, err := range results {
		assert.Error(t, err, "Operation %d should have failed", i)
	}
}
