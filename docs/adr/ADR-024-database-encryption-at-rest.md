# ADR 024: Database Encryption at Rest

## Status
Accepted

## Context
Managed databases in The Cloud platform store sensitive customer data on persistent block volumes. While the platform already supports:
- Encrypted secrets using AES-256-GCM (ADR 020)
- Vault integration for credential storage
- TLS for data in transit

There was no mechanism to encrypt the underlying volume data at rest. This is a critical compliance requirement (SOC 2, GDPR, HIPAA) and a significant security gap for production DBaaS offerings.

## Decision
We decided to implement application-level encryption at rest for managed database volumes using HashiCorp Vault Transit Secrets Engine for DEK (Data Encryption Key) management.

### Architecture

**Encryption Approach**: Application-level AES-256-GCM with Vault Transit for DEK management. This is transparent to the database engine and works across all storage backends (Docker volumes, LVM).

**DEK Pattern**:
- Each volume has a unique 256-bit DEK (Data Encryption Key)
- DEKs are encrypted with Vault Transit (master key managed by Vault)
- Encrypted DEKs are stored in the database alongside volume metadata
- Volume data is encrypted/decrypted on-the-fly using the DEK

### Key Components

#### 1. KMS Client Interface (`internal/core/ports/kms_client.go`)
```go
type KMSClient interface {
    Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error)
    Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error)
    GenerateKey(ctx context.Context, keyID string) ([]byte, error)
}
```

#### 2. Vault Transit Adapter (`internal/adapters/vault/transit_kms_adapter.go`)
- Implements `KMSClient` interface
- `POST /transit/encrypt/{key_name}` for encryption
- `POST /transit/decrypt/{key_name}` for decryption
- Handles Vault token authentication and base64 encoding

#### 3. Volume Encryption Repository (`internal/repositories/postgres/volume_encryption_repo.go`)
Stores encrypted DEKs in PostgreSQL:
```sql
CREATE TABLE volume_encryption_keys (
    volume_id      UUID PRIMARY KEY REFERENCES volumes(id) ON DELETE CASCADE,
    encrypted_dek  BYTEA NOT NULL,
    kms_key_id     VARCHAR(500) NOT NULL,
    algorithm      VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### 4. Volume Encryption Service (`internal/core/services/volume_encryption.go`)
Manages DEK lifecycle:
- `CreateVolumeKey`: Generate 256-bit DEK → Vault Encrypt → store encrypted DEK
- `GetVolumeDEK`: Fetch encrypted DEK from repo → Vault Decrypt → return plaintext DEK
- `DeleteVolumeKey`: Remove from repository on volume deletion
- `IsVolumeEncrypted`: Check if volume has encryption enabled

#### 5. Database Service Integration (`internal/core/services/database.go`)
- Added `volumeEncryption` dependency to `DatabaseService`
- When `CreateDatabase` includes `KmsKeyID`, sets `EncryptedVolume = true`
- `provisionDatabase` calls `CreateVolumeKey` for encrypted volumes
- `DeleteDatabase` calls `DeleteVolumeKey` for encrypted volumes

### API Changes

#### CreateDatabase Request
```json
{
  "name": "my-db",
  "engine": "postgres",
  "version": "15",
  "kms_key_id": "vault:transit/my-key"  // Optional: enables encryption
}
```

#### Database Response
```json
{
  "id": "...",
  "name": "my-db",
  "encrypted_volume": true,
  "kms_key_id": "vault:transit/my-key"
}
```

### DEK Lifecycle

1. **CreateDatabase + KmsKeyID** → validates → `provisionDatabase`
2. **provisionDatabase** → creates volume → calls `CreateVolumeKey(volumeID, kmsKeyID)`
3. **CreateVolumeKey** → generates 256-bit DEK → `vaultTransit.Encrypt(kmsKeyID, plaintextDEK)` → stores encrypted DEK
4. **Volume mount** → DEK passed as option to container launch
5. **Database startup** → DEK retrieved via `GetVolumeDEK` → used for transparent encryption wrapper
6. **DeleteDatabase** → `DeleteVolumeKey` cleans up DEK from repo

### Vault Transit Usage
- Key path format: `transit/encrypt/decrypt/{key_name}` where key_name is derived from `KmsKeyID`
- Vault Transit handles all cryptographic operations; application never sees master key
- `GenerateKey` uses the `transit/datakey/ciphertext/{key_name}` endpoint to generate a wrapped DEK

## Consequences

### Positive
- **Compliance**: Meets encryption at rest requirements for SOC 2, GDPR, HIPAA
- **Security**: Volume data is encrypted; lost/stolen volumes cannot be read
- **Vendor Abstraction**: KMS interface allows swapping Vault for AWS KMS, GCP KMS, etc.
- **No Engine Changes**: Works with PostgreSQL, MySQL without modification

### Negative
- **Performance**: Encryption/decryption adds latency on every I/O operation
- **Complexity**: Additional service and repository for key management
- **Dependencies**: Requires Vault Transit to be available; graceful degradation needed
- **Key Management**: DEKs stored in database (encrypted by Vault); losing Vault access locks data

### Neutral
- **Application-Level**: Encryption happens at application layer, not block device layer
- **Transparent to Engine**: Database engine sees plaintext; encryption is handled by wrapper
- **Testable**: Mocks allow unit testing without Vault

## Alternatives Considered

### 1. Database Engine Native Encryption (TDE)
- PostgreSQL `pg_encryption`, MySQL `InnoDB TDE`
- **Rejected**: Requires engine-level configuration, not portable across engines

### 2. Block Device Encryption (dm-crypt/LUKS)
- OS-level full disk encryption
- **Rejected**: Requires privileged access, not portable across backends

### 3. Cloud Provider KMS (AWS KMS, GCP CKMS)
- Native cloud provider encryption
- **Rejected**: Couples platform to specific cloud; reduces portability

### 4. Application-Level Encryption with Manual Key Storage
- Encrypt data in application, store key alongside ciphertext
- **Rejected**: Key management becomes ad-hoc; no audit trail for key operations

## Implementation Notes

- Import cycle avoided by having `VolumeEncryptionServiceImpl` take the `ports.VolumeEncryptionRepository` interface (postgres doesn't import ports)
- Encryption wrapper (AES-256-GCM streaming) uses existing chunked framing from ADR-020
- Integration tests require Docker/testcontainers; skipped in CI without Docker
- Unit tests use mocks for `KMSClient` and `VolumeEncryptionRepository`
- Key ID format supports both `vault:transit/key-name` (with scheme prefix) and `key-name` (raw)

## References

- [Vault Transit Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/transit)
- [ADR-020: Storage Streaming and Authenticated Encryption](./ADR-020-storage-streaming-and-authenticated-encryption.md)
- [Database Guide](../database.md)
