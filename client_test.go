package mpesa

import (
	"bytes"
	"net/url" // Added for redirectingTransport
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time" // Added to fix undefined: time error
	"encoding/pem"
)

// redirectingTransport is a custom http.RoundTripper to redirect specific requests to a mock server.
// It checks if a request is targeted for a specific host and path (like the M-Pesa session endpoint)
// and, if so, rewrites its URL to go to the mockServerURL, sending it via transportToMockServer.
// Other requests are passed through to defaultTransport.

type redirectingTransport struct {
	defaultTransport      http.RoundTripper // mpesa.Client's original transport, for non-redirected requests.
	transportToMockServer http.RoundTripper // httptest.Server's transport, for redirected requests.
	mockServerURL         *url.URL          // Parsed URL of the mock server (e.g., http://127.0.0.1:xxxx)
	targetScheme          string            // Scheme of the original request to intercept (e.g., "https")
	targetHost            string            // Host of the original request to intercept (e.g., "sandbox.safaricom.co.ke")
	targetPath            string            // Path of the original request to intercept (e.g., SessionEndPath)
}

// multiRedirectingTransport can redirect to different mock servers based on the request path.
type multiRedirectingTransport struct {
	defaultTransport http.RoundTripper
	redirects        map[string]*url.URL
}

func (mrt *multiRedirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for path, mockURL := range mrt.redirects {
		if strings.Contains(req.URL.Path, path) {
			clonedReq := req.Clone(req.Context())
			clonedReq.URL.Scheme = mockURL.Scheme
			clonedReq.URL.Host = mockURL.Host
			return http.DefaultTransport.RoundTrip(clonedReq)
		}
	}
	return mrt.defaultTransport.RoundTrip(req)
}

func (rt *redirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == rt.targetScheme && req.URL.Host == rt.targetHost && req.URL.Path == rt.targetPath {
		// Clone the request to avoid modifying the original request shared by other middlewares or retries.
		clonedReq := req.Clone(req.Context())
		
		// Rewrite the URL to point to the mock server.
		clonedReq.URL.Scheme = rt.mockServerURL.Scheme
		clonedReq.URL.Host = rt.mockServerURL.Host
		// The path remains the same (rt.targetPath), as mock server should handle it.

		// If the original request had a specific Host header, it might need to be preserved
		// or set to the mock server's host, depending on server requirements.
		// For typical httptest.Server usage, this is not strictly necessary as it doesn't usually check Host header.
		// clonedReq.Host = rt.targetHost // Example: if server validated Host header.
		return rt.transportToMockServer.RoundTrip(clonedReq)
	}
	// If not the target request, pass it through to the original default transport.
	return rt.defaultTransport.RoundTrip(req)
}

func TestNewClient(t *testing.T) {
	t.Run("SandboxClientCreation", func(t *testing.T) {
		apiKey := "testapikey"
		client, err := NewClient(apiKey, Sandbox, 30)
		if err != nil {
			t.Errorf("NewClient() error = %v, wantErr nil", err)
			return
		}
		if client == nil {
			t.Errorf("NewClient() client = nil, want non-nil client")
			return
		}
		if client.Keys.ApiKey != apiKey {
			t.Errorf("NewClient() ApiKey = %v, want %v", client.Keys.ApiKey, apiKey)
		}
		if client.Environment != Sandbox {
			t.Errorf("NewClient() Environment = %v, want %v", client.Environment, Sandbox)
		}
		if client.Keys.PublicKey != SandboxPublicKey {
			t.Errorf("NewClient() PublicKey = %v, want %v", client.Keys.PublicKey, SandboxPublicKey)
		}
	})

	t.Run("ProductionClientCreation", func(t *testing.T) {
		apiKey := "testprodapikey"
		client, err := NewClient(apiKey, Production, 30)
		if err != nil {
			t.Errorf("NewClient() error = %v, wantErr nil", err)
			return
		}
		if client == nil {
			t.Errorf("NewClient() client = nil, want non-nil client")
			return
		}
		if client.Keys.ApiKey != apiKey {
			t.Errorf("NewClient() ApiKey = %v, want %v", client.Keys.ApiKey, apiKey)
		}
		if client.Environment != Production {
			t.Errorf("NewClient() Environment = %v, want %v", client.Environment, Production)
		}
		if client.Keys.PublicKey != OpenapiPublicKey {
			t.Errorf("NewClient() PublicKey = %v, want %v", client.Keys.PublicKey, OpenapiPublicKey)
		}
	})

	t.Run("EmptyAPIKey", func(t *testing.T) {
		_, err := NewClient("", Sandbox, 30)
		if err == nil {
			t.Errorf("NewClient() error = nil, wantErr 'api Key is to create a Client'")
		} else if err.Error() != "api Key is to create a Client" {
			t.Errorf("NewClient() error = %v, wantErr 'api Key is to create a Client'", err)
		}
	})
}

