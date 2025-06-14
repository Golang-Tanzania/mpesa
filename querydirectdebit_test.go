package mpesa

import (
	"bytes"
	"context"
	"encoding/json"
	// "fmt" // Removed as it might be unused after changes
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	// "github.com/stretchr/testify/assert" // Removed testify import
)

// queryTestRedirectingTransport is an http.RoundTripper that rewrites all requests
// to a specific target scheme and host, preserving the original path and query.
// It also sets the Host header correctly for the target, crucial for httptest.Server.
type queryTestRedirectingTransport struct {
	targetScheme string
	targetHost   string
	transport    http.RoundTripper // This is the underlying transport, e.g., http.DefaultTransport
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *queryTestRedirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	clonedReq := req.Clone(req.Context())

	// Preserve original path and query, but change scheme and host
	clonedReq.URL.Scheme = rt.targetScheme
	clonedReq.URL.Host = rt.targetHost

	// Set the Host header to match the target, which is crucial for httptest.Server
	clonedReq.Host = rt.targetHost

	// Proceed with the modified request using the underlying transport
	return rt.transport.RoundTrip(clonedReq)
}

// apiError is used to mock API error responses.
type apiError struct {
	ResponseCode string `json:"output_ResponseCode"`
	ResponseDesc string `json:"output_ResponseDesc"`
}

// jsonResponseBody is a helper to create an io.ReadCloser from a byte slice.
func jsonResponseBody(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(data))
}

