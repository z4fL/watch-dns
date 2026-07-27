package nextdns

import "context"

type AddDenylistRequest struct {
	ID string `json:"id"`
}

type AddDenylistResponse = Response[struct {
	ID string `json:"id"`
}]

func (c *Client) AddToDenylist(ctx context.Context, profileID, domain string) error {
	req := AddDenylistRequest{ID: domain}
	var resp AddDenylistResponse
	return c.do(ctx, "POST", "/profiles/"+profileID+"/denylist", nil, req, &resp)
}
