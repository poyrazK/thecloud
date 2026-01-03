# Secrets Manager

Secrets Manager allows you to store and manage sensitive information such as API keys, database passwords, and other confidential configuration values. All secrets are encrypted at rest using AES-256-GCM.

## Overview

- **Encryption**: Secrets are encrypted using a master key (`SECRETS_ENCRYPTION_KEY`) and derived per-user using HKDF.
- **Isolation**: Secrets are strictly scoped to the user who created them.
- **Audit Logging**: All access to secrets (creation, retrieval, deletion) is logged in the system events.
- **Auto-Redaction**: Secret values are redacted automatically when listing to prevent accidental exposure.

## CLI Usage

### 1. Create a Secret

```bash
cloud secrets create --name STRIPE_API_KEY --value "sk_test_..." --description "Production Stripe Key"
```

### 2. List Secrets

Listing retrieves metadata only. Values are redacted for security.

```bash
cloud secrets list
```

### 3. Retrieve a Secret

Retrieving a secret by name or ID will decrypt the value.

```bash
cloud secrets get STRIPE_API_KEY
```

### 4. Delete a Secret

```bash
cloud secrets rm STRIPE_API_KEY
```

## Security Best Practices

1. **Master Key**: Ensure `SECRETS_ENCRYPTION_KEY` is set to a cryptographically strong 32-byte string in production.
2. **Access Control**: Use dedicated API keys for services that only need read access to specific secrets.
3. **Audit**: Regularly review `cloud events` to monitor secret access patterns.

## Future Roadmap

- [ ] **Instance Injection**: Automatically inject secrets into compute instances as environment variables.
- [ ] **Rotation**: Automatic rotation of secrets with versioning support.
- [ ] **Policies**: Fine-grained access policies for sharing secrets between users/teams.
