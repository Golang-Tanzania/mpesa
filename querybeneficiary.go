package mpesa

import (
	"context"
)

func (c *Client) QueryBeneficiaryName(ctx context.Context, payload QueryBenRequest) (*QueryBenResponse, error) {

	req, err := c.NewReqWithQueryParams(ctx, "GET", c.makeUrl(QueryBeneficialPath), payload)
	if err != nil {
		return nil, err
	}
	resp := &QueryBenResponse{}

	err = c.SendWithSessionKey(req, resp, nil)

	if err != nil {
		return nil, err
	}

	return resp, err
}
