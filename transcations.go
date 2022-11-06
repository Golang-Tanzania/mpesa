package gopesa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
