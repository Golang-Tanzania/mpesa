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
	C2BPaymentPath  = "c2bPayment/singleStage"
	B2BPaymentPath  = "b2bPayment"
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

	SessionKeyResponse struct {
		OutputResponseCode string `json:"output_ResponseCode"`
		OutputResponseDesc string `json:"output_ResponseDesc"`
		OutputSessionID    string `json:"output_SessionID"`
	}

	C2BPaymentRequest struct {
		InputAmount                   string `json:"input_Amount"`
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN"`
		InputCountry                  string `json:"input_Country"`
		InputCurrency                 string `json:"input_Currency"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputTransactionReference     string `json:"input_TransactionReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputPurchasedItemsDesc       string `json:"input_PurchasedItemsDesc"`
	}

	C2BPaymentResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionID            string `json:"output_TransactionID"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	B2BPaymentRequest struct {
		InputAmount                   string `json:"input_Amount"`
		InputReceiverPartyCode        string `json:"input_ReceiverPartyCode"`
		InputCountry                  string `json:"input_Country"`
		InputCurrency                 string `json:"input_Currency"`
		InputPrimaryPartyCode         string `json:"input_PrimaryPartyCode"`
		InputTransactionReference     string `json:"input_TransactionReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputPurchasedItemsDesc       string `json:"input_PurchasedItemsDesc"`
	}
	B2BPaymentResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionID            string `json:"output_TransactionID"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}
)
