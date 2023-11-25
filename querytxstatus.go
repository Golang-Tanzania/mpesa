package mpesa

import "context"

func (c *Client) QueryTxStatus(ctx context.Context, payload QueryTxStatusRequest) (*QueryTxStatusResponse, error) {

	req, err := c.NewRequest(ctx, "GET", c.makeUrl(QueryTxStatusPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &QueryTxStatusResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
