package mpesa

import (
	"net/http"
)

const (
	Address = "openapi.m-pesa.com"
	Ssl     = true
	Port    = 443
	Sandbox = "sandbox"
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
)
