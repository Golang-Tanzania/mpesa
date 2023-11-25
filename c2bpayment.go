package mpesa

import (
	"context"
)

func (c *Client) C2BPayment(ctx context.Context, payload C2BPaymentRequest) (*C2BPaymentResponse, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(C2BPaymentPath), payload)
	if err != nil {
		return &C2BPaymentResponse{}, err
	}
	resp := &C2BPaymentResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
