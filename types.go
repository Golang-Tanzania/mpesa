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
	"sync"
	"time"
)

const (
	Address                             = "openapi.m-pesa.com"
	Ssl                                 = true
	Port                                = 443
	Sandbox                             = "sandbox"
	Production                          = "production"
	ProdEndpoint                        = "/openapi/ipg/v2/vodacomTZN/"
	SandboxEndpoint                     = "/sandbox/ipg/v2/vodacomTZN/"
	SessionEndPath                      = "getSession"
	C2BPaymentPath                      = "c2bPayment/singleStage"
	B2BPaymentPath                      = "b2bPayment"
	B2CPaymentPath                      = "b2cPayment"
	ReversalPath                        = "reversal"
	QueryTxStatusPath                   = "queryTransactionStatus"
	DirectDebitPath                     = "directDebitCreation"
	DebitDBPaymentPath                  = "directDebitPayment"
	QueryBeneficialPath                 = "queryBeneficiaryName"
	QueryDirectDBPath                   = "queryDirectDebit"
	CancelDirectDBPath                  = "directDebitCancel"
	SandboxPublicKey                    = "MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEArv9yxA69XQKBo24BaF/D+fvlqmGdYjqLQ5WtNBb5tquqGvAvG3WMFETVUSow/LizQalxj2ElMVrUmzu5mGGkxK08bWEXF7a1DEvtVJs6nppIlFJc2SnrU14AOrIrB28ogm58JjAl5BOQawOXD5dfSk7MaAA82pVHoIqEu0FxA8BOKU+RGTihRU+ptw1j4bsAJYiPbSX6i71gfPvwHPYamM0bfI4CmlsUUR3KvCG24rB6FNPcRBhM3jDuv8ae2kC33w9hEq8qNB55uw51vK7hyXoAa+U7IqP1y6nBdlN25gkxEA8yrsl1678cspeXr+3ciRyqoRgj9RD/ONbJhhxFvt1cLBh+qwK2eqISfBb06eRnNeC71oBokDm3zyCnkOtMDGl7IvnMfZfEPFCfg5QgJVk1msPpRvQxmEsrX9MQRyFVzgy2CWNIb7c+jPapyrNwoUbANlN8adU1m6yOuoX7F49x+OjiG2se0EJ6nafeKUXw/+hiJZvELUYgzKUtMAZVTNZfT8jjb58j8GVtuS+6TM2AutbejaCV84ZK58E2CRJqhmjQibEUO6KPdD7oTlEkFy52Y1uOOBXgYpqMzufNPmfdqqqSM4dU70PO8ogyKGiLAIxCetMjjm6FCMEA3Kc8K0Ig7/XtFm9By6VxTJK1Mg36TlHaZKP6VzVLXMtesJECAwEAAQ=="
	OpenapiPublicKey                    = "MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAietPTdEyyoV/wvxRjS5pSn3ZBQH9hnVtQC9SFLgM9IkomEX9Vu9fBg2MzWSSqkQlaYIGFGH3d69Q5NOWkRo+Y8p5a61sc9hZ+ItAiEL9KIbZzhnMwi12jUYCTff0bVTsTGSNUePQ2V42sToOIKCeBpUtwWKhhW3CSpK7S1iJhS9H22/BT/pk21Jd8btwMLUHfVD95iXbHNM8u6vFaYuHczx966T7gpa9RGGXRtiOr3ScJq1515tzOSOsHTPHLTun59nxxJiEjKoI4Lb9h6IlauvcGAQHp5q6/2XmxuqZdGzh39uLac8tMSmY3vC3fiHYC3iMyTb7eXqATIhDUOf9mOSbgZMS19iiVZvz8igDl950IMcelJwcj0qCLoufLE5y8ud5WIw47OCVkD7tcAEPmVWlCQ744SIM5afw+Jg50T1SEtu3q3GiL0UQ6KTLDyDEt5BL9HWXAIXsjFdPDpX1jtxZavVQV+Jd7FXhuPQuDbh12liTROREdzatYWRnrhzeOJ5Se9xeXLvYSj8DmAI4iFf2cVtWCzj/02uK4+iIGXlX7lHP1W+tycLS7Pe2RdtC2+oz5RSSqb5jI4+3iEY/vZjSMBVk69pCDzZy4ZE8LBgyEvSabJ/cddwWmShcRS+21XvGQ1uXYLv0FCTEHHobCfmn2y8bJBb/Hct53BaojWUCAwEAAQ=="
	ReqNewSessionKeyBeforeExpiresIn = time.Duration(60) * time.Second
)

