package nextdns

import (
	"context"
	"net/url"
	"strconv"
)

type LogListResponse = Response[[]Log]

type LogsQuery struct {
	From   string
	To     string
	Sort   string
	Limit  int
	Cursor string
	Device string
	Status string
	Search string
	Raw    bool
}

func (q LogsQuery) Values() url.Values {
	v := url.Values{}
	if q.From != "" {
		v.Set("from", q.From)
	}
	if q.To != "" {
		v.Set("to", q.To)
	}
	if q.Sort != "" {
		v.Set("sort", q.Sort)
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Cursor != "" {
		v.Set("cursor", q.Cursor)
	}
	if q.Device != "" {
		v.Set("device", q.Device)
	}
	if q.Status != "" {
		v.Set("status", q.Status)
	}
	if q.Search != "" {
		v.Set("search", q.Search)
	}
	if q.Raw {
		v.Set("raw", "1")
	}
	return v
}

func (c *Client) GetLogs(ctx context.Context, profileID string, q LogsQuery) (*LogListResponse, error) {
	query := q.Values()
	var resp LogListResponse
	err := c.do(ctx, "GET", "/profiles/"+profileID+"/logs", query, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
