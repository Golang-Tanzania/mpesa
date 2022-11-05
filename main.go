package main

import (
	"fmt"
)

func main() {
	public_key := `
-----BEGIN RSA PUBLIC KEY-----
YOUR PUBLIC KEY
-----END RSA PUBLIC KEY-----`
	api_key := "YOUR API KEY"

	test := APICONTEXT{PUBLICKEY: public_key, APIKEY: api_key}

	// b2BtransactionQuery := make(map[string]string)

	// b2BtransactionQuery["input_Amount"] = "10"
	// b2BtransactionQuery["input_Country"] = "TZN"
	// b2BtransactionQuery["input_Currency"] = "TZS"
	// b2BtransactionQuery["input_CustomerMSISDN"] = "000000000001"
	// b2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	// b2BtransactionQuery["input_ThirdPartyConversationID"] = "2edf7a0206d848f6b6fedea26accdc3a"
	// b2BtransactionQuery["input_TransactionReference"] = "T23434ZE5"
	// b2BtransactionQuery["input_PurchasedItemsDesc"] = "Python Book"

	// // fmt.Println(test.generateSessionID())
	// fmt.Println(test.B2BPayments(b2BtransactionQuery))

	c2BtransactionQuery := make(map[string]string)

	c2BtransactionQuery["input_Amount"] = "10"
	c2BtransactionQuery["input_CustomerMSISDN"] = "000000000001"
	c2BtransactionQuery["input_Country"] = "TZN"
	c2BtransactionQuery["input_Currency"] = "TZS"
	c2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	c2BtransactionQuery["input_TransactionReference"] = "T12344C"
	c2BtransactionQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
	c2BtransactionQuery["input_PurchasedItemsDesc"] = "Shoes"

	fmt.Println(test.C2BPayments(c2BtransactionQuery))
}
