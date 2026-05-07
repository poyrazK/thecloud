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
	var res Response[[]*domain.Snapshot]
	err := c.get("/snapshots", &res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) GetSnapshot(idOrName string) (*domain.Snapshot, error) {
	id, err := c.resolveID("snapshot", func() ([]interface{}, error) {
		snaps, err := c.ListSnapshots()
		return interfaceSlice(snaps), err
	}, func(v interface{}) string { return v.(*domain.Snapshot).ID.String() }, func(v interface{}) string { return v.(*domain.Snapshot).VolumeName }, idOrName)
	if err != nil {
		return nil, err
	}
	var snapshot domain.Snapshot
	err = c.get(fmt.Sprintf("/snapshots/%s", id), &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (c *Client) DeleteSnapshot(idOrName string) error {
	id, err := c.resolveID("snapshot", func() ([]interface{}, error) {
		snaps, err := c.ListSnapshots()
		return interfaceSlice(snaps), err
	}, func(v interface{}) string { return v.(*domain.Snapshot).ID.String() }, func(v interface{}) string { return v.(*domain.Snapshot).VolumeName }, idOrName)
	if err != nil {
		return err
	}
	return c.delete(fmt.Sprintf("/snapshots/%s", id), nil)
}

func (c *Client) RestoreSnapshot(idOrName string, newVolumeName string) (*domain.Volume, error) {
	id, err := c.resolveID("snapshot", func() ([]interface{}, error) {
		snaps, err := c.ListSnapshots()
		return interfaceSlice(snaps), err
	}, func(v interface{}) string { return v.(*domain.Snapshot).ID.String() }, func(v interface{}) string { return v.(*domain.Snapshot).VolumeName }, idOrName)
	if err != nil {
		return nil, err
	}
	req := map[string]interface{}{
		"new_volume_name": newVolumeName,
	}

	var vol domain.Volume
	err = c.post(fmt.Sprintf("/snapshots/%s/restore", id), req, &vol)
	if err != nil {
		return nil, err
	}
	return &vol, nil
}
