/*
Copyright (c) 2022-2025 Golang Tanzania

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mpesa

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestDirectDebitPayment(t *testing.T) {
	client, pubKeyBase64, privKey := newTestClientWithKeys(t)

	t.Run("SuccessfulDirectDebitPayment", func(t *testing.T) {
		client.SessionKey = ""
		client.ExpiresAt = time.Time{}
		// Test data
		const testSessionID = "test-session-id-12345"
		expectedPayload := DebitDBPaymentReq{
			MsisdnToken:              "AbCd123=",
			CustomerMSISDN:           "000000000001",
			Country:                  "TZN",
			ServiceProviderCode:      "000000",
			ThirdPartyReference:      "5db410b459bd433ca8e5",
			ThirdPartyConversationID: "AAA6d1f939c1005v2de053v4912jbasdj1j2kk",
			Amount:                   "10",
			Currency:                 "TZS",
			MandateID:                "15045",
		}

		// Mock session server for genSessionKey call
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

		// Mock direct debit payment server
		debitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify path
			if !strings.Contains(r.URL.Path, DebitDBPaymentPath) {
				t.Errorf("Expected request to direct debit payment path, got: %s", r.URL.Path)
				return
			}

			// Verify method
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
				return
			}

			// Verify Authorization header
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
			if string(decryptedSessionKeyBytes) != testSessionID {
				t.Errorf("Decrypted session key '%s' does not match expected '%s'", string(decryptedSessionKeyBytes), testSessionID)
			}

			// Send successful response
			response := DebitDBPaymentRes{
				ResponseCode:             "0",
				ResponseDesc:             "OK",
				TransactionID:            "123456789",
				ConversationID:           "AG_20230101_12345abc",
				ThirdPartyConversationID: expectedPayload.ThirdPartyConversationID,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer debitServer.Close()

		// Create client with the redirector

		// Set up the redirector for both session key and direct debit requests
		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		debitURL, _ := url.Parse(client.makeUrl(DebitDBPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		debitMockURL, _ := url.Parse(debitServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				debitURL.Path:   debitMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		// Make the direct debit payment request
		resp, err := client.DirectDebitPayment(context.Background(), expectedPayload)

		// Verify results
		if err != nil {
			t.Fatalf("DirectDebitPayment returned an error: %v", err)
		}
		if resp == nil {
			t.Fatalf("DirectDebitPayment returned nil response")
		}
		if resp.ResponseCode != "0" {
			t.Errorf("Expected ResponseCode '0', got '%s'", resp.ResponseCode)
		}
		if resp.TransactionID != "123456789" {
			t.Errorf("Expected TransactionID '123456789', got '%s'", resp.TransactionID)
		}
		if resp.ThirdPartyConversationID != expectedPayload.ThirdPartyConversationID {
			t.Errorf("Expected ThirdPartyConversationID '%s', got '%s'",
				expectedPayload.ThirdPartyConversationID, resp.ThirdPartyConversationID)
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		client.SessionKey = ""
		client.ExpiresAt = time.Time{}
		// Mock session server for genSessionKey call
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

		// Mock server that returns an error
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := `{"output_ResponseCode":"100","output_ResponseDesc":"Invalid mandate ID"}`
			w.Write([]byte(errorResponse))
		}))
		defer errorServer.Close()

		// Create client with the redirector

		// Set up the redirector
		sessionURL, _ := url.Parse(client.makeUrl(SessionEndPath))
		debitURL, _ := url.Parse(client.makeUrl(DebitDBPaymentPath))
		sessionMockURL, _ := url.Parse(sessionServer.URL)
		errorMockURL, _ := url.Parse(errorServer.URL)

		client.Client.Transport = &multiRedirectingTransport{
			redirects: map[string]*url.URL{
				sessionURL.Path: sessionMockURL,
				debitURL.Path:   errorMockURL,
			},
			defaultTransport: http.DefaultTransport,
		}

		// Test payload
		payload := DebitDBPaymentReq{
			MsisdnToken:              "AbCd123=",
			CustomerMSISDN:           "000000000001",
			Country:                  "TZN",
			ServiceProviderCode:      "000000",
			ThirdPartyReference:      "invalid-ref",
			ThirdPartyConversationID: "invalid-conv-id",
			Amount:                   "10",
			Currency:                 "TZS",
			MandateID:                "invalid-mandate",
		}

		// Make the direct debit payment request
		resp, err := client.DirectDebitPayment(context.Background(), payload)

		if err == nil {
			t.Fatal("Expected error from DirectDebitPayment, got nil")
		}
		if resp != nil {
			t.Fatalf("Expected nil response, got %+v", resp)
		}

		if !strings.Contains(err.Error(), "Invalid mandate ID") {
			t.Errorf("Expected error to contain 'Invalid mandate ID', got %v", err)
		}
	})
}
