package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type KEYS struct {
	PUBLICKEY string
	APIKEY    string
}

func main() {

	keys, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error opening config file: ", err)
	}
	var Keys KEYS
	err = json.Unmarshal(keys, &Keys)
	if err != nil {
		log.Fatal("Error during Unmarshal: ", err)
	}

	publicKey := fmt.Sprintf(`
-----BEGIN RSA PUBLIC KEY-----
%s
-----END RSA PUBLIC KEY-----`, Keys.PUBLICKEY)
	apiKey := Keys.APIKEY

	test := APICONTEXT{PUBLICKEY: publicKey, APIKEY: apiKey}

	// b2BtransactionQuery := make(map[string]string)

	// b2BtransactionQuery["input_Amount"] = "10"
	// b2BtransactionQuery["input_Country"] = "TZN"
	// b2BtransactionQuery["input_Currency"] = "TZS"
	// b2BtransactionQuery["input_PrimaryPartyCode"] = "000000"
	// b2BtransactionQuery["input_ReceiverPartyCode"] = "000001"
	// b2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	// b2BtransactionQuery["input_ThirdPartyConversationID"] = "8a89835c71f15e99396"
	// b2BtransactionQuery["input_TransactionReference"] = "T1234C"
	// b2BtransactionQuery["input_PurchasedItemsDesc"] = "Shoes"

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
