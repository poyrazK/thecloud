package dns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestPowerDNSGeo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	backend, _ := NewPowerDNSBackend(ts.URL, "key", "localhost", nil)
	ctx := context.Background()
	hostname := "api.example.com"
	ip1 := "1.1.1.1"
	ip2 := "2.2.2.2"

	t.Run("CreateGeoRecord", func(t *testing.T) {
		endpoints := []domain.GlobalEndpoint{
			{Healthy: true, TargetType: "IP", TargetIP: &ip1},
			{Healthy: true, TargetType: "IP", TargetIP: &ip2},
			{Healthy: false, TargetType: "IP", TargetIP: &ip1}, // Should be ignored
		}

		err := backend.CreateGeoRecord(ctx, hostname, endpoints)
		assert.NoError(t, err)
	})

	t.Run("CreateGeoRecord Empty", func(t *testing.T) {
		// Should call DeleteGeoRecord
		err := backend.CreateGeoRecord(ctx, hostname, nil)
		assert.NoError(t, err)
	})

	t.Run("DeleteGeoRecord", func(t *testing.T) {
		err := backend.DeleteGeoRecord(ctx, hostname)
		assert.NoError(t, err)
	})

	t.Run("Invalid Hostname", func(t *testing.T) {
		err := backend.CreateGeoRecord(ctx, "invalid", nil)
		assert.Error(t, err)
		
		err = backend.DeleteGeoRecord(ctx, "invalid")
		assert.Error(t, err)
	})
}
