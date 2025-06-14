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

func TestDirectDebitCreate(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulDirectDebitCreate", func(t *testing.T) {
		const testSessionID = "test-session-id-ddcreate-success"
		expectedPayload := DirectDBCreateReq{
			CustomerMSISDN:           "255700000004",
			Country:                  "TZN",
			ServiceProviderCode:      "123456",
			ThirdPartyReference:      "ref-ddcreate-success",
			ThirdPartyConversationID: "conv-id-ddcreate-success",
			Frequency:                "06", // Monthly
			FirstPaymentDate:         "20250101",
			ExpiryDate:               "20260101",

		}

		sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		createMandateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, DirectDebitPath) {
				t.Errorf("Expected request to direct debit create path, got: %s", r.URL.Path)
			}

			response := DirectDBCreateRes{
				ResponseCode:         "0",
				ResponseDesc:         "Direct Debit Mandate created successfully.",
				TransactionReference: "ddcreate-txn-123",
				ConversationID:       "server-conv-id-ddcreate",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer createMandateServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		createURL, _ := url.Parse(client.makeUrl(DirectDebitPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		createMockURL, _ := url.Parse(createMandateServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				createURL.Path:  createMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.DirectDebitCreate(context.Background(), expectedPayload)
		if err != nil {
			t.Fatalf("DirectDebitCreate failed: %v", err)
		}

		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.TransactionReference != "ddcreate-txn-123" {
			t.Errorf("Expected TransactionReference 'ddcreate-txn-123', got '%s'", resp.TransactionReference)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionKeyResponse := struct {
				OutputSessionKey string `json:"output_SessionKey"`
				OutputSessionID  string `json:"output_SessionID"`
			}{
				OutputSessionKey: pubKeyBase64,
				OutputSessionID:  "error-case-session-id-ddcreate",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionKeyResponse)
		}))
		defer sessionServer.Close()

		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := `{"output_ResponseCode":"301","output_ResponseDesc":"Invalid customer MSISDN"}`
			w.Write([]byte(errorResponse))
		}))
		defer errorServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		createURL, _ := url.Parse(client.makeUrl(DirectDebitPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				createURL.Path:  errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := DirectDBCreateReq{CustomerMSISDN: "invalid-number"}
		resp, err := client.DirectDebitCreate(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from DirectDebitCreate, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "Invalid customer MSISDN") {
			t.Errorf("Expected error to contain 'Invalid customer MSISDN', got '%v'", err)
		}
	})
}
