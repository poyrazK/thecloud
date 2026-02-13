// Package domain defines core business entities.
package domain

import (
	"github.com/google/uuid"
)

// Permission represents a specific authorization grant string.
// Format: "resource:action" (e.g. "instance:launch")
type Permission string

// Permissions define coarse-grained access control for resources.
const (
	// Compute Permissions
	PermissionInstanceLaunch    Permission = "instance:launch"
	PermissionInstanceTerminate Permission = "instance:terminate"
	PermissionInstanceRead      Permission = "instance:read"
	PermissionInstanceUpdate    Permission = "instance:update"

	// SSH Key Permissions
	PermissionSSHKeyCreate Permission = "ssh_key:create"
	PermissionSSHKeyRead   Permission = "ssh_key:read"
	PermissionSSHKeyDelete Permission = "ssh_key:delete"

	// VPC Permissions
	PermissionVpcCreate Permission = "vpc:create"
	PermissionVpcDelete Permission = "vpc:delete"
	PermissionVpcRead   Permission = "vpc:read"
	PermissionVpcUpdate Permission = "vpc:update"

	// Elastic IP Permissions
	PermissionEipAllocate  Permission = "eip:allocate"
	PermissionEipRelease   Permission = "eip:release"
	PermissionEipRead      Permission = "eip:read"
	PermissionEipAssociate Permission = "eip:associate"

	// Storage Permissions
	PermissionVolumeCreate Permission = "volume:create"
	PermissionVolumeDelete Permission = "volume:delete"
	PermissionVolumeRead   Permission = "volume:read"

	// Snapshot Permissions
	PermissionSnapshotCreate  Permission = "snapshot:create"
	PermissionSnapshotDelete  Permission = "snapshot:delete"
	PermissionSnapshotRead    Permission = "snapshot:read"
	PermissionSnapshotRestore Permission = "snapshot:restore"

	// Load Balancer Permissions
	PermissionLbCreate Permission = "lb:create"
	PermissionLbDelete Permission = "lb:delete"
	PermissionLbRead   Permission = "lb:read"
	PermissionLbUpdate Permission = "lb:update"

	// Database Permissions
	PermissionDBCreate Permission = "db:create"
	PermissionDBDelete Permission = "db:delete"
	PermissionDBRead   Permission = "db:read"

	// Secret Permissions
	PermissionSecretCreate Permission = "secret:create"
	PermissionSecretDelete Permission = "secret:delete"
	PermissionSecretRead   Permission = "secret:read"

	// Function Permissions
	PermissionFunctionInvoke Permission = "function:invoke"
	PermissionFunctionCreate Permission = "function:create"
	PermissionFunctionDelete Permission = "function:delete"
	PermissionFunctionRead   Permission = "function:read"

	// Cache Permissions
	PermissionCacheCreate Permission = "cache:create"
	PermissionCacheDelete Permission = "cache:delete"
	PermissionCacheRead   Permission = "cache:read"
	PermissionCacheUpdate Permission = "cache:update"

	// Queue Permissions
	PermissionQueueCreate Permission = "queue:create"
	PermissionQueueDelete Permission = "queue:delete"
	PermissionQueueRead   Permission = "queue:read"
	PermissionQueueWrite  Permission = "queue:write"

	// Notify Permissions
	PermissionNotifyCreate Permission = "notify:create"
	PermissionNotifyDelete Permission = "notify:delete"
	PermissionNotifyRead   Permission = "notify:read"
	PermissionNotifyWrite  Permission = "notify:publish"

	// Cron Permissions
	PermissionCronCreate Permission = "cron:create"
	PermissionCronDelete Permission = "cron:delete"
	PermissionCronRead   Permission = "cron:read"
	PermissionCronUpdate Permission = "cron:update"

	// Gateway Permissions
	PermissionGatewayCreate Permission = "gateway:create"
	PermissionGatewayDelete Permission = "gateway:delete"
	PermissionGatewayRead   Permission = "gateway:read"

	// IaC Permissions
	PermissionStackCreate Permission = "stack:create"
	PermissionStackDelete Permission = "stack:delete"
	PermissionStackRead   Permission = "stack:read"

	// Auto-Scaling Permissions
	PermissionAsCreate Permission = "as:create"
	PermissionAsDelete Permission = "as:delete"
	PermissionAsRead   Permission = "as:read"
	PermissionAsUpdate Permission = "as:update"

	// Container Permissions
	PermissionContainerCreate Permission = "container:create"
	PermissionContainerDelete Permission = "container:delete"
	PermissionContainerRead   Permission = "container:read"
	PermissionContainerUpdate Permission = "container:scale"

	// Image Permissions
	PermissionImageCreate Permission = "image:create"
	PermissionImageRead   Permission = "image:read"
	PermissionImageDelete Permission = "image:delete"

	// Cluster Permissions
	PermissionClusterCreate Permission = "cluster:create"
	PermissionClusterDelete Permission = "cluster:delete"
	PermissionClusterRead   Permission = "cluster:read"
	PermissionClusterUpdate Permission = "cluster:update"

	// System Permissions
	PermissionFullAccess Permission = "*"
)

// Role represents a named collection of permissions.
type Role struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// Default Roles
const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
)
