package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (api *APICONTEXT) C2BPayments(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()
	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, _ := json.Marshal(api.parameters)

	endpoint := "c2bPayment/singleStage"

	req, _ := http.NewRequest("POST", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
	}

	return string(body)

}

func (api *APICONTEXT) B2BPayments(transactionQuery map[string]string) string {
	api.APIKEY = api.generateSessionID()

	for k, v := range transactionQuery {
		api.addParameter(k, v)
	}

	bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
	api.addHeader("Authorization", bearer)

	jsonParameters, _ := json.Marshal(api.parameters)

	endpoint := "b2bPayment"

	req, _ := http.NewRequest("POST", api.getURL(endpoint), bytes.NewBuffer(jsonParameters))

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return string(body)
}