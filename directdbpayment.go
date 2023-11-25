package mpesa

import (
	"context"
)

func (c *Client) DirectDebitPayment(ctx context.Context, payload DebitDBPaymentReq) (*DebitDBPaymentRes, error) {

	req, err := c.NewRequest(ctx, "POST", c.makeUrl(DebitDBPaymentPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &DebitDBPaymentRes{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
