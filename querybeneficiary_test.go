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

func TestQueryBeneficiaryName(t *testing.T) {
	client, pubKeyBase64, _ := newTestClientWithKeys(t)

	t.Run("SuccessfulQueryBeneficiaryName", func(t *testing.T) {
		client.SessionKey = ""
		client.ExpiresAt = time.Time{}
		const testSessionID = "test-session-id-queryben-success"
		expectedPayload := QueryBenRequest{
			CustomerMSISDN:           "255700000005",
			Country:                  "TZN",
			ServiceProviderCode:      "123456",
			ThirdPartyConversationID: "conv-id-queryben-success",
			KycQueryType:             "02", // Basic KYC
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

		queryBenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, QueryBeneficialPath) {
				t.Errorf("Expected request to query beneficiary path, got: %s", r.URL.Path)
			}

			// Check query parameters (optional, but good practice)
			queryParams := r.URL.Query()
			if queryParams.Get("input_CustomerMSISDN") != expectedPayload.CustomerMSISDN {
				t.Errorf("Expected CustomerMSISDN '%s', got '%s'", expectedPayload.CustomerMSISDN, queryParams.Get("input_CustomerMSISDN"))
			}

			response := QueryBenResponse{
				ResponseCode:             "0",
				ResponseDesc:             "Beneficiary Found",
				CustomerFirstName:        "John",
				CustomerLastName:         "Doe",
				ConversationID:           "server-conv-id-queryben",
				ThirdPartyConversationID: expectedPayload.ThirdPartyConversationID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer queryBenServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		queryBenURL, _ := url.Parse(client.makeUrl(QueryBeneficialPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		queryBenMockURL, _ := url.Parse(queryBenServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path:  sessionMockURL,
				queryBenURL.Path: queryBenMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		resp, err := client.QueryBeneficiaryName(context.Background(), expectedPayload)
		if err != nil {
			t.Fatalf("QueryBeneficiaryName failed: %v", err)
		}

		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.CustomerFirstName != "John" {
			t.Errorf("Expected CustomerFirstName 'John', got '%s'", resp.CustomerFirstName)
		}
		if resp.CustomerLastName != "Doe" {
			t.Errorf("Expected CustomerLastName 'Doe', got '%s'", resp.CustomerLastName)
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
				OutputSessionID:  "error-case-session-id-queryben",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionKeyResponse)
		}))
		defer sessionServer.Close()

		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			errorResponse := `{"output_ResponseCode":"404","output_ResponseDesc":"Beneficiary not found"}`
			w.Write([]byte(errorResponse))
		}))
		defer errorServer.Close()

		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		queryBenURL, _ := url.Parse(client.makeUrl(QueryBeneficialPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path:  sessionMockURL,
				queryBenURL.Path: errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		payload := QueryBenRequest{CustomerMSISDN: "non-existent-number"}
		resp, err := client.QueryBeneficiaryName(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from QueryBeneficiaryName, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "Beneficiary not found") {
			t.Errorf("Expected error to contain 'Beneficiary not found', got '%v'", err)
		}
	})
}
