package mpesa

import "context"

func (c *Client) Reversal(ctx context.Context, payload ReversalRequest) (*ReversalResponse, error) {

	req, err := c.NewRequest(ctx, "PUT", c.makeUrl(ReversalPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &ReversalResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
