package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// Ensure PowerDNSBackend implements GeoDNSBackend
var _ ports.GeoDNSBackend = (*PowerDNSBackend)(nil)

const MaxDNSRecordTTL = 60

// Currently, this implements a Weighted Round-Robin strategy via standard A records
// for all healthy IP-based endpoints. Implementation of advanced LUA records
// for latency-based or geo-location routing is planned for future iterations.
func (b *PowerDNSBackend) CreateGeoRecord(ctx context.Context, hostname string, endpoints []domain.GlobalEndpoint) error {
	// 1. Identify Target Zone
	// Assumes a standard FQDN structure where the zone corresponds to the root domain.
	// Optimization: Implement a lookup service to precisely identify the authoritative zone.
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid hostname: %s", hostname)
	}
	// Simple heuristic: the zone is derived from the last two segments (e.g., example.com).
	zoneName := strings.Join(parts[len(parts)-2:], ".")

	// 2. Prepare Records
	var records []string
	for _, ep := range endpoints {
		if ep.Healthy {
			if ep.TargetType == "IP" && ep.TargetIP != nil && *ep.TargetIP != "" {
				records = append(records, *ep.TargetIP)
			}
			// Note: Load Balancer (LB) targets are temporarily deferred. These require
			// sophisticated CNAME or ALIAS record management to avoid conflicts at the zone root.
		}
	}

	if len(records) == 0 {
		return b.DeleteGeoRecord(ctx, hostname)
	}

	// 3. Commit Resource Record Set (RRSet)
	// Currently restricted to Type A records for IP-based resolution.
	rs := ports.RecordSet{
		Name:    hostname,
		Type:    "A",
		TTL:     MaxDNSRecordTTL, // Short TTL for dynamic responses
		Records: records,
	}

	return b.AddRecords(ctx, zoneName, []ports.RecordSet{rs})
}

// DeleteGeoRecord removes the GLB records.
func (b *PowerDNSBackend) DeleteGeoRecord(ctx context.Context, hostname string) error {
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid hostname: %s", hostname)
	}
	zoneName := strings.Join(parts[len(parts)-2:], ".")

	return b.DeleteRecords(ctx, zoneName, hostname, "A")
}
