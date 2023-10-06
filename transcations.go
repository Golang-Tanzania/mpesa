/*
Copyright (c) 2022 Golang Tanzania

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

package gopesa

import "net/http"

// A method to conduct Customer to Business transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) C2BPayment(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "c2bPayment/singleStage")
}

// A method to conduct Business to Customer transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) B2CPayment(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "b2cPayment")
}

// A method to conduct Business to Business transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) B2BPayment(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "b2bPayment")
}

// A method to conduct Reverse Payments.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) ReversePayment(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPut, "reversal")
}

// A method to query a trasaction status.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) TransactionStatus(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodGet, "queryTransactionStatus")
}

func (api *APICONTEXT) QueryBeneficiaryName(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodGet, "queryBeneficiaryName")
}

func (api *APICONTEXT) QueryDirectDebit(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "queryDirectDebit")
}

func (api *APICONTEXT) DirectDebitCreate(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "directDebitCreation")
}

func (api *APICONTEXT) DirectDebitPayment(transactionQuery map[string]string) string {
	return api.sendRequest(transactionQuery, http.MethodPost, "directDebitPayment")
}
