package mpesa

import (
	"context"
)

func (c *Client) B2BPayment(ctx context.Context, payload B2BPaymentRequest) (*B2BPaymentResponse, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(B2BPaymentPath), payload)
	if err != nil {
		return &B2BPaymentResponse{}, err
	}
	resp := &B2BPaymentResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}


