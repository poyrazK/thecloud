// Package domain defines core business entities.
package domain

// VolumeAttachment represents a request to attach a volume to an instance.
// Used during instance creation or update operations.
type VolumeAttachment struct {
	VolumeIDOrName string `json:"volume_id"`  // UUID or Name of volume to attach
	MountPath      string `json:"mount_path"` // Target path inside container/VM (e.g. "/mnt/data")
}
