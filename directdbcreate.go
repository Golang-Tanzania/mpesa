package mpesa

import (
	"context"
)

func (c *Client) DirectDebitCreate(ctx context.Context, payload DirectDebitRequest) (*DirectDebitResponse, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(DirectDebitPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &DirectDebitResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
