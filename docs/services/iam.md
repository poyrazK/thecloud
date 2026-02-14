# Cloud IAM (Identity & Access Management)

Cloud IAM provides granular, document-based access control for all resources in **The Cloud**. It supplements the legacy RBAC system with fine-grained policies (ABAC), allowing for precise permission management.

## Core Concepts

### 1. Policies
A policy is a JSON document that defines a set of permissions. It consists of one or more **Statements**.

### 2. Statements
Each statement contains:
- **Effect**: Either `Allow` or `Deny`.
- **Action**: A list of operations (e.g., `instance:launch`, `vpc:read`, `*`).
- **Resource**: A list of resource identifiers the actions apply to (e.g., `vpc:123`, `instance:*`, `*`).
- **Condition**: (Optional) Logic to further restrict when the policy applies.

### 3. Evaluation Logic
The IAM evaluator follows these rules in order:
1. **Explicit Deny**: If any applicable policy has a `Deny` effect for the action/resource, the request is denied. **Deny always wins.**
2. **Explicit Allow**: If any applicable policy has an `Allow` effect and no `Deny` exists, the request is allowed.
3. **Implicit Deny**: If no policy explicitly allows the action, the request is denied (unless falling back to legacy roles).

## Policy Examples

### Full Admin Access
```json
{
  "name": "AdministratorAccess",
  "statements": [
    {
      "effect": "Allow",
      "action": ["*"],
      "resource": ["*"]
    }
  ]
}
```

### Read-Only Access to Instances
```json
{
  "name": "InstanceReadOnly",
  "statements": [
    {
      "effect": "Allow",
      "action": ["instance:read", "instance:list"],
      "resource": ["*"]
    }
  ]
}
```

### Deny Deletion of a Specific VPC
```json
{
  "name": "ProtectProductionVPC",
  "statements": [
    {
      "effect": "Deny",
      "action": ["vpc:delete"],
      "resource": ["vpc:prod-vpc-uuid"]
    }
  ]
}
```

## Wildcard Support
IAM Policies support wildcards in both actions and resources:
- `*`: Matches everything.
- `instance:*`: Matches all actions starting with `instance:`.
- `vpc:prod-*`: Matches all resources starting with `vpc:prod-`.

## Integration with RBAC
The system is designed for backward compatibility:
- If a user has **IAM Policies attached**, the system evaluates them first.
- If evaluation results in `Allow`, access is granted.
- If evaluation results in an **Explicit Deny**, access is blocked (even if the user's role allows it).
- If no IAM policies apply, the system falls back to the user's assigned **Legacy Role** (`admin`, `developer`, `viewer`).

## API Usage

### Create a Policy
`POST /iam/policies`
```json
{
  "name": "MyPolicy",
  "statements": [...]
}
```

### Attach to User
`POST /iam/users/{userId}/policies/{policyId}`

### List User Policies
`GET /iam/users/{userId}/policies`
