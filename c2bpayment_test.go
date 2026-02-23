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

func TestC2BPayment(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulC2BPayment", func(t *testing.T) {
		client.SessionKey = ""
		client.ExpiresAt = time.Time{}
		const testSessionID = "test-session-id-c2b-success"
		expectedPayload := C2BPaymentRequest{
			Amount:                   "200",
			Country:                  "TZN",
			Currency:                 "TZS",
			CustomerMSISDN:           "255700000000",
			ServiceProviderCode:      "12345",
			ThirdPartyConversationID: "conv-id-c2b-success",
			TransactionReference:     "ref-c2b-success",
			PurchasedItemsDesc:       "Test Item",
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

		c2bServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, C2BPaymentPath) {
				t.Errorf("Expected request to C2B payment path, got: %s", r.URL.Path)
			}

			response := C2BPaymentResponse{
				ResponseCode:             "0",
				ResponseDesc:             "OK",
				TransactionID:            "c2b-txn-123",
				ThirdPartyConversationID: expectedPayload.ThirdPartyConversationID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer c2bServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		c2bURL, _ := url.Parse(client.makeUrl(C2BPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		c2bMockURL, _ := url.Parse(c2bServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				c2bURL.Path:     c2bMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.C2BPayment(context.Background(), expectedPayload)
		if err != nil {
			t.Fatalf("C2BPayment failed: %v", err)
		}

		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.TransactionID != "c2b-txn-123" {
			t.Errorf("Expected TransactionID 'c2b-txn-123', got '%s'", resp.TransactionID)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		client.SessionKey = ""
		client.ExpiresAt = time.Time{}
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
		c2bURL, _ := url.Parse(client.makeUrl(C2BPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				c2bURL.Path:     errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := C2BPaymentRequest{}
		resp, err := client.C2BPayment(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from C2BPayment, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "Invalid request") {
			t.Errorf("Expected error to contain 'Invalid request', got '%v'", err)
		}
	})
}
