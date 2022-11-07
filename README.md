# GoPesa

![GoPesa](./assets/gopesa.svg)

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/Golang-Tanzania/GoPesa)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/Golang-Tanzania/GoPesa)

Golang bindings for the [Mpesa Payment API](openapiportal.m-pesa.com/). Make your MPESA payments *Ready... To... Gooo!* (*pun intended*). Made with love for gophers.

## Features

- [x] Customer to Bussiness (C2B) Single Payment
- [x] Bussiness to Bussiness (B2B)
- [x] Bussiness to Customer (B2C)
- [x] Payment Reversal
- [x] Query Transaction status
- [x] Query Beneficiary Name
- [ ] Query Direct Debit
- [ ] Direct Debit Create
- [ ] Direct Debit Payment

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
	gopesa "github.com/Golang-Tanzania/GoPesa@v0.1.2"
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
}

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
}
```


### Payment Reversal

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

    paymentReversaltranscQuery := make(map[string]string)

    paymentReversaltranscQuery["input_ReversalAmount"] = "25"
    paymentReversaltranscQuery["input_Country"] = "TZN"
    paymentReversaltranscQuery["input_ServiceProviderCode"] = "000000"
    paymentReversaltranscQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
    paymentReversaltranscQuery["input_TransactionID"] = "0000000000001"

    fmt.Println(test.ReversePayment(paymentReversaltranscQuery))
}
```

### Query Transaction status

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

    transanctionStatusQuery := make(map[string]string)

    transanctionStatusQuery["input_QueryReference"] = "000000000000000000001"
    transanctionStatusQuery["input_ServiceProviderCode"] = "000000"
    transanctionStatusQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
    transanctionStatusQuery["input_Country"] = "TZN"

    fmt.Println(test.TransactionStatus(transanctionStatusQuery))
}
```

### Query Beneficiary Name

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

    BeneficiaryNameQuery := make(map[string]string)

    BeneficiaryNameQuery["input_CustomerMSISDN"] = "255742051622"
    BeneficiaryNameQuery["input_ServiceProviderCode"] = "000000"
    BeneficiaryNameQuery["input_ThirdPartyConversationID"] = "asv02e5958774f7ba228d83d0d689761"
    BeneficiaryNameQuery["input_Country"] = "TZN"
    BeneficiaryNameQuery["input_KycQueryType"] = "Name"
    fmt.Println(test.QueryBeneficiaryName(BeneficiaryNameQuery))
}
```
### Query Direct Debit

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

    DirectDebitQuery := make(map[string]string)

    DirectDebitQuery["input_QueryBalanceAmount"] = "True"
    DirectDebitQuery["input_BalanceAmount"] = "100"
    DirectDebitQuery["input_Country"] = "TZN"
    DirectDebitQuery["input_CustomerMSISDN"] = "255744553111"
    DirectDebitQuery["input_MsisdnToken"] = "cvgwUBZ3lAO9ivwhWAFeng=="
    DirectDebitQuery["input_ServiceProviderCode"] = "112244"
    DirectDebitQuery["input_ThirdPartyConversationID"] = "GPO3051656128"
    DirectDebitQuery["input_ThirdPartyReference"] = "Test123"
    DirectDebitQuery["input_MandateID"] = "15045"
    DirectDebitQuery["input_Currency"] = "TZS"


    fmt.Println(test.QueryDirectDebit(DirectDebitQuery))
}
```
## Authors

This package is authored and maintained by [Mojo](https://github.com/AvicennaJr) and [Hopertz](https://github.com/Hopertz)

## License

MIT License

Copyright (c) 2022 Golang Tanzania