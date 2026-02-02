package mock

import (
	"context"
	"log/slog"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type MockGeoDNS struct {
	Records map[string][]domain.GlobalEndpoint
}

func NewMockGeoDNS() ports.GeoDNSBackend {
	return &MockGeoDNS{
		Records: make(map[string][]domain.GlobalEndpoint),
	}
}

func (m *MockGeoDNS) CreateGeoRecord(ctx context.Context, hostname string, endpoints []domain.GlobalEndpoint) error {
	slog.Info("GeoDNS: Creating record", "hostname", hostname, "endpoints", len(endpoints))
	m.Records[hostname] = endpoints
	return nil
}

func (m *MockGeoDNS) DeleteGeoRecord(ctx context.Context, hostname string) error {
	slog.Info("GeoDNS: Deleting record", "hostname", hostname)
	delete(m.Records, hostname)
	return nil
}
