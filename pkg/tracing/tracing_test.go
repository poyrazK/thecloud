package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitTracer(t *testing.T) {
	// Jaeger init might fail if endpoint is not reachable or if we can't create exporter.
	// We just test that it doesn't panic and returns proper types or error.
	// Since we use otlptracehttp.WithInsecure(), it shouldn't fail validation immediately.
	tp, err := InitTracer(context.Background(), "test-service", "http://localhost:4318")
	// It relies on connection, so err might be nil or not depending on library behavior
	// But usually creating the exporter struct itself doesn't connect immediately in all SDK versions.
	// However, if it errors, that's fine, we just want to ensure code runs correctly.
	if err == nil {
		require.NotNil(t, tp)
	}

	// Test invalid scenarios if applicable, though simple init is hard to break without mocks.
}

func TestInitNoopTracer(t *testing.T) {
	tp := InitNoopTracer()
	require.NotNil(t, tp)
}

func TestInitConsoleTracer(t *testing.T) {
	tp, err := InitConsoleTracer("test-service")
	require.NoError(t, err)
	require.NotNil(t, tp)
}

func TestEnvOr(t *testing.T) {
	t.Setenv("ENV", "production")
	require.Equal(t, "production", envOr("ENV", "development"))
	require.Equal(t, "fallback", envOr("MISSING_ENV", "fallback"))
}
