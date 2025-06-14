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

func TestB2CPayment(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulB2CPayment", func(t *testing.T) {
		const testSessionID = "test-session-id-b2c-success"
		expectedPayload := B2CPaymentRequest{
			Amount:                   "300",
			Country:                  "TZN",
			Currency:                 "TZS",
			CustomerMSISDN:           "255700000002",
			ServiceProviderCode:      "123456",
			TransactionReference:     "ref-b2c-success",
			ThirdPartyConversationID: "conv-id-b2c-success",
			PaymentItemsDesc:         "Test B2C Item",
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

		b2cServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, B2CPaymentPath) {
				t.Errorf("Expected request to B2C payment path, got: %s", r.URL.Path)
			}

			response := B2CPaymentResponse{
				ResponseCode:             "0",
				ResponseDesc:             "OK",
				TransactionID:            "b2c-txn-123",
				ThirdPartyConversationID: expectedPayload.ThirdPartyConversationID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer b2cServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		b2cURL, _ := url.Parse(client.makeUrl(B2CPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		b2cMockURL, _ := url.Parse(b2cServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				b2cURL.Path:     b2cMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.B2CPayment(context.Background(), expectedPayload)
		if err != nil {
			t.Fatalf("B2CPayment failed: %v", err)
		}

		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.TransactionID != "b2c-txn-123" {
			t.Errorf("Expected TransactionID 'b2c-txn-123', got '%s'", resp.TransactionID)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionKeyResponse := struct {
				OutputSessionKey string `json:"output_SessionKey"`
				OutputSessionID  string `json:"output_SessionID"`
			}{
				OutputSessionKey: pubKeyBase64,
				OutputSessionID:  "error-case-session-id",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionKeyResponse)
		}))
		defer sessionServer.Close()

		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := `{"output_ResponseCode":"101","output_ResponseDesc":"Invalid request"}`
			w.Write([]byte(errorResponse))
		}))
		defer errorServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		b2cURL, _ := url.Parse(client.makeUrl(B2CPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				b2cURL.Path:     errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := B2CPaymentRequest{}
		resp, err := client.B2CPayment(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from B2CPayment, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "Invalid request") {
			t.Errorf("Expected error to contain 'Invalid request', got '%v'", err)
		}
	})
}
