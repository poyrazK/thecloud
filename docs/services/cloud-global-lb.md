# Global Load Balancer (GLB)

## Overview
The **Global Load Balancer** (GLB) provides multi-region traffic distribution at the DNS level. Unlike regional load balancers that operate at the network layer within a VPC, GLB utilizes GeoDNS to steer users to the optimal regional endpoint based on policies like latency, health, and weight.

## Features
- **Global Traffic Steering**: Route traffic across multiple regions.
- **Health-Aware Routing**: Automatically removes unhealthy regional endpoints from DNS resolution.
- **Policy-Based Distribution**:
    - **Latency**: Directs users to the region with the lowest network latency.
    - **Geolocation**: Routes traffic based on the user's geographic location.
    - **Weighted**: Distributes traffic proportionally across regions.
    - **Failover**: Priority-based failover for disaster recovery.
- **Unified Hostname**: Provide a single global hostname for your application (e.g., `api.global.example.com`).

## Architecture
GLB follows the standard hexadecimal architecture of the platform:

1.  **Service Layer**: Handles the business logic of GLB creation, endpoint management, and synchronization with DNS.
2.  **GeoDNS Adapter**: Communicates with the authoritative DNS server (PowerDNS) to update records in real-time.
3.  **Repository**: Stores metadata in PostgreSQL about GLB configurations and endpoint health status.

## Configuration
### Health Checks
GLB performs synthesized health checks from multiple points of presence. Configuration includes:
- **Protocol**: HTTP, HTTPS, or TCP.
- **Port**: The destination port to probe.
- **Interval**: Frequency of probes (default 30s).
- **Thresholds**: Number of consecutive successes/failures to change health status.

- **Regional Load Balancers**: Linked by ID to existing platform resource. **Security**: GLB verifies that the regional LB belongs to the same user.
- **External IPs**: Arbitrary static IPs for hybrid-cloud scenarios.

## Security & Multi-tenancy
GLB is a multi-tenant service. Security is enforced at several layers:
- **Data Isolation**: Users can only see and manage Global Load Balancers they have created. `List` and `Get` operations are scoped to the authenticated user's ID.
- **Resource Ownership Verification**: When adding a Regional Load Balancer as an endpoint, the system verifies that the target LB belongs to the user attempting to add it. This prevents unauthorized traffic steering of other users' resources.

## Resource Synchronization
The service ensures that the authoritative DNS state always reflects the database state:
- **Transactional Consistency**: Changes in the database are followed by immediate synchronization calls to the GeoDNS backend.
- **Removal Logic**: Deleting an endpoint or the entire GLB automatically triggers the removal of associated DNS records, preventing orphaned routing entries.

## API Usage
Create a GLB:
```bash
cloud global-lb create --name "prod-api" --hostname "api.global.com" --policy "LATENCY"
```

Add a regional endpoint:
```bash
cloud global-lb add-endpoint --id <glb-id> --region "us-east-1" --target-ip "1.2.3.4" --weight 100
```
