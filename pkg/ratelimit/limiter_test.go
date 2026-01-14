package ratelimit_test

import (
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/pkg/ratelimit"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"golang.org/x/time/rate"
)

func BenchmarkRateLimiterGetLimiter(b *testing.B) {
	limiter := ratelimit.NewIPRateLimiter(rate.Limit(100), 10, slog.Default())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.GetLimiter(testutil.TestIPPrivate)
	}
}

func BenchmarkRateLimiterGetLimiterParallel(b *testing.B) {
	limiter := ratelimit.NewIPRateLimiter(rate.Limit(1000), 100, slog.Default())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = limiter.GetLimiter(testutil.TestIPPrivate)
		}
	})
}

func BenchmarkRateLimiterGetLimiterParallelMultiKey(b *testing.B) {
	limiter := ratelimit.NewIPRateLimiter(rate.Limit(1000), 100, slog.Default())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Simulate different clients
			key := "key-" + string(rune(i%100))
			_ = limiter.GetLimiter(key)
			i++
		}
	})
}