func TestClient_QueryDirectDebit(t *testing.T) {
	client, _, _ := newTestClientWithKeys(t) // Corrected call

	ctx := context.Background()

	tests := []struct {
		name                 string
		payload              QueryDirectDBReq
		mockStatus           int
		mockResponse         interface{}
		want                 *QueryDirectDBRes
		wantErr              bool
		wantErrPayload       *apiError // For checking error response body
		expectedPath         string
		expectedQuery        url.Values
		sessionKeyPathSuffix string
		apiPathSuffix        string
	}{
		{
			name: "success",
			payload: QueryDirectDBReq{
				Country:                  "TZN",
				ServiceProviderCode:      "000000",
				ThirdPartyConversationID: "testconvid123",
				ThirdPartyReference:      "testref123",
				MandateID:                "MANDATE123",
				Currency:                 "TZS",
			},
			mockStatus: http.StatusOK,
			mockResponse: QueryDirectDBRes{
				ResponseCode:             "0",
				ResponseDesc:             "Success",
				TransactionReference:     "TXNREF123",
				ConversationID:           "APIConvID123",
				ThirdPartyConversationID: "testconvid123",
				SufficientBalance:        true,
				MandateStatus:            "Active",
				AccountStatus:            "Active",
				FirstPaymentDate:         "2023-01-01",
				Frequency:                "Monthly",
				PaymentDayFrom:           "1",
				PaymentDayTo:             "5",
				ExpiryDate:               "2024-01-01",
			},
			want: &QueryDirectDBRes{
				ResponseCode:             "0",
				ResponseDesc:             "Success",
				TransactionReference:     "TXNREF123",
				ConversationID:           "APIConvID123",
				ThirdPartyConversationID: "testconvid123",
				SufficientBalance:        true,
				MandateStatus:            "Active",
				AccountStatus:            "Active",
				FirstPaymentDate:         "2023-01-01",
				Frequency:                "Monthly",
				PaymentDayFrom:           "1",
				PaymentDayTo:             "5",
				ExpiryDate:               "2024-01-01",
			},
			wantErr:              false,
			wantErrPayload:       nil,
			// sessionKeyPathSuffix and apiPathSuffix are no longer used for path matching in handler
			// sessionKeyPathSuffix: SessionEndPath, 
			// apiPathSuffix:        QueryDirectDBPath,
			expectedQuery: url.Values{
				"input_Country":                  []string{"TZN"},
				"input_ServiceProviderCode":      []string{"000000"},
				"input_ThirdPartyConversationID": []string{"testconvid123"},
				"input_ThirdPartyReference":      []string{"testref123"},
				"input_MandateID":                []string{"MANDATE123"},
				"input_Currency":                 []string{"TZS"},
			},
		},
		{
			name: "api error - invalid mandate id",
			payload: QueryDirectDBReq{
				Country:                  "TZN",
				ServiceProviderCode:      "000000",
				ThirdPartyConversationID: "errconvid456",
				ThirdPartyReference:      "errref456",
				MandateID:                "INVALIDMANDATE456",
				Currency:                 "TZS",
			},
			mockStatus: http.StatusBadRequest, // Or appropriate error code
			mockResponse: apiError{
				ResponseCode: "INS-10", // Example error code
				ResponseDesc: "Invalid Mandate ID",
			},
			want:    nil,
			wantErr: true,
			wantErrPayload: &apiError{
				ResponseCode: "INS-10",
				ResponseDesc: "Invalid Mandate ID",
			},
			// sessionKeyPathSuffix and apiPathSuffix are no longer used for path matching in handler
			// sessionKeyPathSuffix: SessionEndPath,
			// apiPathSuffix:        QueryDirectDBPath,
			expectedQuery: url.Values{
				"input_Country":                  []string{"TZN"},
				"input_ServiceProviderCode":      []string{"000000"},
				"input_ThirdPartyConversationID": []string{"errconvid456"},
				"input_ThirdPartyReference":      []string{"errref456"},
				"input_MandateID":                []string{"INVALIDMANDATE456"},
				"input_Currency":                 []string{"TZS"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handler logic
				t.Logf("[HANDLER START] For test: %s, Request: %s %s", tt.name, r.Method, r.URL.Path)

				// Determine expected full paths based on client's environment (assuming Sandbox for tests)
				// Constants from types.go: SandboxEndpoint, SessionEndPath, QueryDirectDBPath
				const testSandboxEndpoint = "/sandbox/ipg/v2/vodacomTZN/"
				const testSessionEndPath = "getSession"
				const testQueryDirectDBPath = "queryDirectDebit"

				expectedSessionPath := testSandboxEndpoint + testSessionEndPath + "/"
				expectedAPIPath := testSandboxEndpoint + testQueryDirectDBPath + "/"
				t.Logf("[HANDLER CONFIG] Expected Session Path: '%s', Expected API Path: '%s'", expectedSessionPath, expectedAPIPath)

				if r.URL.Path == expectedSessionPath {
					t.Logf("[HANDLER MATCH] Matched session key path: '%s'", r.URL.Path)
					if r.Method != http.MethodGet {
						t.Errorf("SessionKey: Expected method GET, got %s", r.Method)
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					sessionResp := SessionKeyResponse{
						OutputSessionID:    "testsessionkey",
						OutputResponseCode: "0",
						OutputResponseDesc: "Session key generated successfully",
					}
					respBodyBytes, _ := json.Marshal(sessionResp)
					t.Logf("[HANDLER RESPONSE] SessionKey: Status=%d, Body=%s", http.StatusOK, string(respBodyBytes))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(respBodyBytes)
					return
				} else if r.URL.Path == expectedAPIPath {
					t.Logf("[HANDLER MATCH] Matched API call path: '%s'", r.URL.Path)
					if r.Method != http.MethodGet {
						t.Errorf("API Call: Expected method GET, got %s", r.Method)
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}

					actualQuery := r.URL.Query()
					for k, v := range tt.expectedQuery {
						if !reflect.DeepEqual(actualQuery[k], v) {
							t.Errorf("API Call: Expected query param %s=%v, got %s=%v", k, v, k, actualQuery[k])
						}
					}

					respBodyBytes, _ := json.Marshal(tt.mockResponse)
					t.Logf("[HANDLER RESPONSE] API Call: Status=%d, Body=%s", tt.mockStatus, string(respBodyBytes))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					w.Write(respBodyBytes)
					return
				} else {
					t.Logf("[HANDLER NO MATCH] Path '%s'. Sending 404.", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer mockServer.Close()

			// Create a redirecting transport
			mockServerTargetURL, _ := url.Parse(mockServer.URL)
			rdt := &queryTestRedirectingTransport{ // Renamed
				targetScheme: mockServerTargetURL.Scheme,
				targetHost:   mockServerTargetURL.Host,
				transport:    http.DefaultTransport, // Ensure this field matches the struct definition
			}

			originalHTTPClient := client.Client // Store the original client
			
			// Create a new http.Client that uses our queryTestRedirectingTransport
			mockRedirectingHttpClient := &http.Client{
				Transport: rdt,
			}
			client.SetHttpClient(mockRedirectingHttpClient) // Set this new client on our mpesa client

			// Restore original http client after test
			defer func() {
				client.SetHttpClient(originalHTTPClient)
			}()

			// var errPayload apiError // Removed as it was unused
			got, err := client.QueryDirectDebit(ctx, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Client.QueryDirectDebit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrPayload != nil {
					asMockErr, ok := tt.mockResponse.(apiError)
					if ok {
						// Check if the actual error string contains the expected response description
						// This is a basic check; more sophisticated error checking might involve unmarshalling the error response
						if !strings.Contains(err.Error(), asMockErr.ResponseDesc) {
							t.Errorf("Client.QueryDirectDebit() error string = %q, want to contain %q", err.Error(), asMockErr.ResponseDesc)
						}
						// Additionally, check if the error string contains the expected response code
						if !strings.Contains(err.Error(), asMockErr.ResponseCode) {
							t.Errorf("Client.QueryDirectDebit() error string = %q, want to contain %q", err.Error(), asMockErr.ResponseCode)
						}
					} else {
						t.Logf("Warning: mockResponse for error case is not of type apiError, cannot verify error description/code in error string precisely.")
					}
				}
			} else {
				// Success case: check if the response matches the expected one
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Client.QueryDirectDebit() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

