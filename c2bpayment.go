/*
Copyright (c) 2022-2023 Golang Tanzania

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mpesa

import (
	"context"
)

// A method to perform Customer to Business transactions.
// It accepts C2BPaymentRequest request payload parameter.
// It returns the C2BPaymentResponse as a pointer struct.
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
