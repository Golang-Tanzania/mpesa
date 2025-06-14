package mpesa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// queryTxStatusTestRedirectingTransport is a custom http.RoundTripper to redirect requests to the mock server.
// It also sets the Host header correctly for the mock server.
type queryTxStatusTestRedirectingTransport struct {
	targetURL *url.URL
	transport http.RoundTripper
}

func (t *queryTxStatusTestRedirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the request URL to point to the mock server
	req.URL.Scheme = t.targetURL.Scheme
	req.URL.Host = t.targetURL.Host
	req.Host = t.targetURL.Host // Also set the Host header

	// Use the underlying transport (e.g., http.DefaultTransport) to execute the request
	if t.transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.transport.RoundTrip(req)
}

// Helper to create a JSON response body for the mock server
func jsonResponseBodyTxStatus(t *testing.T, data interface{}) []byte {
	t.Helper()
	bytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON response: %v", err)
	}
	return bytes
}

// No longer needed, using MpesaError directly.
// type apiErrorTxStatus struct {
// 	ResponseCode string `json:"output_ResponseCode"`
// 	ResponseDesc string `json:"output_ResponseDesc"`
// }

func TestClient_QueryTxStatus(t *testing.T) {
	// Common session key response for mock server
	sessionKeyResp := SessionKeyResponse{
		OutputResponseCode: "0", // Corrected field name
		OutputResponseDesc: "Session key generated successfully", // Corrected field name
		OutputSessionID:    "testsessionkey",
	}

	tests := []struct {
		name             string
		payload          QueryTxStatusRequest
		mockStatus       int
		mockResponse     interface{}
		want             *QueryTxStatusResponse
		wantErr          bool
		wantErrPayload   *MpesaError // Changed to MpesaError
		expectedQuery    url.Values
	}{
		{
			name: "success",
			payload: QueryTxStatusRequest{
				QueryReference:           "someTxnID123",
				Country:                  "TZN",
				ServiceProviderCode:      "000000",
				ThirdPartyConversationID: "testconvid123",
			},
			mockStatus: http.StatusOK,
			mockResponse: QueryTxStatusResponse{
				ResponseCode:              "0",
				ResponseDesc:              "Transaction status fetched successfully",
				ResponseTransactionStatus: "Completed",
				ConversationID:            "serverConvID456",
				ThirdPartyConversationID:  "testconvid123",
				OriginalTransactionID:     "originalTxnID789",
			},
			want: &QueryTxStatusResponse{
				ResponseCode:              "0",
				ResponseDesc:              "Transaction status fetched successfully",
				ResponseTransactionStatus: "Completed",
				ConversationID:            "serverConvID456",
				ThirdPartyConversationID:  "testconvid123",
				OriginalTransactionID:     "originalTxnID789",
			},
			wantErr: false,
			expectedQuery: url.Values{
				"input_QueryReference":           []string{"someTxnID123"},
				"input_Country":                  []string{"TZN"},
				"input_ServiceProviderCode":      []string{"000000"},
				"input_ThirdPartyConversationID": []string{"testconvid123"},
			},
		},
		{
			name: "api error - transaction not found",
			payload: QueryTxStatusRequest{
				QueryReference:           "nonExistentTxnID456",
				Country:                  "TZN",
				ServiceProviderCode:      "000000",
				ThirdPartyConversationID: "testconvid456",
			},
			mockStatus: http.StatusBadRequest, // Or appropriate error code
			mockResponse: MpesaError{ // Changed to MpesaError
				ResponseCode: "INS-1", // Example error code
				ResponseDesc: "Transaction not found",
				HTTPStatusCode: http.StatusBadRequest, // Store the HTTP status for completeness
			},
			want:    nil,
			wantErr: true,
			wantErrPayload: &MpesaError{ // Changed to MpesaError
				ResponseCode: "INS-1",
				ResponseDesc: "Transaction not found",
				HTTPStatusCode: http.StatusBadRequest,
			},
			expectedQuery: url.Values{
				"input_QueryReference":           []string{"nonExistentTxnID456"},
				"input_Country":                  []string{"TZN"},
				"input_ServiceProviderCode":      []string{"000000"},
				"input_ThirdPartyConversationID": []string{"testconvid456"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handler logic
				t.Logf("[HANDLER START] For test: %s, Request: %s %s", tt.name, r.Method, r.URL.Path)

				// Determine expected full paths
				const testSandboxEndpoint = "/sandbox/ipg/v2/vodacomTZN/"
				const testSessionEndPath = "getSession"
				const testQueryTxStatusPath = "queryTransactionStatus"

				expectedSessionPath := testSandboxEndpoint + testSessionEndPath + "/"
				expectedAPIPath := testSandboxEndpoint + testQueryTxStatusPath + "/"
				t.Logf("[HANDLER CONFIG] Expected Session Path: '%s', Expected API Path: '%s'", expectedSessionPath, expectedAPIPath)

				if r.URL.Path == expectedSessionPath {
					t.Logf("[HANDLER MATCH] Matched session key path: '%s'", r.URL.Path)
					if r.Method != http.MethodGet {
						t.Errorf("SessionKey: Expected method GET, got %s", r.Method)
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					respBodyBytes := jsonResponseBodyTxStatus(t, sessionKeyResp)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(respBodyBytes)
					t.Logf("[HANDLER RESPONSE] SessionKey: Status=%d, Body=%s", http.StatusOK, string(respBodyBytes))
					return
				} else if r.URL.Path == expectedAPIPath {
					t.Logf("[HANDLER MATCH] Matched API call path: '%s'", r.URL.Path)
					if r.Method != http.MethodGet {
						t.Errorf("API Call: Expected method GET, got %s", r.Method)
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}

					// Check query parameters
					actualQuery := r.URL.Query()
					if !reflect.DeepEqual(actualQuery, tt.expectedQuery) {
						t.Errorf("API Call: Query parameters mismatch. Got %v, want %v", actualQuery, tt.expectedQuery)
					}

					respBodyBytes := jsonResponseBodyTxStatus(t, tt.mockResponse)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					w.Write(respBodyBytes)
					t.Logf("[HANDLER RESPONSE] API Call: Status=%d, Body=%s", tt.mockStatus, string(respBodyBytes))
					return
				}

				t.Logf("[HANDLER NO MATCH] Path: '%s' did not match expected session or API paths.", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
				io.WriteString(w, "Mock server: Not found")
			}))
			defer mockServer.Close()

			mockServerURL, _ := url.Parse(mockServer.URL)

			// Create client with custom transport to redirect to mock server
			client, err := NewClient("testapikey", Sandbox, 1)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			client.SetHttpClient(&http.Client{
				Transport: &queryTxStatusTestRedirectingTransport{targetURL: mockServerURL},
				Timeout:   10 * time.Second, // Add a timeout for tests
			})

			// Make the actual call
			got, err := client.QueryTxStatus(context.Background(), tt.payload)

			// Assertions
			if tt.wantErr {
				if err == nil {
					t.Errorf("QueryTxStatus() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Logf("Received expected error: %v", err)
				// Optionally, check the type and content of the error if it's a custom error type
				// or if it should contain specific information from the API error response.
				if tt.wantErrPayload != nil {
					apiErr, ok := err.(*MpesaError) // Changed to MpesaError
					if ok {
						// Compare relevant fields. HTTPStatusCode in tt.wantErrPayload might be set by the test setup,
						// while apiErr.HTTPStatusCode would be set by sendLocked if it's modified to do so.
						// For now, primarily check ResponseCode and ResponseDesc.
						if apiErr.ResponseCode != tt.wantErrPayload.ResponseCode || apiErr.ResponseDesc != tt.wantErrPayload.ResponseDesc {
							t.Errorf("QueryTxStatus() MpesaError fields mismatch. Got ResponseCode: '%s', ResponseDesc: '%s'. Want ResponseCode: '%s', ResponseDesc: '%s'", 
								apiErr.ResponseCode, apiErr.ResponseDesc, tt.wantErrPayload.ResponseCode, tt.wantErrPayload.ResponseDesc)
						}
					} else {
						// If it's not an ErrorResponse, check if the error message contains the expected description.
						// This is a fallback and might need adjustment based on how non-API errors are returned.
						if !strings.Contains(err.Error(), tt.wantErrPayload.ResponseDesc) {
							t.Errorf("QueryTxStatus() error message = %s, expected to contain %s", err.Error(), tt.wantErrPayload.ResponseDesc)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("QueryTxStatus() unexpected error = %v", err)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("QueryTxStatus() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
