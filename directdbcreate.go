package mpesa

import (
	"context"
)

func (c *Client) DirectDebitCreate(ctx context.Context, payload DirectDBCreateReq) (*DirectDBCreateRes, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(DirectDebitPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &DirectDBCreateRes{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
