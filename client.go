package mpesa

import (
	"bytes"
	"context"
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

func (c *Client) createBearerToken(apiKey string) (string, error) {

	keyDer, _ := pem.Decode([]byte(c.fmtPubKey(c.Keys.PublicKey)))
	pub, err := x509.ParsePKIXPublicKey([]byte(keyDer.Bytes))
	if err != nil {
		return "", err
	}
	pubKey := pub.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(apiKey))
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

func (c *Client) genSessionKey() (*SessionKeyResponse, error) {

	req, err := http.NewRequest("GET", c.makeUrl(SessionEndPath), nil)

	if err != nil {
		return nil, err
	}

	var res *SessionKeyResponse

	err = c.SendWithAuth(req, &res, nil)

	if err != nil {
		return nil, err
	}

	return res, nil

}

func (c *Client) fmtPubKey(publicKey string) string {
	pubKey := fmt.Sprintf(`
-----BEGIN RSA PUBLIC KEY-----
%s
-----END RSA PUBLIC KEY-----`, publicKey)
	return pubKey
}

func (c *Client) Send(req *http.Request, v interface{}, e interface{}) error {
	var (
		err  error
		resp *http.Response
		data []byte
	)

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Host", Address)
	req.Header.Set("Origin", "*")

	if req.Header.Get("Content-type") == "" {
		req.Header.Set("Content-type", "application/json")
	}

	resp, err = c.Client.Do(req)

	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) error {
		return Body.Close()
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {

		data, err = io.ReadAll(resp.Body)

		if e == nil {
			return fmt.Errorf("unknown error (%s), status code: %d", string(data), resp.StatusCode)
		}

		if err == nil && len(data) > 0 {
			err := json.Unmarshal(data, e)
			if err != nil {
				return err
			}
		}

		return fmt.Errorf("unknown error, status code: %d", resp.StatusCode)
	}
	if v == nil {
		return nil
	}

	if w, ok := v.(io.Writer); ok {
		_, err := io.Copy(w, resp.Body)
		return err
	}
	return json.NewDecoder(resp.Body).Decode(v)

}

func (c *Client) SendWithAuth(req *http.Request, v interface{}, e interface{}) error {

	bearerToken, err := c.createBearerToken(c.Keys.ApiKey)

	if err != nil {
		return err
	}
	bearer := fmt.Sprintf("Bearer %v", bearerToken)

	req.Header.Set("Authorization", bearer)

	return c.Send(req, v, e)
}

func (c *Client) SendWithSessionKey(req *http.Request, v interface{}, e interface{}) error {

	sessionkey, err := c.genSessionKey()

	if err != nil {
		return err
	}

	tokenKey, err := c.createBearerToken(sessionkey.OutputSessionID)

	if err != nil {
		return err
	}

	bearer := fmt.Sprintf("Bearer %v", tokenKey)

	req.Header.Set("Authorization", bearer)

	return c.Send(req, v, e)
}

func (c *Client) NewRequest(ctx context.Context, method, url string, payload interface{}) (*http.Request, error) {
	var buf io.Reader

	if payload != nil {
		b, err := json.Marshal(&payload)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	}
	return http.NewRequestWithContext(ctx, method, url, buf)
}