type (
	Client struct {
		mu          sync.Mutex
		Client      *http.Client
		Keys        *Keys
		Environment string
		SessionKey  string
		ExpiresAt   time.Time
		SessionLife int32
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
		Amount                   string `json:"input_Amount"`
		CustomerMSISDN           string `json:"input_CustomerMSISDN"`
		Country                  string `json:"input_Country"`
		Currency                 string `json:"input_Currency"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		TransactionReference     string `json:"input_TransactionReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		PurchasedItemsDesc       string `json:"input_PurchasedItemsDesc"`
	}

	C2BPaymentResponse struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionID            string `json:"output_TransactionID"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	B2BPaymentRequest struct {
		Amount                   string `json:"input_Amount"`
		ReceiverPartyCode        string `json:"input_ReceiverPartyCode"`
		Country                  string `json:"input_Country"`
		Currency                 string `json:"input_Currency"`
		PrimaryPartyCode         string `json:"input_PrimaryPartyCode"`
		TransactionReference     string `json:"input_TransactionReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		PurchasedItemsDesc       string `json:"input_PurchasedItemsDesc"`
	}
	B2BPaymentResponse struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionID            string `json:"output_TransactionID"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	B2CPaymentRequest struct {
		Amount                   string `json:"input_Amount"`
		CustomerMSISDN           string `json:"input_CustomerMSISDN"`
		Country                  string `json:"input_Country"`
		Currency                 string `json:"input_Currency"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		TransactionReference     string `json:"input_TransactionReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		PaymentItemsDesc         string `json:"input_PaymentItemsDesc"`
	}

	B2CPaymentResponse struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionID            string `json:"output_TransactionID"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	ReversalRequest struct {
		ReversalAmount           string `json:"input_ReversalAmount,omitempty"`
		Country                  string `json:"input_Country"`
		TransactionID            string `json:"input_TransactionID"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
	}

	ReversalResponse struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionID            string `json:"output_TransactionID"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	QueryTxStatusRequest struct {
		QueryReference           string `json:"input_QueryReference"`
		Country                  string `json:"input_Country"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
	}

	QueryTxStatusResponse struct {
		ResponseCode              string `json:"output_ResponseCode"`
		ResponseDesc              string `json:"output_ResponseDesc"`
		ResponseTransactionStatus string `json:"output_ResponseTransactionStatus"`
		ConversationID            string `json:"output_ConversationID"`
		ThirdPartyConversationID  string `json:"output_ThirdPartyConversationID"`
		OriginalTransactionID     string `json:"output_OriginalTransactionID"`
	}

	DirectDBCreateReq struct {
		CustomerMSISDN           string `json:"input_CustomerMSISDN"`
		Country                  string `json:"input_Country"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyReference      string `json:"input_ThirdPartyReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		AgreedTC                 string `json:"input_AgreedTC"`
		FirstPaymentDate         string `json:"input_FirstPaymentDate,omitempty"`
		Frequency                string `json:"input_Frequency,omitempty"`
		StartRangeOfDays         string `json:"input_StartRangeOfDays,omitempty"`
		EndRangeOfDays           string `json:"input_EndRangeOfDays,omitempty"`
		ExpiryDate               string `json:"input_ExpiryDate,omitempty"`
	}

	DirectDBCreateRes struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionReference     string `json:"output_TransactionReference"`
		MsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	DebitDBPaymentReq struct {
		MsisdnToken              string `json:"input_MsisdnToken,omitempty"`
		CustomerMSISDN           string `json:"input_CustomerMSISDN,omitempty"`
		Country                  string `json:"input_Country"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyReference      string `json:"input_ThirdPartyReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		Amount                   string `json:"input_Amount"`
		Currency                 string `json:"input_Currency"`
		MandateID                string `json:"input_MandateID,omitempty"`
	}

	DebitDBPaymentRes struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionID            string `json:"output_TransactionID"`
		MsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID,omitempty"`
	}

	QueryBenRequest struct {
		CustomerMSISDN           string `json:"input_CustomerMSISDN"`
		Country                  string `json:"input_Country"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		KycQueryType             string `json:"input_KycQueryType"`
	}

	QueryBenResponse struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		CustomerFirstName        string `json:"output_CustomerFirstName"`
		CustomerLastName         string `json:"output_CustomerLastName"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}

	QueryDirectDBReq struct {
		QueryBalanceAmount       bool   `json:"input_QueryBalanceAmount"`
		BalanceAmount            string `json:"input_BalanceAmount,omitempty"`
		Country                  string `json:"input_Country"`
		CustomerMSISDN           string `json:"input_CustomerMSISDN,omitempty"`
		MsisdnToken              string `json:"input_MsisdnToken,omitempty"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		ThirdPartyReference      string `json:"input_ThirdPartyReference"`
		MandateID                string `json:"input_MandateID,omitempty"`
		Currency                 string `json:"input_Currency"`
	}
	QueryDirectDBRes struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionReference     string `json:"output_TransactionReference"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
		SufficientBalance        bool   `json:"output_SufficientBalance"`
		MsisdnToken              string `json:"output_MsisdnToken,omitempty"`
		MandateID                string `json:"output_MandateID"`
		MandateStatus            string `json:"output_MandateStatus"`
		AccountStatus            string `json:"output_AccountStatus"`
		FirstPaymentDate         string `json:"output_FirstPaymentDate,omitempty"`
		Frequency                string `json:"output_Frequency,omitempty"`
		PaymentDayFrom           string `json:"output_PaymentDayFrom,omitempty"`
		PaymentDayTo             string `json:"output_PaymentDayTo,omitempty"`
		ExpiryDate               string `json:"output_ExpiryDate,omitempty"`
	}

	CancelDirectDBReq struct {
		MsisdnToken              string `json:"input_MsisdnToken"`
		CustomerMSISDN           string `json:"input_CustomerMSISDN"`
		Country                  string `json:"input_Country"`
		ServiceProviderCode      string `json:"input_ServiceProviderCode"`
		ThirdPartyReference      string `json:"input_ThirdPartyReference"`
		ThirdPartyConversationID string `json:"input_ThirdPartyConversationID"`
		MandateID                string `json:"input_MandateID"`
	}

	CancelDirectDBRes struct {
		ResponseCode             string `json:"output_ResponseCode"`
		ResponseDesc             string `json:"output_ResponseDesc"`
		TransactionReference     string `json:"output_TransactionReference"`
		MsisdnToken              string `json:"output_MsisdnToken"`
		ConversationID           string `json:"output_ConversationID"`
		ThirdPartyConversationID string `json:"output_ThirdPartyConversationID"`
	}
)
