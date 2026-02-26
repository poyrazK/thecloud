package sdk

import (
	"net/url"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// GetBillingSummary retrieves the billing summary for the current user.
func (c *Client) GetBillingSummary(start, end *time.Time) (*domain.BillSummary, error) {
	params := url.Values{}
	if start != nil {
		params.Add("start", start.Format(time.RFC3339))
	}
	if end != nil {
		params.Add("end", end.Format(time.RFC3339))
	}

	path := "/billing/summary"
	if params.Encode() != "" {
		path += "?" + params.Encode()
	}

	var res Response[domain.BillSummary]
	if err := c.get(path, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// ListUsageRecords retrieves detailed usage records for the current user.
func (c *Client) ListUsageRecords(start, end *time.Time) ([]domain.UsageRecord, error) {
	params := url.Values{}
	if start != nil {
		params.Add("start", start.Format(time.RFC3339))
	}
	if end != nil {
		params.Add("end", end.Format(time.RFC3339))
	}

	path := "/billing/usage"
	if params.Encode() != "" {
		path += "?" + params.Encode()
	}

	var res Response[[]domain.UsageRecord]
	if err := c.get(path, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
