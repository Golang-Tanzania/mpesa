package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
)

func (api *APICONTEXT) createBearerToken(apiKey string) string {
	keyDer, _ := pem.Decode([]byte(api.PUBLICKEY))
	pub, _ := x509.ParsePKIXPublicKey([]byte(keyDer.Bytes))
	pubKey := pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(apiKey))
	if err != nil {
		fmt.Println("Error:", err.Error())
	}
	encryptedKey := base64.StdEncoding.EncodeToString([]byte(cipherText))

	return encryptedKey
}

func (api *APICONTEXT) generateSessionID() string {
	api.setDefault()
	endpoint := "getSession"
	req, _ := http.NewRequest("GET", api.getURL(endpoint), nil)

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

	var result map[string]string
	json.Unmarshal([]byte(body), &result)

	return result["output_SessionID"]
}