func TestMakeUrl(t *testing.T) {
	clientSandbox, _ := NewClient("testkey", Sandbox, 30)
	clientProduction, _ := NewClient("testkey", Production, 30)

	tests := []struct {
		name     string
		client   *Client
		endpoint string
		want     string
	}{
		{
			name:     "Sandbox URL",
			client:   clientSandbox,
			endpoint: "testendpoint", // Removed leading slash
			want:     "https://openapi.m-pesa.com:443/sandbox/ipg/v2/vodacomTZN/testendpoint/",
		},
		{
			name:     "Production URL",
			client:   clientProduction,
			endpoint: "anotherendpoint", // Removed leading slash
			want:     "https://openapi.m-pesa.com:443/openapi/ipg/v2/vodacomTZN/anotherendpoint/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.client.makeUrl(tt.endpoint)
			if got != tt.want {
				t.Errorf("makeUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFmtPubKey(t *testing.T) {
	// Create a dummy client, the actual client details don't matter for this test
	client, _ := NewClient("dummyAPIKey", Sandbox, 30)

	rawPublicKey := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxX2X/vaYx4mY..."
	// The expected format should match what fmtPubKey produces
	expectedFormattedKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxX2X/vaYx4mY...
-----END PUBLIC KEY-----`

	t.Run("FormatPublicKey", func(t *testing.T) {
		formattedKey := client.fmtPubKey(rawPublicKey)
		if formattedKey != expectedFormattedKey {
			t.Errorf("fmtPubKey() got = %v, want %v", formattedKey, expectedFormattedKey)
		}
	})
}

func TestQueryValuesFromStruct(t *testing.T) {
	client, _ := NewClient("dummyAPIKey", Sandbox, 30) // Client isn't strictly needed but good for consistency

	type testStructSimple struct {
		Param1 string `json:"param_one"`
		Param2 int    `json:"param_two"`
	}

	type testStructMixedTags struct {
		TaggedField   string `json:"tagged_field"`
		UntaggedField string
		AnotherTagged int `json:"another_tagged"`
	}

	type testStructEmptyValues struct {
		FieldA string `json:"field_a"`
		FieldB string `json:"field_b"`
	}

	type testStructDifferentTypes struct {
		StringField string `json:"string_field"`
		IntField    int    `json:"int_field"`
		BoolField   bool   `json:"bool_field"`
	}

	tests := []struct {
		name    string
		payload interface{}
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "Simple struct with tags",
			payload: testStructSimple{Param1: "value1", Param2: 123},
			want:    map[string]string{"param_one": "value1", "param_two": "123"},
			wantErr: false,
		},
		{
			name:    "Struct with mixed tags",
			payload: testStructMixedTags{TaggedField: "hello", UntaggedField: "world", AnotherTagged: 456},
			want:    map[string]string{"tagged_field": "hello", "another_tagged": "456"},
			wantErr: false,
		},
		{
			name:    "Struct with empty values",
			payload: testStructEmptyValues{FieldA: "", FieldB: "non-empty"},
			want:    map[string]string{"field_a": "", "field_b": "non-empty"},
			wantErr: false,
		},
		{
			name:    "Struct with different types",
			payload: testStructDifferentTypes{StringField: "text", IntField: 789, BoolField: true},
			want:    map[string]string{"string_field": "text", "int_field": "789", "bool_field": "true"},
			wantErr: false,
		},
		{
			name:    "Non-struct payload",
			payload: "not a struct",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Nil payload",
			payload: nil,
			want:    nil,
			wantErr: true, // reflect.ValueOf(nil) is not a struct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.QueryValuesFromStruct(tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("QueryValuesFromStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != "payload is not a struct" && tt.payload != nil { // Specific error check for non-nil, non-struct
					t.Errorf("QueryValuesFromStruct() error = %v, want 'payload is not a struct'", err)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("QueryValuesFromStruct() got %v values, want %v values. Got: %v, Want: %v", len(got), len(tt.want), got, tt.want)
				return
			}

			for k, expectedValue := range tt.want {
				actualValue := got.Get(k)
				if actualValue != expectedValue {
					t.Errorf("QueryValuesFromStruct() for key %s: got %v, want %v", k, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	client, _ := NewClient("dummyAPIKey", Sandbox, 30)
	ctx := context.Background()

	type simplePayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string
		method      string
		url         string
		payload     interface{}
		wantBody    string // Expected JSON string for the body, or empty if no body/error
		wantErr     bool
		checkBody   bool
	}{
		{
			name:        "GET request with nil payload",
			method:      http.MethodGet,
			url:         "http://localhost/test",
			payload:     nil,
			wantBody:    "",
			wantErr:     false,
			checkBody:   false, // No body to check for GET with nil payload
		},
		{
			name:        "POST request with struct payload",
			method:      http.MethodPost,
			url:         "http://localhost/submit",
			payload:     simplePayload{Name: "Alice", Age: 30},
			wantBody:    "{\"name\":\"Alice\",\"age\":30}", // Corrected expected body
			wantErr:     false,
			checkBody:   true,
		},
		{
			name:        "PUT request with map payload",
			method:      http.MethodPut,
			url:         "http://localhost/update",
			payload:     map[string]interface{}{"key": "value", "id": 123},
			wantBody:    "{\"id\":123,\"key\":\"value\"}", // Corrected expected body (order might vary)
			wantErr:     false,
			checkBody:   true,
		},
		{
			name:        "Payload causing marshal error",
			method:      http.MethodPost,
			url:         "http://localhost/error",
			payload:     make(chan int), // Cannot be marshalled
			wantBody:    "",
			wantErr:     true,
			checkBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.NewRequest(ctx, tt.method, tt.url, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				// If an error was expected, no further checks on req
				return
			}

			if req == nil {
				t.Errorf("NewRequest() returned nil request, want non-nil")
				return
			}

			if req.Method != tt.method {
				t.Errorf("NewRequest() method = %v, want %v", req.Method, tt.method)
			}

			if req.URL.String() != tt.url {
				t.Errorf("NewRequest() URL = %v, want %v", req.URL.String(), tt.url)
			}

			if tt.checkBody {
				if req.Body == nil && tt.wantBody != "" {
					t.Errorf("NewRequest() body is nil, want body %s", tt.wantBody)
					return
				}
				if req.Body != nil {
					bodyBytes, ioErr := io.ReadAll(req.Body)
					if ioErr != nil {
						t.Fatalf("Failed to read request body: %v", ioErr)
					}
					gotBody := string(bodyBytes)
					if gotBody != tt.wantBody {
						t.Errorf("NewRequest() body = %s, want %s", gotBody, tt.wantBody)
					}
					// Restore the body for potential future reads if necessary, though not typical in unit tests
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				} else if tt.wantBody != "" { // req.Body is nil but we expected a body
				    t.Errorf("NewRequest() body is nil, want body %s", tt.wantBody)
				}
			}
		})
	}
}

func TestNewReqWithQueryParams(t *testing.T) {
	client, _ := NewClient("dummyAPIKey", Sandbox, 30)
	ctx := context.Background()

	type queryPayload struct {
		Query1 string `json:"q1"`
		Query2 int    `json:"q2"`
	}

	type emptyQueryPayload struct {
		NoTagField string
	}

	tests := []struct {
		name        string
		method      string
		baseUrl     string
		payload     interface{}
		wantURL     string
		wantErr     bool
		expErrStr   string // Expected error string if wantErr is true
	}{
		{
			name:    "Valid GET request with query params",
			method:  http.MethodGet,
			baseUrl: "http://localhost/api/items",
			payload: queryPayload{Query1: "test", Query2: 123},
			wantURL: "http://localhost/api/items?q1=test&q2=123",
			wantErr: false,
		},
		{
			name:    "Base URL with existing query params (should be overwritten)",
			method:  http.MethodGet,
			baseUrl: "http://localhost/api/items?existing=true",
			payload: queryPayload{Query1: "new", Query2: 456},
			wantURL: "http://localhost/api/items?q1=new&q2=456",
			wantErr: false,
		},
		{
			name:    "Payload with no tagged fields",
			method:  http.MethodGet,
			baseUrl: "http://localhost/api/search",
			payload: emptyQueryPayload{NoTagField: "value"},
			wantURL: "http://localhost/api/search", // No query params expected
			wantErr: false,
		},
		{
			name:      "Nil payload",
			method:    http.MethodPost,
			baseUrl:   "http://localhost/api/action",
			payload:   nil,
			wantURL:   "",
			wantErr:   true,
			expErrStr: "payload is not a struct",
		},
		{
			name:      "Non-struct payload",
			method:    http.MethodDelete,
			baseUrl:   "http://localhost/api/resource",
			payload:   "not-a-struct",
			wantURL:   "",
			wantErr:   true,
			expErrStr: "payload is not a struct",
		},
		{
			name:      "Invalid base URL",
			method:    http.MethodGet,
			baseUrl:   "://invalid-url", // Scheme is missing
			payload:   queryPayload{Query1: "any", Query2: 0},
			wantURL:   "",
			wantErr:   true,
			// Error string from url.Parse can vary, so we check for non-nil error only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.NewReqWithQueryParams(ctx, tt.method, tt.baseUrl, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewReqWithQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expErrStr != "" && err.Error() != tt.expErrStr {
					t.Errorf("NewReqWithQueryParams() error string = '%s', want '%s'", err.Error(), tt.expErrStr)
				}
				// If an error was expected, no further checks on req
				return
			}

			if req == nil {
				t.Errorf("NewReqWithQueryParams() returned nil request, want non-nil")
				return
			}

			if req.Method != tt.method {
				t.Errorf("NewReqWithQueryParams() method = %v, want %v", req.Method, tt.method)
			}

			if req.URL.String() != tt.wantURL {
				t.Errorf("NewReqWithQueryParams() URL = %v, want %v", req.URL.String(), tt.wantURL)
			}

			if req.Body != nil {
				t.Errorf("NewReqWithQueryParams() Body is not nil, want nil")
			}
		})
	}
}

func TestCreateBearerToken(t *testing.T) {
	// Generate a new RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test RSA key pair: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Convert the test public key to the raw string format expected by fmtPubKey
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	rawPublicKeyStr := base64.StdEncoding.EncodeToString(pubASN1)

	client := &Client{
		Keys: &Keys{
			PublicKey: rawPublicKeyStr,
		},
		Environment: Sandbox,
	}

	t.Run("SuccessfulEncryption", func(t *testing.T) {
		apiKey := "mySecretApiKey123"
		bearerToken, err := client.createBearerToken(apiKey)

		if err != nil {
			t.Fatalf("createBearerToken() error = %v, wantErr nil", err)
		}
		if bearerToken == "" {
			t.Fatal("createBearerToken() returned empty token, want non-empty")
		}

		decodedCiphertext, err := base64.StdEncoding.DecodeString(bearerToken)
		if err != nil {
			t.Fatalf("Failed to base64 decode bearer token: %v", err)
		}

		decryptedAPIKeyBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decodedCiphertext)
		if err != nil {
			t.Fatalf("Failed to decrypt token: %v", err)
		}

		if string(decryptedAPIKeyBytes) != apiKey {
			t.Errorf("Decrypted API key = %s, want %s", string(decryptedAPIKeyBytes), apiKey)
		}
	})

	t.Run("MalformedPublicKey_NonPKIX", func(t *testing.T) {
		malformedClient := &Client{
			Keys: &Keys{
				PublicKey: base64.StdEncoding.EncodeToString([]byte("not a real der encoded pkix key")),
			},
			Environment: Sandbox,
		}
		_, err := malformedClient.createBearerToken("anyApiKey")
		if err == nil {
			t.Error("createBearerToken() with malformed public key, error = nil, want error")
		}
	})

	t.Run("MalformedPublicKey_EmptyString", func(t *testing.T) {
		extremelyMalformedClient := &Client{
			Keys: &Keys{
				PublicKey: "", 
			},
			Environment: Sandbox,
		}
		_, err := extremelyMalformedClient.createBearerToken("anyApiKey")
		if err == nil {
			t.Errorf("createBearerToken() with empty PublicKey string, error = nil, want error")
		}
	})

	t.Run("MalformedPublicKey_InvalidPEMStructure", func(t *testing.T) {
		// This key is so malformed that pem.Decode will return nil for the block
		// (after fmtPubKey wraps it), triggering the new error check.
		malformedKey := "THIS IS NOT A VALID PEM ENCODED KEY AT ALL"
		// We need a client instance, but NewClient might try to use the public key early
		// or might not be suitable if we want to directly set a malformed key.
		// So, we construct a client and directly set its Keys.PublicKey.
		client := &Client{
			Keys: &Keys{
				PublicKey: malformedKey,
				ApiKey:    "testAPIKey", // ApiKey is needed by createBearerToken
			},
			Environment: Sandbox, // Environment might influence key formatting or other logic
		}

		_, err := client.createBearerToken(client.Keys.ApiKey)
		if err == nil {
			t.Fatal("createBearerToken() with completely invalid PEM structure, error = nil, want error")
		}

		expectedError := "failed to decode PEM block containing public key"
		if err.Error() != expectedError {
			t.Errorf("createBearerToken() error = %v, want %s", err, expectedError)
		}
	})
}

// mockResponsePayload is a helper struct for testing successful JSON responses.
type mockResponsePayload struct {
	Message string `json:"message"`
	Value   int    `json:"value"`
}

// mockErrorPayload is a helper struct for testing error JSON responses.
type mockErrorPayload struct {
	ErrorID  string `json:"errorId"`
	ErrorMsg string `json:"errorMessage"`
}

func TestSend(t *testing.T) {
	client := &Client{Client: &http.Client{}} // Basic client for Send method

	// Test case: Successful request, unmarshal into struct
	t.Run("SuccessfulRequest_UnmarshalStruct", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Accept") != "application/json" {
				t.Errorf("Expected Accept header 'application/json', got '%s'", r.Header.Get("Accept"))
			}
			// Host header check removed: httptest server receives its own host (e.g., 127.0.0.1:port),
			// not the req.Host value ('openapi.m-pesa.com') that client.Send sets for the outgoing request.
			if r.Header.Get("Origin") != "*" {
				t.Errorf("Expected Origin header '*', got '%s'", r.Header.Get("Origin"))
			}
			if r.Header.Get("Content-type") != "application/json" {
				t.Errorf("Expected Content-type header 'application/json', got '%s'", r.Header.Get("Content-type"))
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponsePayload{Message: "Success", Value: 123})
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)
		var respPayload mockResponsePayload

		err := client.Send(req, &respPayload, nil)
		if err != nil {
			t.Fatalf("Send() error = %v, want nil", err)
		}

		if respPayload.Message != "Success" || respPayload.Value != 123 {
			t.Errorf("Send() response payload = %+v, want %+v", respPayload, mockResponsePayload{Message: "Success", Value: 123})
		}
	})

	// Test case: Successful request, write to io.Writer
	t.Run("SuccessfulRequest_IoWriter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"IoWriter Success","value":456}`)) // Raw JSON string
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)
		var bodyBuffer bytes.Buffer

		err := client.Send(req, &bodyBuffer, nil)
		if err != nil {
			t.Fatalf("Send() error = %v, want nil", err)
		}

		expectedBody := `{"message":"IoWriter Success","value":456}`
		if strings.TrimSpace(bodyBuffer.String()) != expectedBody { // TrimSpace to handle potential newline from Encode
			t.Errorf("Send() response body = %s, want %s", bodyBuffer.String(), expectedBody)
		}
	})

	// Test case: Successful request, v is nil
	t.Run("SuccessfulRequest_VIsNil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"V is nil"}`)) 
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)

		err := client.Send(req, nil, nil)
		if err != nil {
			t.Fatalf("Send() error = %v, want nil when v is nil", err)
		}
	})

	// Test case: HTTP error, unmarshal into error struct
	t.Run("ErrorRequest_UnmarshalErrorStruct", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(mockErrorPayload{ErrorID: "E001", ErrorMsg: "Bad Request Occurred"})
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)
		var errPayload mockErrorPayload

		err := client.Send(req, nil, &errPayload)
		if err == nil {
			t.Fatal("Send() error = nil, want error for status 400")
		}

		// Check if the error message from Send() indicates the status code and that 'e' was used.
		expectedOuterError := fmt.Sprintf("API error (status %d), payload unmarshalled into provided type: %v", http.StatusBadRequest, &errPayload)
		if err.Error() != expectedOuterError {
			t.Errorf("Send() outer error = %v, want %s", err, expectedOuterError)
		}

		// Check if the error payload was populated
		if errPayload.ErrorID != "E001" || errPayload.ErrorMsg != "Bad Request Occurred" {
			t.Errorf("Send() error payload = %+v, want %+v", errPayload, mockErrorPayload{ErrorID: "E001", ErrorMsg: "Bad Request Occurred"})
		}
	})

	// Test case: HTTP error, e is nil
	t.Run("ErrorRequest_EIsNil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error Details"))
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)

		err := client.Send(req, nil, nil)
		if err == nil {
			t.Fatal("Send() error = nil, want error for status 500 with e=nil")
		}

		// When 'e' is nil and unmarshal into MpesaError fails (because body is not MpesaError compatible JSON)
		expectedErrorMsg := fmt.Sprintf("API request failed with status %d: %s", http.StatusInternalServerError, "Internal Server Error Details")
		if err.Error() != expectedErrorMsg {
			t.Errorf("Send() error = %v, want %s", err, expectedErrorMsg)
		}
	})

	// Test case: HTTP error, unmarshal error fails (e.g. bad JSON in error response)
	t.Run("ErrorRequest_UnmarshalErrorFails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("This is not valid JSON"))
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL, nil)
		var errPayload mockErrorPayload

		err := client.Send(req, nil, &errPayload)
		if err == nil {
			t.Fatal("Send() error = nil, want error for status 409 with unmarshal failure")
		}

		// When 'e' is provided, but unmarshal into 'e' fails (because body is not JSON for 'e'),
		// and unmarshal into MpesaError also fails, sendLocked returns a generic message.
		expectedErrorMsg := fmt.Sprintf("API request failed with status %d: %s", http.StatusConflict, "This is not valid JSON")
		if err.Error() != expectedErrorMsg {
			t.Errorf("Send() error = %v, want error containing 'invalid character'", err)
		}
	})

	// Test case: Client.Do fails (e.g., network error)
	t.Run("ClientDoError", func(t *testing.T) {
		// Create a client that will always fail by providing a non-existent URL for the request
		// The server itself is not used here, but the request is made to an invalid address.
		req, _ := http.NewRequest("GET", "http://localhost:0/nonexistent", nil) // Port 0 is usually invalid

		err := client.Send(req, nil, nil)
		if err == nil {
			t.Fatal("Send() error = nil, want error for Client.Do failure")
		}
		// We can't check the exact error message as it's OS/network-dependent,
		// but it should be a URL error or connection error.
		// Check for common network error substrings.
		isNetworkError := strings.Contains(err.Error(), "dial tcp") ||
			strings.Contains(err.Error(), "lookup") ||
			strings.Contains(err.Error(), "connect: connection refused") ||
			strings.Contains(err.Error(), "no such host")
		if !isNetworkError {
			t.Errorf("Send() error = %v, want a network dial or lookup error", err)
		}
	})
} // Closing bracket for TestSend

