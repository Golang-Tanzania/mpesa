package mpesa

import (
	"net/http"
)

const (
	Address             = "openapi.m-pesa.com"
	Ssl                 = true
	Port                = 443
	Sandbox             = "sandbox"
	ProdEndpoint        = "/openapi/ipg/v2/vodacomTZN/"
	SandboxEndpoint     = "/sandbox/ipg/v2/vodacomTZN/"
	SessionEndPath      = "getSession"
	C2BPaymentPath      = "c2bPayment/singleStage"
	B2BPaymentPath      = "b2bPayment"
	B2CPaymentPath      = "b2cPayment"
	ReversalPath        = "reversal"
	QueryTxStatusPath   = "queryTransactionStatus"
	DirectDebitPath     = "directDebitCreation"
	DebitDBPaymentPath  = "directDebitPayment"
	QueryBeneficialPath = "queryBeneficiaryName"
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

	B2CPaymentRequest struct {
		InputAmount                   string `json:"input_Amount"`
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN"`
		InputCountry                  string `json:"input_Country"`
		InputCurrency                 string `json:"input_Currency"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputTransactionReference     string `json:"input_TransactionReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputPaymentItemsDesc         string `json:"input_PaymentItemsDesc"`
	}

	B2CPaymentResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionID            string `json:"output_TransactionID"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	ReversalRequest struct {
		InputReversalAmount           string `json:"input_ReversalAmount,omitempty"`
		InputCountry                  string `json:"input_Country"`
		InputTransactionID            string `json:"input_TransactionID"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
	}

	ReversalResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionID            string `json:"output_TransactionID"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	QueryTxStatusRequest struct {
		InputQueryReference           string `json:"input_QueryReference"`
		InputCountry                  string `json:"input_Country"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
	}

	QueryTxStatusResponse struct {
		OutputResponseCode              string `json:"output_ResponseCode"`
		OutputResponseDesc              string `json:"output_ResponseDesc"`
		OutputResponseTransactionStatus string `json:"output_ResponseTransactionStatus"`
		OutputConversationID            string `json:"output_ConversationID"`
		OutputThirdPartyConversationID  string `json:"output_ThirdPartyConversationID"`
		OutputOriginalTransactionID     string `json:"output_OriginalTransactionID"`
	}

	DirectDebitRequest struct {
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN"`
		InputCountry                  string `json:"input_Country"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyReference      string `json:"input_ThirdPartyReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputAgreedTC                 string `json:"input_AgreedTC"`
		InputFirstPaymentDate         string `json:"input_FirstPaymentDate,omitempty"`
		InputFrequency                string `json:"input_Frequency,omitempty"`
		InputStartRangeOfDays         string `json:"input_StartRangeOfDays,omitempty"`
		InputEndRangeOfDays           string `json:"input_EndRangeOfDays,omitempty"`
		InputExpiryDate               string `json:"input_ExpiryDate,omitempty"`
	}

	DirectDebitResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionReference     string `json:"output_TransactionReference"`
		OutputMsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	DebitDBPaymentReq struct {
		InputMsisdnToken              string `json:"input_MsisdnToken,omitempty"`
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN,omitempty"`
		InputCountry                  string `json:"input_Country"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyReference      string `json:"input_ThirdPartyReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputAmount                   string `json:"input_Amount"`
		InputCurrency                 string `json:"input_Currency"`
		InputMandateID                string `json:"input_MandateID,omitempty"`
	}

	DebitDBPaymentRes struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionID            string `json:"output_TransactionID"`
		OutputMsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID,omitempty"`
	}

	QueryBenRequest struct {
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN"`
		InputCountry                  string `json:"input_Country"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputKycQueryType             string `json:"input_KycQueryType"`
	}

	QueryBenResponse struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputCustomerFirstName        string `json:"output_CustomerFirstName"`
		OutputCustomerLastName         string `json:"output_CustomerLastName"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}
)
