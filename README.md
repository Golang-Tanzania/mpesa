# GoPesa

Golang bindings for the [Mpesa Payment API](openapiportal.m-pesa.com/). Make your MPESA payments *Ready... To... Gooo!* (*pun intended*). Made with love for gophers.

## Features

- [x] Customer to Bussiness (C2B) Single Payment
- [x] Bussiness to Bussiness (B2B)
- [x] Bussiness to Customer (B2C)
- [ ] Payment Reversal
- [ ] Query Transaction status
- [ ] Direct debit creation and Payment

## Pre-requisites

- First sign up with [Mpesa](https://openapiportal.m-pesa.com/sign-up) to get your API and PUBLIC keys. You can go through this blog, [Getting Started With Mpesa Developer API](https://dev.to/alphaolomi/getting-started-with-mpesa-developer-portal-46a4) for a more detailed guide.

- Then place your Keys (API and Public key) in a file called `config.json`.

## Installation

Simply install with the `go get` command:
```
go get github.com/Golang-Tanzania/GoPesa
```
Then import it to your main package as:
```
package main

import (
	gopesa "github.com/Golang-Tanzania/GoPesa"
)
```

## Usage

First create a new variable of type `gopesa.APICONTEXT` and then call the `Initialize` method with the path to your `config.json` as follows:
```
var test gopesa.APICONTEXT

test.Initialize("config.json")
```
Assuming you want to make a `Customer To Business` transaction, create a new `map` with the required parameters as below:
```
// create a new map with a string key and a string value

transactionQuery := make(map[string]string)

// map each transaction query key to its value

transactionQuery["input_Amount"] = "10"
transactionQuery["input_CustomerMSISDN"] = "000000000001"
transactionQuery["input_Country"] = "TZN"
transactionQuery["input_Currency"] = "TZS"
transactionQuery["input_ServiceProviderCode"] = "000000"
transactionQuery["input_TransactionReference"] = "T12344C"
transactionQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
transactionQuery["input_PurchasedItemsDesc"] = "Shoes"
```
Then finally call the Customer To Business method to request a payment:
```
fmt.Println(test.C2BPayments(transactionQuery))

// Output
{
    "output_ResponseCode":"INS-0",
    "output_ResponseDesc":"Request processed successfully"
    "output_TransactionID":"cUmNsY2j0Fr5",
    "output_ConversationID":"8cba707babcf4b36921f9ff1bd957cb1",
    "output_ThirdPartyConversationID":"8a89835c71f15e99396"
}
```
And that's it!

## Setting Environment

You can set your desired environment, ie `Production` or `Sandbox` with the `ENVIRONMENT` keyword:
```
var test gopesa.APICONTEXT
test.ENVIRONMENT = "Production"
```
**The default environment is Sandbox**

## More Examples

Below are more examples on how to make API transactions.

### Customer To Business
```
package main

import (
	gopesa "github.com/Golang-Tanzania/GoPesa"
    "fmt"
)

func main() {

   // Create a new variable of type gopesa.APICONTEXT

    var test gopesa.APICONTEXT

    // Initialize and set defaults

    test.Initialize("config.json")

    // Create a new map query that maps strings to strings

	c2BtransactionQuery := make(map[string]string)

	c2BtransactionQuery["input_Amount"] = "10"
	c2BtransactionQuery["input_CustomerMSISDN"] = "000000000001"
	c2BtransactionQuery["input_Country"] = "TZN"
	c2BtransactionQuery["input_Currency"] = "TZS"
	c2BtransactionQuery["input_ServiceProviderCode"] = "000000"
	c2BtransactionQuery["input_TransactionReference"] = "T12344C"
	c2BtransactionQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
	c2BtransactionQuery["input_PurchasedItemsDesc"] = "Shoes"

	fmt.Println(test2.C2BPayment(c2BtransactionQuery))
}

```
### Business To Customer
```
package main

import (
	gopesa "github.com/Golang-Tanzania/GoPesa"
    "fmt"
)

func main() {

    // Create a new variable of type gopesa.APICONTEXT

    var test gopesa.APICONTEXT

    // Initialize and set defaults

    test.Initialize("config.json")

    // Create a new map query that maps strings to strings

    b2CtransactionQuery := make(map[string]string)

    b2CtransactionQuery["input_Amount"] = "250"
	b2CtransactionQuery["input_Country"] = "TZN"
	b2CtransactionQuery["input_Currency"] = "TZS"
	b2CtransactionQuery["input_CustomerMSISDN"] = "000000000001"
	b2CtransactionQuery["input_ServiceProviderCode"] = "000000"
	b2CtransactionQuery["input_ThirdPartyConversationID"] = "f5e420e99594a9c496d8600"
	b2CtransactionQuery["input_TransactionReference"] = "T12345C"
	b2CtransactionQuery["input_PaymentItemsDesc"] = "Donation"

    fmt.Println(test.B2CPayment(b2CtransactionQuery))

```
### Business To Business

```
package main

import (
	gopesa "github.com/Golang-Tanzania/GoPesa"
    "fmt"
)

func main() {

    // Create a new variable of type gopesa.APICONTEXT

    var test gopesa.APICONTEXT

    // Initialize and set defaults

    test.Initialize("config.json")

    // Create a new map query that maps strings to strings

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

	fmt.Println(test.B2BPayment(b2BtransactionQuery))
```

## Authors

This package is authored and maintained by [Mojo](https://github.com/AvicennaJr) and [HoperTZ](https://github.com/Hopertz)

## License

MIT License

Copyright (c) 2022 Golang Tanzania