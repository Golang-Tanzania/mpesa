package mpesa

import (
	"context"
)

func (c *Client) B2CPayment(ctx context.Context, payload B2CPaymentRequest) (*B2CPaymentResponse, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(B2CPaymentPath), payload)
	if err != nil {
		return &B2CPaymentResponse{}, err
	}
	resp := &B2CPaymentResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}


