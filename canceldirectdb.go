package mpesa

import (
	"context"
)

func (c *Client) CancelDirectDebit(ctx context.Context, payload CancelDirectDBReq) (*CancelDirectDBRes, error) {

	req, err := c.NewRequest(ctx, "PUT", c.makeUrl(CancelDirectDBPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &CancelDirectDBRes{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
