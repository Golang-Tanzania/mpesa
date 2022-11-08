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
	"log"
	"net/http"
)

// A helper function that will check errors
func mustNot(message string, err error) {
	if err != nil {
		log.Println(message, err)
	}
}

// Type APICONTEXT that stores the API's configurable info
type APICONTEXT struct {
	PUBLICKEY   string
	APIKEY      string
	ENVIRONMENT string
	ssl         bool
	address     string
	port        int
	headers     map[string]string
	parameters  map[string]string
}

// setDefualt to set default values on the APICONTEXT type
// It will also add the default headers
func (api *APICONTEXT) setDefault() {

	api.address = "openapi.m-pesa.com"
	api.ssl = true
	api.port = 443

	// Defualt Headers

	api.headers = make(map[string]string)
	api.parameters = make(map[string]string)

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	api.addHeader("Host", api.address)

	api.addHeader("Origin", "*")
	api.addHeader("Content-Type", "application/json")

}

// getURL will get the absolute URL
// It will return http or https depending on the value of ssl
// Default URL is https://
func (api *APICONTEXT) getURL(endpoint string) string {
	url := ""
	if api.ssl {
		url = fmt.Sprintf("https://%v:%v%v", api.address, api.port, api.getPath(endpoint))
	} else {
		url = fmt.Sprintf("http://%v:%v%v", api.address, api.port, api.getPath(endpoint))
	}

	return url
}

// addHeader will add key/value header pairs to APICONTEXT.headers
func (api *APICONTEXT) addHeader(key, value string) {
	api.headers[key] = value
}

// Get all headers assigned to APICONTEXT.headers
func (api *APICONTEXT) getHeaders() map[string]string {
	return api.headers
}

// addParameter will add key/value parameters to APICONTEXT.parameters
// Paramters are the transaction queries
func (api *APICONTEXT) addParameter(key, value string) {
	api.parameters[key] = value
}

// Get all paramters assigned to APICONTEXT.parameters
func (api *APICONTEXT) getParameters() map[string]string {
	return api.parameters
}

// getPath will determine the endpoints to be used depending on the kind of transaction.
func (api *APICONTEXT) getPath(url string) string {
	if api.ENVIRONMENT == "production" {
		return fmt.Sprintf("/openapi/ipg/v2/vodacomTZN/%v/", url)
	} else {
		return fmt.Sprintf("/sandbox/ipg/v2/vodacomTZN/%v/", url)
	}
}

func (api *APICONTEXT) sendRequest(transactionQuery map[string]string, method string, endpoint string) string {
	api.APIKEY = api.generateSessionID()
	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, err := json.Marshal(api.parameters)
	mustNot("Error parsing transaction queries: ", err)

	req, err := http.NewRequest(method, api.getURL(endpoint), bytes.NewBuffer(jsonParameters))
	mustNot("Error creating New request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	mustNot("Error getting response", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	mustNot("Error reading response body: ", err)

	return string(body)

}
