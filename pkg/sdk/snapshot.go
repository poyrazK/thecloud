// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (c *Client) CreateSnapshot(volumeID uuid.UUID, description string) (*domain.Snapshot, error) {
	req := map[string]interface{}{
		"volume_id":   volumeID,
		"description": description,
	}

	var snapshot domain.Snapshot
	err := c.post("/snapshots", req, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (c *Client) ListSnapshots() ([]*domain.Snapshot, error) {
	var snapshots []*domain.Snapshot
	err := c.get("/snapshots", &snapshots)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (c *Client) GetSnapshot(id string) (*domain.Snapshot, error) {
	var snapshot domain.Snapshot
	err := c.get(fmt.Sprintf("/snapshots/%s", id), &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (c *Client) DeleteSnapshot(id string) error {
	return c.delete(fmt.Sprintf("/snapshots/%s", id), nil)
}

func (c *Client) RestoreSnapshot(id string, newVolumeName string) (*domain.Volume, error) {
	req := map[string]interface{}{
		"new_volume_name": newVolumeName,
	}

	var vol domain.Volume
	err := c.post(fmt.Sprintf("/snapshots/%s/restore", id), req, &vol)
	if err != nil {
		return nil, err
	}
	return &vol, nil
}