// TestSendWithAuth tests the SendWithAuth method of the M-Pesa client.
func TestSendWithAuth(t *testing.T) {
	// Generate a real key pair for testing successful token generation
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA private key: %v", err)
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	actualGoodPubKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}))

	// Use a consistent API key for tests
	const testAPIKey = "testapikeyforsendwithauth"

	client, clientErr := NewClient(testAPIKey, Sandbox, 30) // Use NewClient that returns error
	if clientErr != nil {
		t.Fatalf("Failed to create client for TestSendWithAuth: %v", clientErr)
	}
	client.Keys.PublicKey = actualGoodPubKey // Override with our generated key

	// Test case 1: Successful request
	t.Run("SuccessfulRequest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
			}
			if len(strings.TrimPrefix(authHeader, "Bearer ")) == 0 {
				t.Errorf("Authorization token is empty")
			}

			if r.Header.Get("Accept") != "application/json" {
				t.Errorf("Expected Accept header 'application/json', got '%s'", r.Header.Get("Accept"))
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponsePayload{Message: "AuthSuccess", Value: 456})
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL+"/testauth", nil)
		var respPayload mockResponsePayload
		err := client.SendWithAuth(req, &respPayload, nil)

		if err != nil {
			t.Fatalf("SendWithAuth failed: %v", err)
		}
		if respPayload.Message != "AuthSuccess" || respPayload.Value != 456 {
			t.Errorf("Expected response payload {AuthSuccess 456}, got %+v", respPayload)
		}
	})

	// Test case 2: createBearerToken fails
	t.Run("CreateBearerTokenFails", func(t *testing.T) {
		badClient, badClientErr := NewClient(testAPIKey, Sandbox, 30)
		if badClientErr != nil {
			t.Fatalf("Failed to create badClient for TestSendWithAuth: %v", badClientErr)
		}
		badClient.Keys.PublicKey = "MALFORMED PUBLIC KEY" // Force createBearerToken to fail

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Server should not be called when createBearerToken fails")
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL+"/testauthfail", nil)
		var respPayload mockResponsePayload
		err := badClient.SendWithAuth(req, &respPayload, nil)

		if err == nil {
			t.Fatal("SendWithAuth should have failed due to createBearerToken error, but it succeeded")
		}
		expectedErrStr := "failed to decode PEM block containing public key"
		if !strings.Contains(err.Error(), expectedErrStr) {
			t.Errorf("Expected error to contain '%s', got '%v'", expectedErrStr, err)
		}
	})

	// Test case 3: Send fails (e.g., HTTP error from server)
	t.Run("SendFails_HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			// Use mockErrorPayload consistent with other tests
			json.NewEncoder(w).Encode(mockErrorPayload{ErrorID: "SE001", ErrorMsg: "ServerExploded"})
		}))
		defer server.Close()

		req, _ := http.NewRequest("GET", server.URL+"/testsendfail", nil)
		var respPayload mockResponsePayload // Won't be populated
		var errPayload mockErrorPayload
		err := client.SendWithAuth(req, &respPayload, &errPayload)

		if err == nil {
			t.Fatal("SendWithAuth should have failed due to underlying Send error, but it succeeded")
		}

		if errPayload.ErrorID != "SE001" || errPayload.ErrorMsg != "ServerExploded" {
			t.Errorf("Expected error payload {SE001 ServerExploded}, got %+v", errPayload)
		}
		
		// The error string now includes the type of the unmarshalled payload if 'e' was provided.
		// Note: errPayload was populated by sendLocked because it was passed as 'e'.
		expectedErrStr := fmt.Sprintf("API error (status %d), payload unmarshalled into provided type: %v", http.StatusInternalServerError, &errPayload)
		if err.Error() != expectedErrStr {
			t.Errorf("Expected error message '%s', got '%v'", expectedErrStr, err)
		}
	})
}

