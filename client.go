package mpesa

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

	"github.com/spf13/viper"
)

func (c *Client) LoadKeys(path, filename, filetype, env string) {

	viper.AddConfigPath(path)
	viper.SetConfigName(filename)
	viper.SetConfigType(filetype)
	viper.ReadInConfig()

	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file, " + err.Error())
	}

	c.Keys = &Keys{
		PublicKey: viper.GetString("public_key"),
		ApiKey:    viper.GetString("api_key"),
	}

	c.Environment = env
}

func (c *Client) SetHttpClient(client *http.Client) {
	if client == nil {
		c.Client = http.DefaultClient
		return
	}
	c.Client = client
}

func (c *Client) createBearerToken() (string, error) {

	keyDer, _ := pem.Decode([]byte(c.fmtPubKey(c.Keys.PublicKey)))
	pub, err := x509.ParsePKIXPublicKey([]byte(keyDer.Bytes))
	if err != nil {
		return "", err
	}
	pubKey := pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(c.Keys.ApiKey))
	if err != nil {
		return "", err
	}
	encryptedKey := base64.StdEncoding.EncodeToString([]byte(cipherText))

	return encryptedKey, nil
}

func (c *Client) makeUrl(endpoint string) string {

	var url string

	if c.Environment == "production" {
		url = fmt.Sprintf("https://%v:%v%v%v/", Address, Port, ProdEndpoint, endpoint)
	} else {
		url = fmt.Sprintf("https://%v:%v%v%v/", Address, Port, SandboxEndpoint, endpoint)
	}

	return url
}

func (c *Client) genSessionId() (*SessionIDResponse, error) {

	req, err := http.NewRequest("GET", c.makeUrl(SessionEndPath), nil)

	if err != nil {
		return nil, err
	}

	bearerToken, err := c.createBearerToken()

	if err != nil {
		return nil, err
	}
	bearer := fmt.Sprintf("Bearer %v", bearerToken)

	req.Header.Set("Authorization", bearer)
	req.Header.Set("Host", Address)
	req.Header.Set("Origin", "*")
	req.Header.Set("Content-Type", "application/json")

	client := c.Client

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var res SessionIDResponse
	err = json.Unmarshal([]byte(body), &res)

	if err != nil {
		return nil, err
	}

	return &res, nil

}

func (c *Client) fmtPubKey(publicKey string) string {
	pubKey := fmt.Sprintf(`
-----BEGIN RSA PUBLIC KEY-----
%s
-----END RSA PUBLIC KEY-----`, publicKey)
	return pubKey
}


