package mpesa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"
)

// TestHelperFunctions is a dummy test to ensure this file is compiled.
func TestHelperFunctions(t *testing.T) {}

// allRedirectingTransport is a shared http.RoundTripper that redirects ALL requests
// to a single mock server URL. It preserves the original path and query.
type allRedirectingTransport struct {
	targetURL *url.URL
	transport http.RoundTripper
}

func (t *allRedirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	clonedReq.URL.Scheme = t.targetURL.Scheme
	clonedReq.URL.Host = t.targetURL.Host
	clonedReq.Host = t.targetURL.Host

	if t.transport == nil {
		return http.DefaultTransport.RoundTrip(clonedReq)
	}
	return t.transport.RoundTrip(clonedReq)
}

// newTestClientWithKeys generates an RSA key pair, creates a new M-Pesa client,
// and sets the public key on the client. It returns the configured client,
// the base64 encoded public key, and the private key for decryption in tests.
func newTestClientWithKeys(t *testing.T) (*Client, string, *rsa.PrivateKey) {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA private key: %v", err)
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	pubKeyBase64 := base64.StdEncoding.EncodeToString(pubASN1)

	client, err := NewClient("test-api-key", Sandbox, 30)
	if err != nil {
		t.Fatalf("Failed to create new client: %v", err)
	}
	client.Keys.PublicKey = pubKeyBase64

	return client, pubKeyBase64, privKey
}
