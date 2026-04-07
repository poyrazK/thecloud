# ADR 023: Tri-State RBAC Evaluation Model

## Status
Accepted

## Context
Our RBAC system initially used a binary "Allow/Deny" logic. As we introduced IAM Policies, we needed a way to support explicit `Deny` statements that override any `Allow` statements (from other policies or from the user's base role). A simple boolean return from the policy evaluator was insufficient to distinguish between "No policy matched" and "Policy explicitly denied".

## Decision
We decided to implement a **Tri-State Evaluation Model** for IAM Policies and integrate it into the RBAC authorization flow.

### 1. Evaluation States
The `PolicyEvaluator.Evaluate` method now returns a `domain.PolicyEffect` (string) which can be:
- `Allow`: An explicit statement granted access.
- `Deny`: An explicit statement denied access.
- `""` (Empty): No statements matched the requested action/resource.

### 2. Authorization Hierarchy
The `RBACService` now follows this precedence order when checking permissions:
1. **System Bypass**: If `IsInternalCall` is true, access is granted immediately.
2. **Explicit Deny**: If any IAM policy explicitly denies the action, access is **Forbidden**, regardless of roles.
3. **Explicit Allow**: If any IAM policy explicitly allows the action, access is **Granted**.
4. **Role-Based Fallback**: If no IAM policies match (`NoMatch`), the system falls back to the permissions defined in the user's database role (e.g., `admin`, `developer`, `viewer`).
   - The `admin` role has a hardcoded bypass for all permissions unless an explicit IAM Deny exists.

## Consequences
- **Security**: "Explicit Deny" provides a powerful way to restrict even administrative users for sensitive resources.
- **Predictability**: The precedence rules align with industry standards (e.g., AWS IAM).
- **Extensibility**: The tri-state model allows for future additions (like Organization Service Control Policies) to be inserted into the hierarchy easily.
- **Breaking Change**: The `PolicyEvaluator` interface changed from returning `bool` to `domain.PolicyEffect`, requiring updates to all implementations and mocks.
