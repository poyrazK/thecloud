package sdk

import (
	"strconv"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ListAuditLogs retrieves audit logs for the current user.
func (c *Client) ListAuditLogs(limit int) ([]domain.AuditLog, error) {
	path := "/audit"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}

	var res Response[[]domain.AuditLog]
	if err := c.get(path, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
