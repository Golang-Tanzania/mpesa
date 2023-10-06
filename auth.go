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

// Golang bindings for the
// Mpesa Payment API (see https://openapiportal.m-pesa.com),
// to easily make your MPESA payments ready to GO.
package gopesa

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
	"os"
	"time"
)

// Type kEYS that will receive values from the config.json
type kEYS struct {
	PUBLICKEY string
	APIKEY    string
}

// Initialization method that will read keys from
// config.json and assign them to PUBLICKEY AND APIKEY
// It will return a pointer of APICONTEXT with the
// new keys
func (api *APICONTEXT) Initialize(file string) *APICONTEXT {

	keys, err := os.ReadFile(file)
	mustNot("Error reading file: ", err)

	var Keys kEYS
	err = json.Unmarshal(keys, &Keys)
	mustNot("Error during Unmarshal: ", err)

	publicKey := fmt.Sprintf(`
-----BEGIN RSA PUBLIC KEY-----
%s
-----END RSA PUBLIC KEY-----`, Keys.PUBLICKEY)
	apiKey := Keys.APIKEY

	api.APIKEY = apiKey
	api.PUBLICKEY = publicKey

	return api
}

// A method that will create a new bearer token
// It will return a string with the value of the
// bearer token
func (api *APICONTEXT) createBearerToken(apiKey string) string {
	keyDer, _ := pem.Decode([]byte(api.PUBLICKEY))
	pub, _ := x509.ParsePKIXPublicKey([]byte(keyDer.Bytes))
	pubKey := pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(apiKey))
	mustNot("Error during encryption: ", err)
	encryptedKey := base64.StdEncoding.EncodeToString([]byte(cipherText))

	return encryptedKey
}

//	A method that will generate a new session ID
//
// It will return the session ID as a string
func (api *APICONTEXT) generateSessionID() string {
	api.setDefault()
	endpoint := "getSession"
	req, err := http.NewRequest("GET", api.getURL(endpoint), nil)
	mustNot("Error making new request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)

	mustNot("Error requesting Session ID: ", err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	mustNot("Error reading response body: ", err)

	var result map[string]string
	json.Unmarshal([]byte(body), &result)

	if result["output_SessionID"] != "" {
		return result["output_SessionID"]
	} else {
		return string(body)
	}
}
