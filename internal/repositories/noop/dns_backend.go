package noop

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

type NoopDNSBackend struct{}

func NewNoopDNSBackend() ports.DNSBackend {
	return &NoopDNSBackend{}
}

func (b *NoopDNSBackend) CreateZone(ctx context.Context, zoneName string, nameservers []string) error {
	return nil
}

func (b *NoopDNSBackend) DeleteZone(ctx context.Context, zoneName string) error {
	return nil
}

func (b *NoopDNSBackend) GetZone(ctx context.Context, zoneName string) (*ports.ZoneInfo, error) {
	return &ports.ZoneInfo{Name: zoneName, Kind: "Native"}, nil
}

func (b *NoopDNSBackend) AddRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	return nil
}

func (b *NoopDNSBackend) UpdateRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	return nil
}

func (b *NoopDNSBackend) DeleteRecords(ctx context.Context, zoneName string, name string, recordType string) error {
	return nil
}

func (b *NoopDNSBackend) ListRecords(ctx context.Context, zoneName string) ([]ports.RecordSet, error) {
	return []ports.RecordSet{}, nil
}
