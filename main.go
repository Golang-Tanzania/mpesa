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

	transactionQuery := make(map[string]string)

	transactionQuery["input_Amount"] = "10"
	transactionQuery["input_Country"] = "TZN"
	transactionQuery["input_Currency"] = "TZS"
	transactionQuery["input_ReceiverPartyCode"] = "000001"
	transactionQuery["input_PrimaryPartyCode"] = "000000"
	transactionQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
	transactionQuery["input_TransactionReference"] = "T1234C"
	transactionQuery["input_PurchasedItemsDesc"] = "Shoes"

	fmt.Println(test.C2B(transactionQuery))
}