func TestSetHttpClient(t *testing.T) {
	client, err := NewClient("testapikey", Sandbox, 30)
	if err != nil {
		t.Fatalf("NewClient() error = %v, wantErr nil", err)
	}

	originalTimeout := client.Client.Timeout

	customTimeout := time.Second * 60
	customHttpClient := &http.Client{
		Timeout: customTimeout,
	}

	client.SetHttpClient(customHttpClient)

	if client.Client.Timeout != customTimeout {
		t.Errorf("SetHttpClient() failed to set custom client. Expected timeout %v, got %v", customTimeout, client.Client.Timeout)
	}

	// Optional: Test that it's not the original client by comparing more than just timeout if necessary
	if client.Client == http.DefaultClient {
		t.Errorf("SetHttpClient() client is still http.DefaultClient, expected custom client")
	}

	// Restore original timeout for other tests if this was a global or shared client (not the case here as NewClient creates a new one)
	// For this specific test structure, it's fine as client is local to this test.
	// If client were somehow shared, you'd do: client.Client.Timeout = originalTimeout
	_ = originalTimeout // use originalTimeout to avoid unused variable error if the above optional check is removed
}

// TestGenSessionKey tests the genSessionKey method of the M-Pesa client.
func TestGenSessionKey(t *testing.T) {
	// Generate a real key pair for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA private key: %v", err)
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	actualGoodPubKeyBase64 := base64.StdEncoding.EncodeToString(pubASN1) // Raw base64 of PKIX key

	const testAPIKey = "testapikeyforgensession"

	// Test case 1: Successful session key retrieval
	t.Run("SuccessfulSessionKeyRetrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != SandboxEndpoint+SessionEndPath+"/" { // Check for full path with trailing slash
				t.Errorf("Expected path '%s', got '%s'", SandboxEndpoint+SessionEndPath+"/", r.URL.Path)
			}
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
			}
			if len(strings.TrimPrefix(authHeader, "Bearer ")) == 0 {
				t.Errorf("Authorization token is empty")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SessionKeyResponse{
				OutputResponseCode: "0",
				OutputResponseDesc: "Session ID generated successfully",
				OutputSessionID:    "testSessionKey123",
			})
		}))
		defer server.Close()

		client, clientErr := NewClient(testAPIKey, Sandbox, 30)
		if clientErr != nil {
			t.Fatalf("Failed to create client: %v", clientErr)
		}
		client.Keys.PublicKey = actualGoodPubKeyBase64

		parsedMockURL, parseErr := url.Parse(server.URL)
		if parseErr != nil {
			t.Fatalf("Failed to parse mock server URL: %v", parseErr)
		}

		originalClientTransport := client.Client.Transport
		if originalClientTransport == nil {
			originalClientTransport = http.DefaultTransport
		}
		transportToMock := server.Client().Transport
		if transportToMock == nil {
			transportToMock = http.DefaultTransport
		}

		client.Client.Transport = &redirectingTransport{
			defaultTransport:      originalClientTransport,
			transportToMockServer: transportToMock,
			mockServerURL:         parsedMockURL,
			targetScheme:          "https",
			targetHost:            fmt.Sprintf("%s:%d", Address, Port),
			targetPath:            SandboxEndpoint + SessionEndPath + "/",
		}

		sessionResp, err := client.genSessionKey()
		if err != nil {
			t.Fatalf("genSessionKey failed: %v", err)
		}
		if sessionResp == nil {
			t.Fatal("genSessionKey returned nil response")
		}
		if sessionResp.OutputSessionID != "testSessionKey123" {
			t.Errorf("Expected session ID 'testSessionKey123', got '%s'", sessionResp.OutputSessionID)
		}
	})

	// Test case 2: createBearerToken fails (leading to genSessionKey failure)
	t.Run("CreateBearerTokenFails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Server should not be called when token generation fails")
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, clientErr := NewClient(testAPIKey, Sandbox, 30)
		if clientErr != nil {
			t.Fatalf("Failed to create client: %v", clientErr)
		}
		client.Keys.PublicKey = "MALFORMED PUBLIC KEY" // Force token generation error

		_, err := client.genSessionKey()
		if err == nil {
			t.Fatal("genSessionKey should have failed due to token generation error, but it succeeded")
		}
		expectedErrStr := "failed to decode PEM block containing public key"
		if !strings.Contains(err.Error(), expectedErrStr) {
			t.Errorf("Expected error to contain '%s', got '%v'", expectedErrStr, err)
		}
	})

	// Test case 3: Server returns an error (SendWithAuth fails)
	t.Run("ServerReturnsError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized) // Simulate an auth error from M-Pesa
			// Use fmt.Fprint to avoid trailing newline. 
			// Respond with fields matching MpesaError for direct unmarshalling.
			fmt.Fprint(w, `{"output_ResponseCode":"AUTH-001","output_ResponseDesc":"Invalid credentials"}`)
		}))
		defer server.Close()

		client, clientErr := NewClient(testAPIKey, Sandbox, 30)
		if clientErr != nil {
			t.Fatalf("Failed to create client: %v", clientErr)
		}
		client.Keys.PublicKey = actualGoodPubKeyBase64

		parsedMockURL, parseErr := url.Parse(server.URL)
		if parseErr != nil {
			t.Fatalf("Failed to parse mock server URL: %v", parseErr)
		}

		originalClientTransport := client.Client.Transport
		if originalClientTransport == nil {
			originalClientTransport = http.DefaultTransport
		}
		transportToMock := server.Client().Transport
		if transportToMock == nil {
			transportToMock = http.DefaultTransport
		}

		client.Client.Transport = &redirectingTransport{
			defaultTransport:      originalClientTransport,
			transportToMockServer: transportToMock,
			mockServerURL:         parsedMockURL,
			targetScheme:          "https",
			targetHost:            fmt.Sprintf("%s:%d", Address, Port),
			targetPath:            SandboxEndpoint + SessionEndPath + "/",
		}

		_, err := client.genSessionKey()
		if err == nil {
			t.Fatal("genSessionKey should have failed due to server error, but it succeeded")
		}

		mpesaErr, ok := err.(*MpesaError)
		if !ok {
			t.Fatalf("Expected error of type *MpesaError, got %T: %v", err, err)
		}

		expectedResponseCode := "AUTH-001"
		expectedResponseDesc := "Invalid credentials"
		expectedStatusCode := http.StatusUnauthorized

		if mpesaErr.ResponseCode != expectedResponseCode {
			t.Errorf("Expected MpesaError.ResponseCode '%s', got '%s'", expectedResponseCode, mpesaErr.ResponseCode)
		}
		if mpesaErr.ResponseDesc != expectedResponseDesc {
			t.Errorf("Expected MpesaError.ResponseDesc '%s', got '%s'", expectedResponseDesc, mpesaErr.ResponseDesc)
		}
		if mpesaErr.HTTPStatusCode != expectedStatusCode {
			t.Errorf("Expected MpesaError.HTTPStatusCode %d, got %d", expectedStatusCode, mpesaErr.HTTPStatusCode)
		}
		// Optional: Check the full error string if needed, though field checks are more robust.
		// expectedFullErrStr := fmt.Sprintf("M-Pesa API Error: Code %s - %s (HTTP Status: %d)", expectedResponseCode, expectedResponseDesc, expectedStatusCode)
		// if mpesaErr.Error() != expectedFullErrStr {
		// 	t.Errorf("Expected full error string '%s', got '%s'", expectedFullErrStr, mpesaErr.Error())
		// }
	})
}

