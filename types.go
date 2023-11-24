package mpesa

import (
	"net/http"
)

const (
	Address         = "openapi.m-pesa.com"
	Ssl             = true
	Port            = 443
	Sandbox         = "sandbox"
	ProdEndpoint    = "/openapi/ipg/v2/vodacomTZN/"
	SandboxEndpoint = "/sandbox/ipg/v2/vodacomTZN/"
	SessionEndPath  = "getSession"
)

type (
	Client struct {
		Client      *http.Client
		Keys        *Keys
		Environment string
	}

	Keys struct {
		PublicKey string
		ApiKey    string
	}

	SessionIDResponse struct {
		OutputResponseCode string `json:"output_ResponseCode"`
		OutputResponseDesc string `json:"output_ResponseDesc"`
		OutputSessionID    string `json:"output_SessionID"`
	}
)
