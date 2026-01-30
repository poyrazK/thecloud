# Managed DNS Service (CloudDNS)

CloudDNS is a highly available and scalable Domain Name System (DNS) web service. It provides a reliable and cost-effective way to route users to internet applications by translating names like `www.example.com` into the numeric IP addresses like `192.0.2.1` that computers use to connect to each other.

The Cloud's DNA service is powered by **PowerDNS** and is fully integrated with our VPC and Compute services.

## Features

- **Multi-Tenant Isolation**: DNS zones are scoped to tenants, ensuring privacy and security.
- **VPC Integration**: Associate DNS zones with VPCs for private network resolution.
- **Automatic Instance Registration**: Instances launched in a VPC with an associated DNS zone automatically get an 'A' record (e.g., `web-01.inst.internal`).
- **Supported Record Types**:
    - **A**: IPv4 address records
    - **AAAA**: IPv6 address records
    - **CNAME**: Canonical name records (aliases)
    - **MX**: Mail exchange records
    - **TXT**: Text records for SPF, DKIM, etc.
- **TTL Management**: Configurable Time-To-Live for all records (min 60s).

## Getting Started

### 1. Create a DNS Zone

A zone is a container for DNS records for a specific domain name.

```bash
cloud dns create-zone example.com --description "Main production zone"
```

To create a **Private Zone** associated with a VPC:

```bash
cloud dns create-zone inst.internal --vpc-id <vpc-uuid>
```

### 2. Create DNS Records

Add records to your zone to route traffic.

**Add an A Record:**
```bash
cloud dns create-record <zone-uuid> --name "www" --type "A" --content "1.2.3.4" --ttl 300
```

**Add an MX Record:**
```bash
cloud dns create-record <zone-uuid> --name "@" --type "MX" --content "mail.example.com" --priority 10
```

### 3. Automatic Registration

When you launch an instance in a VPC that has an associated DNS zone, CloudDNS automatically creates an 'A' record for it:

1. Create a VPC: `cloud vpc create --name my-vpc`
2. Create a Zone for that VPC: `cloud dns create-zone internal.cloud --vpc-id <vpc-uuid>`
3. Launch an Instance: `cloud instance launch --name web-1 --vpc-id <vpc-uuid>`

Result: A record `web-1.internal.cloud` will be created automatically pointing to the instance's private IP.

## CLI Reference

| Command | Description |
|---------|-------------|
| `cloud dns list-zones` | List all DNS zones in your tenant |
| `cloud dns create-zone <name>` | Create a new DNS zone |
| `cloud dns get-zone <id>` | View zone details |
| `cloud dns delete-zone <id>` | Delete a zone and all its records |
| `cloud dns list-records <zone-id>` | List all records in a zone |
| `cloud dns create-record <zone-id>` | Add a new record |
| `cloud dns update-record <record-id>` | Update an existing record |
| `cloud dns delete-record <record-id>` | Remove a record |

## API Reference

Endpoints are prefixed with `/api/v1/dns`:

- `POST /zones`: Create a zone
- `GET /zones`: List zones
- `POST /zones/:id/records`: Create a record
- `GET /zones/:id/records`: List records in a zone
- `DELETE /records/:id`: Delete a record

For full details, see the [API Reference](../api-reference.md).
