package mpesa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCancelDirectDebit(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulCancelDirectDebit", func(t *testing.T) {
		const testSessionID = "test-session-id-canceldd-success"
		expectedPayload := CancelDirectDBReq{
			MsisdnToken:              "msisdn-token-canceldd",
			CustomerMSISDN:           "255700000003",
			Country:                  "TZN",
			ServiceProviderCode:      "123456",
			ThirdPartyReference:      "ref-canceldd-success",
			ThirdPartyConversationID: "conv-id-canceldd-success",
			MandateID:                "mandate-id-canceldd",
		}

		sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, SessionEndPath) {
				t.Errorf("Expected request to session key path, got: %s", r.URL.Path)
			}
			sessionKeyResponse := struct {
				OutputSessionKey string `json:"output_SessionKey"`
				OutputSessionID  string `json:"output_SessionID"`
			}{
				OutputSessionKey: pubKeyBase64,
				OutputSessionID:  testSessionID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionKeyResponse)
		}))
		defer sessionServer.Close()

		cancelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, CancelDirectDBPath) {
				t.Errorf("Expected request to cancel direct debit path, got: %s", r.URL.Path)
			}

			var receivedPayload CancelDirectDBReq
			if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}
			if receivedPayload.MandateID != expectedPayload.MandateID {
				t.Errorf("Expected MandateID '%s', got '%s'", expectedPayload.MandateID, receivedPayload.MandateID)
			}

			response := CancelDirectDBRes{
				ResponseCode:             "0",
				ResponseDesc:             "Successfully cancelled direct debit mandate.",
				TransactionReference:     "canceldd-txn-123",
				MsisdnToken:              expectedPayload.MsisdnToken,
				ConversationID:           "server-conv-id-canceldd",
				ThirdPartyConversationID: expectedPayload.ThirdPartyConversationID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer cancelServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		cancelURL, _ := url.Parse(client.makeUrl(CancelDirectDBPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		cancelMockURL, _ := url.Parse(cancelServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				cancelURL.Path:  cancelMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.CancelDirectDebit(context.Background(), expectedPayload)
		if err != nil {
			t.Fatalf("CancelDirectDebit failed: %v", err)
		}

		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.TransactionReference != "canceldd-txn-123" {
			t.Errorf("Expected TransactionReference 'canceldd-txn-123', got '%s'", resp.TransactionReference)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionKeyResponse := struct {
				OutputSessionKey string `json:"output_SessionKey"`
				OutputSessionID  string `json:"output_SessionID"`
			}{
				OutputSessionKey: pubKeyBase64,
				OutputSessionID:  "error-case-session-id-canceldd",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionKeyResponse)
		}))
		defer sessionServer.Close()

		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := `{"output_ResponseCode":"201","output_ResponseDesc":"Invalid mandate ID"}`
			w.Write([]byte(errorResponse))
		}))
		defer errorServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		cancelURL, _ := url.Parse(client.makeUrl(CancelDirectDBPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				cancelURL.Path:  errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := CancelDirectDBReq{MandateID: "invalid-mandate"}
		resp, err := client.CancelDirectDebit(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from CancelDirectDebit, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "Invalid mandate ID") {
			t.Errorf("Expected error to contain 'Invalid mandate ID', got '%v'", err)
		}
	})
}