// TestSendWithSessionKey tests the SendWithSessionKey method of the M-Pesa client.
func TestSendWithSessionKey(t *testing.T) {
	// Shared RSA key setup for createBearerToken calls within sub-tests
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA private key: %v", err)
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	actualGoodPubKeyBase64 := base64.StdEncoding.EncodeToString(pubASN1) // Raw base64 of PKIX key

	const testAPIKey = "testapikeyforsendwithsession"

	t.Run("SuccessfulRequestWithNewSessionKey", func(t *testing.T) {
		const testSessionIDFromMock = "mockSessionID12345" // Specific to this sub-test
		// Mock server for the final data request made by c.Send()
		finalReqServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Verify Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				t.Errorf("Expected Authorization header, got none")
				return
			}
			if !strings.HasPrefix(authHeader, "Bearer ") {
				t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
				return
			}
			encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
			decodedToken, err := base64.StdEncoding.DecodeString(encryptedToken)
			if err != nil {
				t.Errorf("Failed to base64 decode bearer token: %v", err)
				return
			}
			decryptedSessionKeyBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, decodedToken)
			if err != nil {
				t.Errorf("Failed to decrypt session key from token: %v", err)
				return
			}
			if string(decryptedSessionKeyBytes) != testSessionIDFromMock {
				t.Errorf("Decrypted session key '%s' does not match expected '%s'", string(decryptedSessionKeyBytes), testSessionIDFromMock)
			}

            // 2. Verify request path
            expectedPath := SandboxEndpoint + "/test/endpoint" + "/" // As per client.makeUrl
            if r.URL.Path != expectedPath {
                t.Errorf("Expected request to path '%s', got '%s'", expectedPath, r.URL.Path)
            }
            w.WriteHeader(http.StatusOK)
            fmt.Fprint(w, `{"status":"ok"}`)
        }))
        defer finalReqServer.Close()

        // Mock server for genSessionKey's internal HTTP call
        sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Path != SandboxEndpoint+SessionEndPath+"/" {
                t.Errorf("genSessionKey mock: Expected path '%s', got '%s'", SandboxEndpoint+SessionEndPath+"/", r.URL.Path)
            }
            w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SessionKeyResponse{
				OutputResponseCode: "0",
				OutputResponseDesc: "Session ID generated successfully for SendWithSessionKey test",
				OutputSessionID:    testSessionIDFromMock,
			})
        }))
        defer sessionServer.Close()

        client, clientErr := NewClient(testAPIKey, Sandbox, 30)
        if clientErr != nil {
            t.Fatalf("Failed to create client: %v", clientErr)
        }
        client.Keys.PublicKey = actualGoodPubKeyBase64 // Used by genSessionKey's SendWithAuth

        // Configure client to use redirectingTransport for genSessionKey calls
        parsedSessionMockURL, _ := url.Parse(sessionServer.URL)
        originalClientTransport := client.Client.Transport
        if originalClientTransport == nil { originalClientTransport = http.DefaultTransport }
        transportToSessionMock := sessionServer.Client().Transport
        if transportToSessionMock == nil { transportToSessionMock = http.DefaultTransport }

        client.Client.Transport = &redirectingTransport{
            defaultTransport:      originalClientTransport,
            transportToMockServer: transportToSessionMock,
            mockServerURL:         parsedSessionMockURL,
            targetScheme:          "https",
            targetHost:            fmt.Sprintf("%s:%d", Address, Port),
            targetPath:            SandboxEndpoint + SessionEndPath + "/",
        }

        // Configure redirectingTransport for the final API call to finalReqServer
        // The targetHost and targetPath must match what client.makeUrl("/test/endpoint") would produce.
        parsedFinalReqURL, _ := url.Parse(finalReqServer.URL)

        // New approach for this sub-test: The existing transport on the client is for genSessionKey.
        // We need the *defaultTransport* of that redirector to point to *another* redirector for the final call.
        // Let currentSessionRedirectTransport = client.Client.Transport.(*redirectingTransport)
        // currentSessionRedirectTransport.defaultTransport = &redirectingTransport{ for finalReqServer }
        // This is the correct chaining.

        finalCallTargetURL := client.makeUrl("/test/endpoint") // This is what SendWithSessionKey will try to hit
        parsedFinalCallURL, _ := url.Parse(finalCallTargetURL)

        // This new redirector intercepts the actual API call and sends it to finalReqServer
        finalCallRedirector := &redirectingTransport{
            // defaultTransport here should be a real transport if we expect fall-through, or nil if all calls are mocked.
            // Assuming originalClientTransport was http.DefaultTransport or similar.
            defaultTransport:      originalClientTransport, // The one saved before session redirector was set
            transportToMockServer: finalReqServer.Client().Transport,
            mockServerURL:         parsedFinalReqURL, // finalReqServer's URL
            targetScheme:          parsedFinalCallURL.Scheme,
            targetHost:            parsedFinalCallURL.Host,
            targetPath:            parsedFinalCallURL.Path, // Ensure this matches makeUrl's output exactly
        }

        // The client currently has a redirector for genSessionKey. We need its defaultTransport
        // to be the finalCallRedirector.
        if rt, ok := client.Client.Transport.(*redirectingTransport); ok {
            rt.defaultTransport = finalCallRedirector // Chain the redirectors
        } else {
            t.Fatal("Expected client.Client.Transport to be *redirectingTransport for session key")
        }

        // Make the request that SendWithSessionKey would wrap
        req, _ := http.NewRequest("POST", client.makeUrl("/test/endpoint"), strings.NewReader("{}"))
        var result map[string]string
        err = client.SendWithSessionKey(req, &result, nil)

        if err != nil {
            t.Fatalf("SendWithSessionKey failed: %v", err)
        }
        if result == nil || result["status"] != "ok" {
            t.Errorf("Expected result {status:ok}, got %v", result)
        }
    })

    t.Run("SuccessfulRequestWithExistingValidSessionKey", func(t *testing.T) {
        const preExistingSessionID = "existingValidSessionID78910"

        // Mock server for the final data request made by c.Send()
        finalReqServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Verify Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                t.Errorf("Expected Authorization header, got none")
                return
            }
            if !strings.HasPrefix(authHeader, "Bearer ") {
                t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
                return
            }
            encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
            decodedToken, err := base64.StdEncoding.DecodeString(encryptedToken)
            if err != nil {
                t.Errorf("Failed to base64 decode bearer token: %v", err)
                return
            }
            decryptedSessionKeyBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, decodedToken)
            if err != nil {
                t.Errorf("Failed to decrypt session key from token: %v", err)
                return
            }
            if string(decryptedSessionKeyBytes) != preExistingSessionID {
                t.Errorf("Decrypted session key '%s' does not match expected '%s'", string(decryptedSessionKeyBytes), preExistingSessionID)
            }

            // 2. Verify request path
            expectedPath := SandboxEndpoint + "/test/another_endpoint" + "/" // As per client.makeUrl
            if r.URL.Path != expectedPath {
                t.Errorf("Expected request to path '%s', got '%s'", expectedPath, r.URL.Path)
            }
            w.WriteHeader(http.StatusOK)
            fmt.Fprint(w, `{"status":"ok_existing_key"}`)
        }))
        defer finalReqServer.Close()

        // No sessionServer mock is needed here, as genSessionKey should NOT be called.

        client, clientErr := NewClient(testAPIKey, Sandbox, 30)
        if clientErr != nil {
            t.Fatalf("Failed to create client: %v", clientErr)
        }
        // IMPORTANT: Pre-populate session key and ensure it's not expired
        client.SessionKey = preExistingSessionID
        client.ExpiresAt = time.Now().Add(time.Hour * 1) // Valid for 1 hour
        client.Keys.PublicKey = actualGoodPubKeyBase64 // For createBearerToken(sessionID) call

        // Configure redirectingTransport for the final API call to finalReqServer
        // Since genSessionKey is NOT called, client.Client.Transport is the original one (or http.DefaultTransport).
        finalCallTargetURL := client.makeUrl("/test/another_endpoint")
        parsedFinalCallURL, _ := url.Parse(finalCallTargetURL)
        parsedMockFinalReqURL, _ := url.Parse(finalReqServer.URL)

        client.Client.Transport = &redirectingTransport{
            defaultTransport:      http.DefaultTransport, // Or client.Client.Transport if it was set to something specific initially
            transportToMockServer: finalReqServer.Client().Transport,
            mockServerURL:         parsedMockFinalReqURL,
            targetScheme:          parsedFinalCallURL.Scheme,
            targetHost:            parsedFinalCallURL.Host,
            targetPath:            parsedFinalCallURL.Path,
        }

        // Make the request
        req, _ := http.NewRequest("GET", client.makeUrl("/test/another_endpoint"), nil)
        var result map[string]string
        err = client.SendWithSessionKey(req, &result, nil)

        if err != nil {
            t.Fatalf("SendWithSessionKey failed: %v", err)
        }
        if result == nil || result["status"] != "ok_existing_key" {
            t.Errorf("Expected result {status:ok_existing_key}, got %v", result)
        }
        // Additionally, one could verify that the sessionServer (if one was set up to count calls) was not called.
        // For now, the absence of errors from an unmocked genSessionKey call path is an implicit check.
    })

    t.Run("ErrorFromCreateBearerToken", func(t *testing.T) {
        const testSessionIDFromMock = "sessionForTokenFailTest"

        // Mock server for genSessionKey's internal HTTP call
        sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(SessionKeyResponse{
                OutputResponseCode: "0",
                OutputResponseDesc: "Session ID for token failure test",
                OutputSessionID:    testSessionIDFromMock,
            })
        }))
        defer sessionServer.Close()

        client, clientErr := NewClient(testAPIKey, Sandbox, 30)
        if clientErr != nil {
            t.Fatalf("Failed to create client: %v", clientErr)
        }
        // This key is for genSessionKey's SendWithAuth -> createBearerToken (encrypting API key)
        client.Keys.PublicKey = actualGoodPubKeyBase64
        // This key is for SendWithSessionKey -> createBearerToken (encrypting Session ID)
        // It uses client.Keys.PublicKey. Set it to an invalid value to cause createBearerToken to fail.
        // The API key encryption for genSessionKey (if it were called without a pre-set session key)
        // also uses client.Keys.PublicKey. So, if genSessionKey were to be called, it would also fail here.
        // For this test, we assume genSessionKey succeeds (or is bypassed by pre-setting session key)
        // and the failure is specifically in the createBearerToken(c.SessionKey) call.
        client.Keys.PublicKey = "this-is-not-a-valid-pem-encoded-public-key" // Corrected field name

        // Configure client to use redirectingTransport for genSessionKey calls
        parsedSessionMockURL, _ := url.Parse(sessionServer.URL)
        originalClientTransport := client.Client.Transport
        if originalClientTransport == nil { originalClientTransport = http.DefaultTransport }
        transportToSessionMock := sessionServer.Client().Transport
        if transportToSessionMock == nil { transportToSessionMock = http.DefaultTransport }

        client.Client.Transport = &redirectingTransport{
            defaultTransport:      originalClientTransport,
            transportToMockServer: transportToSessionMock,
            mockServerURL:         parsedSessionMockURL,
            targetScheme:          "https",
            targetHost:            fmt.Sprintf("%s:%d", Address, Port),
            targetPath:            SandboxEndpoint + SessionEndPath + "/",
        }

        // No finalReqServer is needed as the call to SendWithSessionKey should fail before c.Send is called.
        // The redirectingTransport for genSessionKey is still active.
        // If SendWithSessionKey logic changes to not pre-fetch session key when one is already set (even if createBearerToken fails),
        // this test might need adjustment. Current SendWithSessionKey fetches session key if c.SessionKey is empty OR expired.
        // If createBearerToken fails, it returns error before c.Send. So no network call for the final request happens.

        req, _ := http.NewRequest("POST", client.makeUrl("/test/should_fail_early"), strings.NewReader("{}"))
        var result map[string]string
        err = client.SendWithSessionKey(req, &result, nil)

        if err == nil {
            t.Fatal("Expected an error from SendWithSessionKey due to createBearerToken failure, got nil")
        }

        // createBearerToken returns fmt.Errorf("failed to encrypt session key: %w", err)
        // The wrapped error for invalid key is typically from pem.Decode or x509.ParsePKIXPublicKey
        // The error is from pem.Decode when an invalid public key string is used.
        expectedErrorPrefix := "failed to decode PEM block containing public key"
        if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
            t.Errorf("Expected error to start with '%s', got '%v'", expectedErrorPrefix, err)
        }
    })

    t.Run("ErrorFromUnderlyingSendCall", func(t *testing.T) {
        const testSessionIDForSendError = "sessionForSendErrorTest"
        type mockAPIError struct {
            RequestID    string `json:"requestId"` // Note: M-Pesa might use different field names
            ErrorCode    string `json:"errorCode"`
            ErrorMessage string `json:"errorMessage"`
        }
        expectedErrPayload := mockAPIError{
            RequestID:    "err-req-123",
            ErrorCode:    "GW_001",
            ErrorMessage: "The underlying service failed processing the request.",
        }

        // Mock server for the final data request made by c.Send()
        finalReqServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Verify Authorization header to ensure createBearerToken part worked
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                t.Errorf("Expected Authorization header, got none")
                return
            }
            if !strings.HasPrefix(authHeader, "Bearer ") {
                t.Errorf("Expected Authorization header to start with 'Bearer ', got '%s'", authHeader)
                return
            }
            encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
            decodedToken, err := base64.StdEncoding.DecodeString(encryptedToken)
            if err != nil {
                t.Errorf("Failed to base64 decode bearer token: %v", err)
                return
            }
            decryptedSessionKeyBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, decodedToken)
            if err != nil {
                t.Errorf("Failed to decrypt session key from token: %v", err)
                return
            }
            if string(decryptedSessionKeyBytes) != testSessionIDForSendError {
                t.Errorf("Decrypted session key '%s' does not match expected '%s'", string(decryptedSessionKeyBytes), testSessionIDForSendError)
            }

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest) // Or any other error status code
            json.NewEncoder(w).Encode(expectedErrPayload)
        }))
        defer finalReqServer.Close()

        client, clientErr := NewClient(testAPIKey, Sandbox, 30)
        if clientErr != nil {
            t.Fatalf("Failed to create client: %v", clientErr)
        }
        client.SessionKey = testSessionIDForSendError
        client.ExpiresAt = time.Now().Add(time.Hour * 1)
        // client.Keys.PublicKey will be the default from NewClient (e.g., SandboxPublicKey), which is what we want for createBearerToken(sessionID) to succeed.
        // actualGoodPubKeyBase64 was for genSessionKey's SendWithAuth. If genSessionKey is not run, this specific one isn't critical,
        // as long as the default client.Keys.PublicKey is valid for createBearerToken(sessionID).
        // For consistency and to ensure createBearerToken(sessionID) uses a known good key if default is not desired:
        client.Keys.PublicKey = actualGoodPubKeyBase64 // Ensure createBearerToken(sessionID) uses this known good key.

        // Configure redirectingTransport for the final API call to finalReqServer
        // Since genSessionKey is NOT called in this path (session key is pre-set),
        // client.Client.Transport is the original one (or http.DefaultTransport).
        finalCallTargetURL := client.makeUrl("/test/service_error_endpoint") // Must match the endpoint in NewRequest below
        parsedFinalCallURL, _ := url.Parse(finalCallTargetURL)
        parsedMockFinalReqURL, _ := url.Parse(finalReqServer.URL)

		client.Client.Transport = &redirectingTransport{
			defaultTransport:      http.DefaultTransport, // Or client.Client.Transport if it was set to something specific initially
			transportToMockServer: finalReqServer.Client().Transport,
			mockServerURL:         parsedMockFinalReqURL,
			targetScheme:          parsedFinalCallURL.Scheme,
			targetHost:            parsedFinalCallURL.Host,
			targetPath:            parsedFinalCallURL.Path, // This path must exactly match what makeUrl produces
		}

		req, _ := http.NewRequest("POST", client.makeUrl("/test/service_error_endpoint"), strings.NewReader("{}"))
		var successResult map[string]string // This should not be populated
		var actualErrorPayload mockAPIError   // This is where the error response should be unmarshalled

		err = client.SendWithSessionKey(req, &successResult, &actualErrorPayload)

		if err == nil {
			t.Fatal("Expected an error from SendWithSessionKey due to underlying Send failure, got nil")
		}

		if successResult != nil {
			t.Errorf("Expected successResult to be nil on error, got %v", successResult)
		}

		// The error string now includes the type of the unmarshalled payload if 'e' was provided.
		expectedErrStr := fmt.Sprintf("API error (status %d), payload unmarshalled into provided type: %v", http.StatusBadRequest, &expectedErrPayload)
		if err.Error() != expectedErrStr {
			t.Errorf("Expected error string '%s', got '%s'", expectedErrStr, err.Error())
		}

		// Check the unmarshalled error payload
		if actualErrorPayload.RequestID != expectedErrPayload.RequestID {
			t.Errorf("Expected error payload RequestID '%s', got '%s'", expectedErrPayload.RequestID, actualErrorPayload.RequestID)
		}
		if actualErrorPayload.ErrorCode != expectedErrPayload.ErrorCode {
			t.Errorf("Expected error payload ErrorCode '%s', got '%s'", expectedErrPayload.ErrorCode, actualErrorPayload.ErrorCode)
		}
		if actualErrorPayload.ErrorMessage != expectedErrPayload.ErrorMessage {
			t.Errorf("Expected error payload ErrorMessage '%s', got '%s'", expectedErrPayload.ErrorMessage, actualErrorPayload.ErrorMessage)
		}
	})
}
