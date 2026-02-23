package mpesa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newTestReversalServer creates a mock HTTP server for reversal tests.
// It handles session key generation and the reversal API call.
func newTestReversalServer(t *testing.T, testName string, reversalReq ReversalRequest, reversalResp interface{}, respCode int, sessionRespBody interface{}, sessionRespCode int) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("[%s HANDLER START] For test: %s, Request: %s %s", strings.ToUpper(testName), testName, r.Method, r.URL.Path)

		expectedSessionPath := "/sandbox/ipg/v2/vodacomTZN/" + SessionEndPath + "/"
		expectedAPIPath := "/sandbox/ipg/v2/vodacomTZN/" + ReversalPath + "/"
		t.Logf("[%s HANDLER CONFIG] Expected Session Path: '%s', Expected API Path: '%s'", strings.ToUpper(testName), expectedSessionPath, expectedAPIPath)

		switch r.URL.Path {
		case expectedSessionPath:
			t.Logf("[%s HANDLER MATCH] Matched session key path: '%s'", strings.ToUpper(testName), r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(sessionRespCode)
			if err := json.NewEncoder(w).Encode(sessionRespBody); err != nil {
				t.Fatalf("[%s HANDLER ERROR] Failed to write session response: %v", strings.ToUpper(testName), err)
			}
			t.Logf("[%s HANDLER RESPONSE] SessionKey: Status=%d, Body=%+v", strings.ToUpper(testName), sessionRespCode, sessionRespBody)
		case expectedAPIPath:
			t.Logf("[%s HANDLER MATCH] Matched API call path: '%s'", strings.ToUpper(testName), r.URL.Path)
			if r.Method != http.MethodPut {
				t.Errorf("[%s HANDLER ERROR] Expected PUT request, got %s", strings.ToUpper(testName), r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(respCode)
			if err := json.NewEncoder(w).Encode(reversalResp); err != nil {
				t.Fatalf("[%s HANDLER ERROR] Failed to write API response: %v", strings.ToUpper(testName), err)
			}
			t.Logf("[%s HANDLER RESPONSE] API Call: Status=%d, Body=%+v", strings.ToUpper(testName), respCode, reversalResp)
		default:
			t.Errorf("[%s HANDLER ERROR] Unexpected path: %s", strings.ToUpper(testName), r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return server
}

func TestClient_Reversal(t *testing.T) {
	client, _, _ := newTestClientWithKeys(t)

	t.Run("success", func(t *testing.T) {
		client.SessionKey = "" // Ensure we fetch a new session key
		client.ExpiresAt = time.Time{}
		ctx := context.Background()
		payload := ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_123",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_001",
		}

		expectedResp := ReversalResponse{
			ResponseCode:             "0",
			ResponseDesc:             "Reversal successful",
			TransactionID:            "REVERSAL_TXN_ID_456",
			ConversationID:           "serverConvID_reversal_789",
			ThirdPartyConversationID: "testconvid_reversal_001",
		}

		sessionResp := SessionKeyResponse{
			OutputResponseCode: "0",
			OutputResponseDesc: "Session key generated successfully",
			OutputSessionID:    "testsessionkey_reversal",
		}

		mockServer := newTestReversalServer(t, "success", payload, expectedResp, http.StatusOK, sessionResp, http.StatusOK)
		defer mockServer.Close()

		// Setup redirecting transport using multiRedirectingTransport
		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		reversalURL, _ := url.Parse(client.makeUrl(ReversalPath))
		mockServerURL, _ := url.Parse(mockServer.URL)

		originalTransport := client.Client.Transport
		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path:  mockServerURL,
				reversalURL.Path: mockServerURL,
			},
			defaultTransport: http.DefaultTransport,
		}
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

	t.Run("api_error", func(t *testing.T) {
		client.SessionKey = "" // Ensure we fetch a new session key
		client.ExpiresAt = time.Time{}
		ctx := context.Background()
		payload := ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_UNKNOWN",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_err_001",
		}

		expectedAPIErr := MpesaError{
			ResponseCode:   "INS-10",
			ResponseDesc:   "Original transaction not found or already reversed.",
			HTTPStatusCode: http.StatusBadRequest,
		}

		sessionResp := SessionKeyResponse{
			OutputResponseCode: "0",
			OutputResponseDesc: "Session key generated successfully",
			OutputSessionID:    "testsessionkey_reversal_err",
		}

		mockServer := newTestReversalServer(t, "api_error", payload, expectedAPIErr, http.StatusBadRequest, sessionResp, http.StatusOK)
		defer mockServer.Close()

		// Setup redirecting transport using multiRedirectingTransport
		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		reversalURL, _ := url.Parse(client.makeUrl(ReversalPath))
		mockServerURL, _ := url.Parse(mockServer.URL)

		originalTransport := client.Client.Transport
		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path:  mockServerURL,
				reversalURL.Path: mockServerURL,
			},
			defaultTransport: http.DefaultTransport,
		}
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

		mpesaErr, ok := err.(*MpesaError)
		if !ok {
			t.Fatalf("Reversal() error type = %T, want *MpesaError", err)
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
		payload := ReversalRequest{
			ReversalAmount:           "100",
			Country:                  "TZN",
			TransactionID:            "ORIGINAL_TXN_ID_SESSION_FAIL",
			ServiceProviderCode:      "171717",
			ThirdPartyConversationID: "testconvid_reversal_session_err_001",
		}

		expectedSessionErr := MpesaError{
			ResponseCode:   "INS-5",
			ResponseDesc:   "Authentication failed for session key generation",
			HTTPStatusCode: http.StatusUnauthorized,
		}

		// The actual API call for reversal should not happen, so these don't strictly matter
		// but we provide empty/default values for the mock server setup.
		var ignoredAPIResp ReversalResponse

		mockServer := newTestReversalServer(t, "session_key_error", payload, ignoredAPIResp, http.StatusOK, expectedSessionErr, http.StatusUnauthorized)
		defer mockServer.Close()

		// Setup redirecting transport using multiRedirectingTransport
		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		reversalURL, _ := url.Parse(client.makeUrl(ReversalPath))
		mockServerURL, _ := url.Parse(mockServer.URL)

		originalTransport := client.Client.Transport
		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path:  mockServerURL,
				reversalURL.Path: mockServerURL,
			},
			defaultTransport: http.DefaultTransport,
		}
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

		mpesaErr, ok := err.(*MpesaError)
		if !ok {
			t.Fatalf("Reversal() error type = %T, want *MpesaError for session key failure", err)
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
