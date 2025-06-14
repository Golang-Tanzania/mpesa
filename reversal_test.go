package mpesa_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

// reversalTestRedirectingTransport is a custom http.RoundTripper that redirects requests to a target URL.
// This is used to route all client HTTP requests to the mock server during tests.
type reversalTestRedirectingTransport struct {
	targetURL *url.URL
	transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface.
// It rewrites the request URL to point to the mock server and then uses the underlying transport.
func (t *reversalTestRedirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.targetURL.Scheme
	req.URL.Host = t.targetURL.Host
	req.Host = t.targetURL.Host // Also set the Host header for the request

	if t.transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.transport.RoundTrip(req)
}

// newTestReversalServer creates a mock HTTP server for reversal tests.
// It handles session key generation and the reversal API call.
func newTestReversalServer(t *testing.T, testName string, reversalReq mpesa.ReversalRequest, reversalResp interface{}, respCode int, sessionRespBody interface{}, sessionRespCode int) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s HANDLER START] For test: %s, Request: %s %s\n", strings.ToUpper(testName), testName, r.Method, r.URL.Path)

		expectedSessionPath := "/sandbox/ipg/v2/vodacomTZN/" + mpesa.SessionEndPath + "/"
		expectedAPIPath := "/sandbox/ipg/v2/vodacomTZN/" + mpesa.ReversalPath + "/"
		fmt.Printf("[%s HANDLER CONFIG] Expected Session Path: '%s', Expected API Path: '%s'\n", strings.ToUpper(testName), expectedSessionPath, expectedAPIPath)

		switch r.URL.Path {
		case expectedSessionPath:
			fmt.Printf("[%s HANDLER MATCH] Matched session key path: '%s'\n", strings.ToUpper(testName), r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(sessionRespCode)
			if err := json.NewEncoder(w).Encode(sessionRespBody); err != nil {
				t.Fatalf("[%s HANDLER ERROR] Failed to write session response: %v", strings.ToUpper(testName), err)
			}
			fmt.Printf("[%s HANDLER RESPONSE] SessionKey: Status=%d, Body=%+v\n", strings.ToUpper(testName), sessionRespCode, sessionRespBody)
		case expectedAPIPath:
			fmt.Printf("[%s HANDLER MATCH] Matched API call path: '%s'\n", strings.ToUpper(testName), r.URL.Path)
			if r.Method != http.MethodPut {
				t.Errorf("[%s HANDLER ERROR] Expected PUT request, got %s", strings.ToUpper(testName), r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			// TODO: Optionally validate request body if needed for specific tests

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(respCode)
			if err := json.NewEncoder(w).Encode(reversalResp); err != nil {
				t.Fatalf("[%s HANDLER ERROR] Failed to write API response: %v", strings.ToUpper(testName), err)
			}
			fmt.Printf("[%s HANDLER RESPONSE] API Call: Status=%d, Body=%+v\n", strings.ToUpper(testName), respCode, reversalResp)
		default:
			t.Errorf("[%s HANDLER ERROR] Unexpected path: %s", strings.ToUpper(testName), r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return server
}

func TestClient_Reversal(t *testing.T) {
	client, err := mpesa.NewClient("testapikey", mpesa.Sandbox, 30)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		client.SessionKey = "" // Ensure we fetch a new session key
		client.ExpiresAt = time.Time{}
		ctx := context.Background()
		payload := mpesa.ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_123",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_001",
		}

		expectedResp := mpesa.ReversalResponse{
			ResponseCode:             "0",
			ResponseDesc:             "Reversal successful",
			TransactionID:            "REVERSAL_TXN_ID_456",
			ConversationID:           "serverConvID_reversal_789",
			ThirdPartyConversationID: "testconvid_reversal_001",
		}

		sessionResp := mpesa.SessionKeyResponse{
			OutputResponseCode: "0",
			OutputResponseDesc: "Session key generated successfully",
			OutputSessionID:    "testsessionkey_reversal",
		}

		mockServer := newTestReversalServer(t, "success", payload, expectedResp, http.StatusOK, sessionResp, http.StatusOK)
		defer mockServer.Close()

		// Setup redirecting transport
		mockServerURL, _ := url.Parse(mockServer.URL)
		originalTransport := client.Client.Transport
		client.Client.Transport = &reversalTestRedirectingTransport{targetURL: mockServerURL, transport: http.DefaultTransport}
		defer func() {
			client.Client.Transport = originalTransport
		}()

		resp, err := client.Reversal(ctx, payload)

		if err != nil {
			t.Fatalf("Reversal() error = %v, want nil", err)
		}

		if resp.ResponseCode != expectedResp.ResponseCode {
			t.Errorf("Reversal() ResponseCode = %s, want %s", resp.ResponseCode, expectedResp.ResponseCode)
		}
		if resp.ResponseDesc != expectedResp.ResponseDesc {
			t.Errorf("Reversal() ResponseDesc = %s, want %s", resp.ResponseDesc, expectedResp.ResponseDesc)
		}
		if resp.TransactionID != expectedResp.TransactionID {
			t.Errorf("Reversal() TransactionID = %s, want %s", resp.TransactionID, expectedResp.TransactionID)
		}
		if resp.ConversationID != expectedResp.ConversationID {
			t.Errorf("Reversal() ConversationID = %s, want %s", resp.ConversationID, expectedResp.ConversationID)
		}
		if resp.ThirdPartyConversationID != expectedResp.ThirdPartyConversationID {
			t.Errorf("Reversal() ThirdPartyConversationID = %s, want %s", resp.ThirdPartyConversationID, expectedResp.ThirdPartyConversationID)
		}
	})

	// TODO: Add test case for API error (e.g., transaction not found, invalid amount)
	// TODO: Add test case for session key generation failure

	t.Run("api_error", func(t *testing.T) {
		client.SessionKey = "" // Ensure we fetch a new session key
		client.ExpiresAt = time.Time{}
		ctx := context.Background()
		payload := mpesa.ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_UNKNOWN",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_err_001",
		}

		expectedAPIErr := mpesa.MpesaError{
			ResponseCode:             "INS-10", // Example error code for transaction not found
			ResponseDesc:             "Original transaction not found or already reversed.",
			HTTPStatusCode:           http.StatusBadRequest, // Store the HTTP status for verification
		}

		sessionResp := mpesa.SessionKeyResponse{
			OutputResponseCode: "0",
			OutputResponseDesc: "Session key generated successfully",
			OutputSessionID:    "testsessionkey_reversal_err",
		}

		// For API error test, reversalResp is the MpesaError struct
		mockServer := newTestReversalServer(t, "api_error", payload, expectedAPIErr, http.StatusBadRequest, sessionResp, http.StatusOK)
		defer mockServer.Close()

		// Setup redirecting transport
		mockServerURL, _ := url.Parse(mockServer.URL)
		originalTransport := client.Client.Transport
		client.Client.Transport = &reversalTestRedirectingTransport{targetURL: mockServerURL, transport: http.DefaultTransport}
		defer func() {
			client.Client.Transport = originalTransport
		}()

		resp, err := client.Reversal(ctx, payload)

		if err == nil {
			t.Fatalf("Reversal() error = nil, want an MpesaError")
		}

		if resp != nil {
			t.Errorf("Reversal() response = %+v, want nil on API error", resp)
		}

		mpesaErr, ok := err.(*mpesa.MpesaError)
		if !ok {
			t.Fatalf("Reversal() error type = %T, want *mpesa.MpesaError", err)
		}

		if mpesaErr.ResponseCode != expectedAPIErr.ResponseCode {
			t.Errorf("MpesaError ResponseCode = %s, want %s", mpesaErr.ResponseCode, expectedAPIErr.ResponseCode)
		}
		if mpesaErr.ResponseDesc != expectedAPIErr.ResponseDesc {
			t.Errorf("MpesaError ResponseDesc = %s, want %s", mpesaErr.ResponseDesc, expectedAPIErr.ResponseDesc)
		}
		if mpesaErr.HTTPStatusCode != expectedAPIErr.HTTPStatusCode {
			t.Errorf("MpesaError HTTPStatusCode = %d, want %d", mpesaErr.HTTPStatusCode, expectedAPIErr.HTTPStatusCode)
		}
	})

	t.Run("session_key_error", func(t *testing.T) {
		client.SessionKey = "" // Ensure we fetch a new session key (which should fail)
		client.ExpiresAt = time.Time{}
		ctx := context.Background()
		payload := mpesa.ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_SESSION_FAIL",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_session_err_001",
		}

		// This is the error returned by the session key endpoint
		expectedSessionErr := mpesa.MpesaError{
			ResponseCode:    "INS-5", // Example error for auth failure
			ResponseDesc:    "Authentication failed for session key generation",
			HTTPStatusCode:  http.StatusUnauthorized,
		}

		// The actual API call for reversal should not happen, so these don't strictly matter
		// but we provide empty/default values for the mock server setup.
		var ignoredAPIResp mpesa.ReversalResponse

		mockServer := newTestReversalServer(t, "session_key_error", payload, ignoredAPIResp, http.StatusOK, expectedSessionErr, http.StatusUnauthorized)
		defer mockServer.Close()

		// Setup redirecting transport
		mockServerURL, _ := url.Parse(mockServer.URL)
		originalTransport := client.Client.Transport
		client.Client.Transport = &reversalTestRedirectingTransport{targetURL: mockServerURL, transport: http.DefaultTransport}
		defer func() {
			client.Client.Transport = originalTransport
		}()

		resp, err := client.Reversal(ctx, payload)

		if err == nil {
			t.Fatalf("Reversal() error = nil, want an MpesaError from session key failure")
		}

		if resp != nil {
			t.Errorf("Reversal() response = %+v, want nil on session key error", resp)
		}

		mpesaErr, ok := err.(*mpesa.MpesaError)
		if !ok {
			t.Fatalf("Reversal() error type = %T, want *mpesa.MpesaError for session key failure", err)
		}

		if mpesaErr.ResponseCode != expectedSessionErr.ResponseCode {
			t.Errorf("MpesaError ResponseCode = %s, want %s", mpesaErr.ResponseCode, expectedSessionErr.ResponseCode)
		}
		if mpesaErr.ResponseDesc != expectedSessionErr.ResponseDesc {
			t.Errorf("MpesaError ResponseDesc = %s, want %s", mpesaErr.ResponseDesc, expectedSessionErr.ResponseDesc)
		}
		if mpesaErr.HTTPStatusCode != expectedSessionErr.HTTPStatusCode {
			t.Errorf("MpesaError HTTPStatusCode = %d, want %d", mpesaErr.HTTPStatusCode, expectedSessionErr.HTTPStatusCode)
		}
	})
}
