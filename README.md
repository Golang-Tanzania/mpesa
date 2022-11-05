# GoPesa

Golang bindings for [Mpesa Payment API](openapiportal.m-pesa.com/). Made by [Mojo](https://github.com/AvicennaJr) and [HoperTZ](https://github.com/Hopertz)

## Getting Started

## Examples
```
package gopesa

import (
	"fmt"
)

func main() {

	var test1 APICONTEXT

	test1.Initialize("config.json")

	// Example on how to make a Business to Business Transaction

	b2BtransactionQuery := make(map[string]string)

	b2BtransactionQuery["input_Amount"] = "10"
	b2BtransactionQuery["input_Country"] = "TZN"
	b2BtransactionQuery["input_Currency"] = "TZS"
	b2BtransactionQuery["input_PrimaryPartyCode"] = "000000"
	b2BtransactionQuery["input_ReceiverPartyCode"] = "000001"
	b2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	b2BtransactionQuery["input_ThirdPartyConversationID"] = "8a89835c71f15e99396"
	b2BtransactionQuery["input_TransactionReference"] = "T1234C"
	b2BtransactionQuery["input_PurchasedItemsDesc"] = "Shoes"

	fmt.Println(test1.B2BPayments(b2BtransactionQuery))

	// Example on how to make a Customer to Business Transaction

	var test2 APICONTEXT

	test2.Initialize("config.json")

	c2BtransactionQuery := make(map[string]string)

	c2BtransactionQuery["input_Amount"] = "10"
	c2BtransactionQuery["input_CustomerMSISDN"] = "000000000001"
	c2BtransactionQuery["input_Country"] = "TZN"
	c2BtransactionQuery["input_Currency"] = "TZS"
	c2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	c2BtransactionQuery["input_TransactionReference"] = "T12344C"
	c2BtransactionQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
	c2BtransactionQuery["input_PurchasedItemsDesc"] = "Shoes"

	fmt.Println(test2.C2BPayments(c2BtransactionQuery))
}
```