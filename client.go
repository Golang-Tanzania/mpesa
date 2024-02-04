/*
Copyright (c) 2022-2023 Golang Tanzania

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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"time"


)

// NewClient returns new Client struct
func NewClient(api_key string, envtype Env, sessionLife int32) (*Client, error) {
	if api_key == "" {
		return nil, errors.New("api Key is to create a Client")
	}

	var keys *Keys

	if envtype == Sandbox {
		keys = &Keys{
			PublicKey: SandboxPublicKey,
			ApiKey:    api_key,
		}

	} else if envtype == Production {
		keys = &Keys{
			PublicKey: OpenapiPublicKey,
			ApiKey:    api_key,
		}

	}

	return &Client{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Keys:        keys,
		Environment: envtype,
		SessionLife: sessionLife,
	}, nil
}

// SetHTTPClient sets *http.Client to current client
func (c *Client) SetHttpClient(client *http.Client) {
	c.Client = client
}

// createBearerToken encrypts the api key using the public key
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

// makeUrl constructs a full url
func (c *Client) makeUrl(endpoint string) string {

	var url string

	if c.Environment == Production {
		url = fmt.Sprintf("https://%v:%v%v%v/", Address, Port, ProdEndpoint, endpoint)
	} else {
		url = fmt.Sprintf("https://%v:%v%v%v/", Address, Port, SandboxEndpoint, endpoint)
	}

	return url
}

// genSessionKey generates a session key
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

// fmtPubKey formats the public key to the required format
func (c *Client) fmtPubKey(publicKey string) string {
	pubKey := fmt.Sprintf(`
-----BEGIN RSA PUBLIC KEY-----
%s
-----END RSA PUBLIC KEY-----`, publicKey)
	return pubKey
}

// Send makes a request to the API, the response body will be
// unmarshalled into v, or if v is an io.Writer, the response will
// be written to it without decoding
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

// SendWithAuth makes a request to the API and apply authentication automatically.
func (c *Client) SendWithAuth(req *http.Request, v interface{}, e interface{}) error {

	bearerToken, err := c.createBearerToken(c.Keys.ApiKey)

	if err != nil {
		return err
	}
	bearer := fmt.Sprintf("Bearer %v", bearerToken)

	req.Header.Set("Authorization", bearer)

	return c.Send(req, v, e)
}

// SendWithSessionKey makes a request to the API using generated sessionkey as bearer token.
func (c *Client) SendWithSessionKey(req *http.Request, v interface{}, e interface{}) error {

	c.mu.Lock()
	if c.SessionKey == "" {
		sessionkey, err := c.genSessionKey()
		if err != nil {
			return err
		}

		c.ExpiresAt = time.Now().Add(time.Duration(c.SessionLife) * time.Second * 3600)
		c.SessionKey = sessionkey.OutputSessionID

	} else if !c.ExpiresAt.IsZero() && time.Until(c.ExpiresAt) < ReqNewSessionKeyBeforeExpiresIn {
		sessionkey, err := c.genSessionKey()
		if err != nil {
			return err
		}

		c.ExpiresAt = time.Now().Add(time.Duration(c.SessionLife) * time.Second * 3600)
		c.SessionKey = sessionkey.OutputSessionID
	}

	c.mu.Unlock()

	tokenKey, err := c.createBearerToken(c.SessionKey)

	if err != nil {
		return err
	}

	bearer := fmt.Sprintf("Bearer %v", tokenKey)

	req.Header.Set("Authorization", bearer)

	return c.Send(req, v, e)
}

// NewRequest constructs a request
// Convert payload to a JSON
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

// QueryValuesFromStruct converts a struct to url.Values
func (c *Client) QueryValuesFromStruct(payload interface{}) (url.Values, error) {
	values := url.Values{}

	payloadValue := reflect.ValueOf(payload)

	if payloadValue.Kind() != reflect.Struct {
		return nil, errors.New("payload is not a struct")
	}

	for i := 0; i < payloadValue.NumField(); i++ {
		field := payloadValue.Type().Field(i)
		fieldValue := payloadValue.Field(i)

		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}

		values.Add(tag, fmt.Sprint(fieldValue.Interface()))
	}

	return values, nil
}

// NewReqWithQueryParams constructs a request with query params
func (c *Client) NewReqWithQueryParams(ctx context.Context, method, baseUrl string, payload interface{}) (*http.Request, error) {

	baseURL, err := url.Parse(baseUrl)

	if err != nil {
		return nil, err
	}

	params, err := c.QueryValuesFromStruct(payload)

	if err != nil {
		return nil, err
	}
	baseURL.RawQuery = params.Encode()

	return http.NewRequestWithContext(ctx, method, baseURL.String(), nil)
}
