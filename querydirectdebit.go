package mpesa

import (
	"context"
)

func (c *Client) QueryDirectDebit(ctx context.Context, payload QueryDirectDBReq) (*QueryDirectDBRes, error) {

	req, err := c.NewReqWithQueryParams(ctx, "GET", c.makeUrl(QueryDirectDBPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &QueryDirectDBRes{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
