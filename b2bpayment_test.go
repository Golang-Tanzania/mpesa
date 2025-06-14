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


func TestB2BPayment(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulB2BPayment", func(t *testing.T) {
		const testSessionID = "test-session-id-b2b-success"
		expectedPayload := B2BPaymentRequest{
			Amount:                   "100",
			Country:                  "TZN",
			Currency:                 "TZS",
			PrimaryPartyCode:         "12345",
			ReceiverPartyCode:        "67890",
			ThirdPartyConversationID: "conv-id-success",
			TransactionReference:     "ref-success",
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

		b2bServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, B2BPaymentPath) {
				t.Errorf("Expected request to b2b payment path, got: %s", r.URL.Path)
				return
			}

			response := B2BPaymentResponse{
				ResponseCode:        "0",
				ResponseDesc: "OK",
				TransactionID:       "b2b-txn-123",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer b2bServer.Close()

		

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		b2bURL, _ := url.Parse(client.makeUrl(B2BPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		b2bMockURL, _ := url.Parse(b2bServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				b2bURL.Path:     b2bMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.B2BPayment(context.Background(), expectedPayload)

		if err != nil {
			t.Fatalf("B2BPayment returned an error: %v", err)
		}
		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
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
		b2bURL, _ := url.Parse(client.makeUrl(B2BPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				b2bURL.Path:     errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := B2BPaymentRequest{}
		resp, err := client.B2BPayment(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from B2BPayment, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
	})
}
