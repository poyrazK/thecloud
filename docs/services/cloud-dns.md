# CloudDNS Service

CloudDNS is the managed DNS service for The Cloud. It provides authoritative DNS hosting and private DNS resolution within VPCs.

## Service Architecture

### Components
1. **API Handler**: Manages REST requests for zones and records.
2. **DNS Service**: Core business logic, validation, and auto-registration triggers.
3. **PostgreSQL Repository**: Stores metadata for zones and records with tenant isolation.
4. **PowerDNS Backend**: The authoritative name server that performs the actual DNS resolution.

### Data Flow
1. User requests record creation via CLI/API.
2. API validates the request (e.g., TTL >= 60s, valid record type).
3. Service records the entry in PostgreSQL for metadata tracking.
4. Service pushes the record to PowerDNS via its HTTP API.
5. PowerDNS serves the record to DNS recursors/clients.

## Implementation Details

### Auto-Registration
The `InstanceService` communicates with the `DNSService` during the provisioning phase. If an instance is launched in a VPC that has an "associated" zone (linked via `vpc_id` in `dns_zones` table), a new `A` record is created using the instance's name and allocated private IP.

### Multi-Tenancy
- **Auth**: All requests require a valid API Key.
- **Scoping**: Every query to the `dns_zones` and `dns_records` tables includes `WHERE tenant_id = $1`.
- **PowerDNS Isolation**: PowerDNS handles the global namespace, but The Cloud's logic ensures users can only manage zones they own.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `POWERDNS_API_URL` | Endpoint for PowerDNS API | `http://localhost:8081` |
| `POWERDNS_API_KEY` | Authentication key for PowerDNS | `thecloud-dns-secret` |
| `POWERDNS_SERVER_ID` | Server ID in PowerDNS | `localhost` |

## Resource Limits

- **Zones per Tenant**: Default 50 (soft limit)
- **Records per Zone**: Default 1000 (soft limit)
- **Minimum TTL**: 60 seconds
- **Max Name Length**: 255 characters
