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
	"io/ioutil"
	"net/http"
)

// Type KEYS that will be assigned from config.json

type KEYS struct {
	PUBLICKEY string
	APIKEY    string
}

// Initialization method

func (api *APICONTEXT) Initialize(file string) *APICONTEXT {

	keys, err := ioutil.ReadFile(file)
	mustNot("Error reading file: ", err)

	var Keys KEYS
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

// Create new bearer token

func (api *APICONTEXT) createBearerToken(apiKey string) string {
	keyDer, _ := pem.Decode([]byte(api.PUBLICKEY))
	pub, _ := x509.ParsePKIXPublicKey([]byte(keyDer.Bytes))
	pubKey := pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(apiKey))
	mustNot("Error during encryption: ", err)
	encryptedKey := base64.StdEncoding.EncodeToString([]byte(cipherText))

	return encryptedKey
}

// Generate a new session ID

func (api *APICONTEXT) generateSessionID() string {
	api.setDefault()
	endpoint := "getSession"
	req, err := http.NewRequest("GET", api.getURL(endpoint), nil)
	mustNot("Error making new request: ", err)

	for k, v := range api.getHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{}

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
