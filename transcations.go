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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// A method to conduct Customer to Business transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) C2BPayment(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()
	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing C2B transaction queries: ", err)

	endpoint := "c2bPayment/singleStage"

	req, err := http.NewRequest("POST", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New C2B request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting C2B response", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading C2B response body: ", err)

	return string(body)

}

// A method to conduct Business to Customer transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) B2CPayment(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()
	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing B2C transaction queries: ", err)

	endpoint := "b2cPayment"

	req, err := http.NewRequest("POST", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New B2C request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting B2C response", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading B2C response body: ", err)

	return string(body)

}

// A method to conduct Business to Business transactions.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) B2BPayment(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()

	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing B2B transaction queries: ", err)

	endpoint := "b2bPayment"

	req, err := http.NewRequest("POST", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New B2B request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting B2B response: ", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading B2B response body: ", err)

	return string(body)
}

// A method to conduct Reverse Payments.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) ReversePayment(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()

	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing reversal transaction queries: ", err)

	endpoint := "reversal"

	req, err := http.NewRequest("PUT", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New payment reversal request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting payment reversal response: ", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading payment reversal response body: ", err)

	return string(body)
}

// A method to query a trasaction status.
// It accepts transaction queries as a parameter.
// It returns the http response as a string.
func (api *APICONTEXT) TransactionStatus(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()

	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing transactionstatus queries: ", err)

	endpoint := "queryTransactionStatus"

	req, err := http.NewRequest("GET", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New transaction status request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting transaction status response: ", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading transaction status response body: ", err)

	return string(body)
}

func (api *APICONTEXT) QueryBeneficiaryName(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()

	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing query Beneficiary Name queries: ", err)

	endpoint := "queryBeneficiaryName"

	req, err := http.NewRequest("GET", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New query Beneficiary Name request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting query Beneficiary Name response: ", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading query Beneficiary Name response body: ", err)

	return string(body)
}





