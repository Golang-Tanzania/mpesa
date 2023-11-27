/*
Copyright (c) 2022-2023 Golang Tanzania

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
	"net/http"
)

const (
	Address             = "openapi.m-pesa.com"
	Ssl                 = true
	Port                = 443
	Sandbox             = "sandbox"
	Production          = "production"
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
	QueryDirectDBPath   = "queryDirectDebit"
	CancelDirectDBPath  = "directDebitCancel"
	SandboxPublicKey    = "MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEArv9yxA69XQKBo24BaF/D+fvlqmGdYjqLQ5WtNBb5tquqGvAvG3WMFETVUSow/LizQalxj2ElMVrUmzu5mGGkxK08bWEXF7a1DEvtVJs6nppIlFJc2SnrU14AOrIrB28ogm58JjAl5BOQawOXD5dfSk7MaAA82pVHoIqEu0FxA8BOKU+RGTihRU+ptw1j4bsAJYiPbSX6i71gfPvwHPYamM0bfI4CmlsUUR3KvCG24rB6FNPcRBhM3jDuv8ae2kC33w9hEq8qNB55uw51vK7hyXoAa+U7IqP1y6nBdlN25gkxEA8yrsl1678cspeXr+3ciRyqoRgj9RD/ONbJhhxFvt1cLBh+qwK2eqISfBb06eRnNeC71oBokDm3zyCnkOtMDGl7IvnMfZfEPFCfg5QgJVk1msPpRvQxmEsrX9MQRyFVzgy2CWNIb7c+jPapyrNwoUbANlN8adU1m6yOuoX7F49x+OjiG2se0EJ6nafeKUXw/+hiJZvELUYgzKUtMAZVTNZfT8jjb58j8GVtuS+6TM2AutbejaCV84ZK58E2CRJqhmjQibEUO6KPdD7oTlEkFy52Y1uOOBXgYpqMzufNPmfdqqqSM4dU70PO8ogyKGiLAIxCetMjjm6FCMEA3Kc8K0Ig7/XtFm9By6VxTJK1Mg36TlHaZKP6VzVLXMtesJECAwEAAQ=="
	OpenapiPublicKey    = "MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAietPTdEyyoV/wvxRjS5pSn3ZBQH9hnVtQC9SFLgM9IkomEX9Vu9fBg2MzWSSqkQlaYIGFGH3d69Q5NOWkRo+Y8p5a61sc9hZ+ItAiEL9KIbZzhnMwi12jUYCTff0bVTsTGSNUePQ2V42sToOIKCeBpUtwWKhhW3CSpK7S1iJhS9H22/BT/pk21Jd8btwMLUHfVD95iXbHNM8u6vFaYuHczx966T7gpa9RGGXRtiOr3ScJq1515tzOSOsHTPHLTun59nxxJiEjKoI4Lb9h6IlauvcGAQHp5q6/2XmxuqZdGzh39uLac8tMSmY3vC3fiHYC3iMyTb7eXqATIhDUOf9mOSbgZMS19iiVZvz8igDl950IMcelJwcj0qCLoufLE5y8ud5WIw47OCVkD7tcAEPmVWlCQ744SIM5afw+Jg50T1SEtu3q3GiL0UQ6KTLDyDEt5BL9HWXAIXsjFdPDpX1jtxZavVQV+Jd7FXhuPQuDbh12liTROREdzatYWRnrhzeOJ5Se9xeXLvYSj8DmAI4iFf2cVtWCzj/02uK4+iIGXlX7lHP1W+tycLS7Pe2RdtC2+oz5RSSqb5jI4+3iEY/vZjSMBVk69pCDzZy4ZE8LBgyEvSabJ/cddwWmShcRS+21XvGQ1uXYLv0FCTEHHobCfmn2y8bJBb/Hct53BaojWUCAwEAAQ=="
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

	DirectDBCreateReq struct {
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

	DirectDBCreateRes struct {
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

	QueryDirectDBReq struct {
		InputQueryBalanceAmount       bool   `json:"input_QueryBalanceAmount"`
		InputBalanceAmount            string `json:"input_BalanceAmount,omitempty"`
		InputCountry                  string `json:"input_Country"`
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN,omitempty"`
		InputMsisdnToken              string `json:"input_MsisdnToken,omitempty"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputThirdPartyReference      string `json:"input_ThirdPartyReference"`
		InputMandateID                string `json:"input_MandateID,omitempty"`
		InputCurrency                 string `json:"input_Currency"`
	}
	QueryDirectDBRes struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionReference     string `json:"output_TransactionReference"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
		OutputSufficientBalance        bool   `json:"output_SufficientBalance"`
		OutputMsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		OutputMandateID                string `json:"output_MandateID"`
		OutputMandateStatus            string `json:"output_MandateStatus"`
		OutputAccountStatus            string `json:"output_AccountStatus"`
		OutputFirstPaymentDate         string `json:"output_FirstPaymentDate,omitempty"`
		OutputFrequency                string `json:"output_Frequency,omitempty"`
		OutputPaymentDayFrom           string `json:"output_PaymentDayFrom,omitempty"`
		OutputPaymentDayTo             string `json:"output_PaymentDayTo,omitempty"`
		OutputExpiryDate               string `json:"output_ExpiryDate,omitempty"`
	}

	CancelDirectDBReq struct {
		InputMsisdnToken              string `json:"input_MsisdnToken"`
		InputCustomerMSISDN           string `json:"input_CustomerMSISDN"`
		InputCountry                  string `json:"input_Country"`
		InputServiceProviderCode      string `json:"input_ServiceProviderCode"`
		InputThirdPartyReference      string `json:"input_ThirdPartyReference"`
		InputThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		InputMandateID                string `json:"input_MandateID"`
	}

	CancelDirectDBRes struct {
		OutputResponseCode             string `json:"output_ResponseCode"`
		OutputResponseDesc             string `json:"output_ResponseDesc"`
		OutputTransactionReference     string `json:"output_TransactionReference"`
		OutputMsisdnToken              string `json:"output_MsisdnToken"`
		OutputConversationID           string `json:"output_ConversationID"`
		OutputThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}
)
