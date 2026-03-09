package dns

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/joeig/go-powerdns/v3"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// PowerDNSBackend implements DNSBackend using PowerDNS API.
type PowerDNSBackend struct {
	client      *powerdns.Client
	restyClient *resty.Client
	serverID    string
	logger      *slog.Logger
}

// NewPowerDNSBackend creates a new PowerDNS backend client.
func NewPowerDNSBackend(apiURL, apiKey, serverID string, logger *slog.Logger) (*PowerDNSBackend, error) {
	client := powerdns.New(apiURL, serverID, powerdns.WithAPIKey(apiKey))

	restyClient := resty.New().
		SetBaseURL(strings.TrimSuffix(apiURL, "/")+"/api/v1").
		SetHeader("X-API-Key", apiKey).
		SetHeader("Content-Type", "application/json")

	return &PowerDNSBackend{
		client:      client,
		restyClient: restyClient,
		serverID:    serverID,
		logger:      logger,
	}, nil
}

func (b *PowerDNSBackend) ensureTrailingDot(name string) string {
	if !strings.HasSuffix(name, ".") {
		return name + "."
	}
	return name
}

// CreateZone creates a new zone in PowerDNS.
func (b *PowerDNSBackend) CreateZone(ctx context.Context, zoneName string, nameservers []string) error {
	zoneName = b.ensureTrailingDot(zoneName)

	zone := &powerdns.Zone{
		Name:        powerdns.String(zoneName),
		Kind:        powerdns.ZoneKindPtr(powerdns.NativeZoneKind),
		Nameservers: nameservers,
	}

	_, err := b.client.Zones.Add(ctx, zone)
	if err != nil {
		b.logger.Error("failed to create zone in PowerDNS", "zone", zoneName, "error", err)
		return fmt.Errorf("failed to create zone: %w", err)
	}

	// Add an initial SOA record to satisfy PowerDNS requirements for Native zones
	soaContent := fmt.Sprintf("%s hostmaster.%s 1 10800 3600 604800 3600", nameservers[0], zoneName)
	soaRecord := ports.RecordSet{
		Name:    zoneName,
		Type:    "SOA",
		Records: []string{soaContent},
		TTL:     3600,
	}

	if err := b.AddRecords(ctx, zoneName, []ports.RecordSet{soaRecord}); err != nil {
		b.logger.Error("failed to add initial SOA record", "zone", zoneName, "error", err)
		return fmt.Errorf("failed to add initial SOA record: %w", err)
	}

	b.logger.Info("created zone in PowerDNS", "zone", zoneName)
	return nil
}

// DeleteZone removes a zone from PowerDNS.
func (b *PowerDNSBackend) DeleteZone(ctx context.Context, zoneName string) error {
	zoneName = b.ensureTrailingDot(zoneName)

	err := b.client.Zones.Delete(ctx, zoneName)
	if err != nil {
		b.logger.Error("failed to delete zone from PowerDNS", "zone", zoneName, "error", err)
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	b.logger.Info("deleted zone from PowerDNS", "zone", zoneName)
	return nil
}

// GetZone retrieves zone info from PowerDNS.
func (b *PowerDNSBackend) GetZone(ctx context.Context, zoneName string) (*ports.ZoneInfo, error) {
	zoneName = b.ensureTrailingDot(zoneName)

	z, err := b.client.Zones.Get(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	return &ports.ZoneInfo{
		Name:           *z.Name,
		Kind:           string(*z.Kind),
		Serial:         *z.Serial,
		NotifiedSerial: *z.NotifiedSerial,
	}, nil
}

// AddRecords adds records to a zone in PowerDNS.
func (b *PowerDNSBackend) AddRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	zoneName = b.ensureTrailingDot(zoneName)

	rrsets := make([]map[string]interface{}, len(records))
	for i, rec := range records {
		recordEntries := make([]map[string]interface{}, len(rec.Records))
		for j, content := range rec.Records {
			recordEntries[j] = map[string]interface{}{
				"content":  content,
				"disabled": false,
			}
		}

		rrsets[i] = map[string]interface{}{
			"name":       b.ensureTrailingDot(rec.Name),
			"type":       rec.Type,
			"ttl":        rec.TTL,
			"changetype": "REPLACE",
			"records":    recordEntries,
		}
	}

	payload := map[string]interface{}{
		"rrsets": rrsets,
	}

	resp, err := b.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		Patch(fmt.Sprintf("servers/%s/zones/%s", b.serverID, zoneName))

	if err != nil {
		return fmt.Errorf("failed to add records: %w", err)
	}

	if resp.IsError() {
		b.logger.Error("PowerDNS API error", "status", resp.StatusCode(), "body", resp.String())
		return fmt.Errorf("failed to add records: %s", resp.String())
	}

	return nil
}

// UpdateRecords updates (replaces) records in a zone.
func (b *PowerDNSBackend) UpdateRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	return b.AddRecords(ctx, zoneName, records)
}

// DeleteRecords removes records from a zone in PowerDNS.
func (b *PowerDNSBackend) DeleteRecords(ctx context.Context, zoneName, name, recordType string) error {
	zoneName = b.ensureTrailingDot(zoneName)
	name = b.ensureTrailingDot(name)

	payload := map[string]interface{}{
		"rrsets": []map[string]interface{}{
			{
				"name":       name,
				"type":       recordType,
				"changetype": "DELETE",
			},
		},
	}

	resp, err := b.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		Patch(fmt.Sprintf("servers/%s/zones/%s", b.serverID, zoneName))

	if err != nil {
		return fmt.Errorf("failed to delete records: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to delete records: %s", resp.String())
	}

	return nil
}

// ListRecords lists all records in a zone.
func (b *PowerDNSBackend) ListRecords(ctx context.Context, zoneName string) ([]ports.RecordSet, error) {
	zoneName = b.ensureTrailingDot(zoneName)

	z, err := b.client.Zones.Get(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone records: %w", err)
	}

	results := make([]ports.RecordSet, 0, len(z.RRsets))
	for _, rr := range z.RRsets {
		var content []string
		for _, r := range rr.Records {
			if r.Content != nil {
				content = append(content, *r.Content)
			}
		}

		results = append(results, ports.RecordSet{
			Name:    *rr.Name,
			Type:    string(*rr.Type),
			TTL:     int(*rr.TTL),
			Records: content,
		})
	}

	return results, nil
}
