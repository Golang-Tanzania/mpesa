# Mpesa

<img src="./assets/mpesa.svg" alt="Mpesa for Mpesa" height="300px" align="right">

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/Golang-Tanzania/Mpesa)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/Golang-Tanzania/mpesa)

Golang client for the [Mpesa Payment API](openapiportal.m-pesa.com/). Make your MPESA payments _Ready... To... Gooo!_ (_pun intended_). Made with love for gophers.

## Features

- [x] Customer to Business (C2B) Single Payment
- [x] Business to Business (B2B)
- [x] Business to Customer (B2C)
- [x] Payment Reversal
- [x] Query Transaction status
- [x] Query Beneficiary Name
- [x] Query Direct Debit
- [x] Direct Debit Create
- [x] Direct Debit Payment

## Pre-requisites

- First sign up with [Mpesa](https://openapiportal.m-pesa.com/sign-up) to get your API Key and PUBLIC key. 

    You can go through this blog, [Getting Started With Mpesa Developer API](https://dev.to/alphaolomi/getting-started-with-mpesa-developer-portal-46a4) for a more detailed guide.



## Installation

Simply install with the `go get` command:

```sh
go get github.com/Golang-Tanzania/mpesa
```

Then import it to your main package as:

```go
package main

import (
	mpesa "github.com/Golang-Tanzania/mpesa
)
```

## Usage 

```go
   // NewClient returns new Client struct
   client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24)
   // "your-api-key" obtained at https://openapiportal.m-pesa.com
   // sandbox is environment type can either be mpesa.Sandbox or mpesa.Production
   // 24 represent hours, its session lifetime before we request another session it can be found in the 
   // https://openapiportal.m-pesa.com/applications where you set for your application,



   // use custom htttp client

   c :=  http.Client{
	      Timeout: 50 * time.Second,
	},

   client.SetHttpClient(c)

```


## Examples

Below are examples on how to make different transactions.

### Customer To Business

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

    client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}

	a := mpesa.C2BPaymentRequest{
		Amount:                   "100",
		CustomerMSISDN:           "000000000001",
		Country:                  "TZN",
		Currency:                 "TZS",
		ServiceProviderCode:      "000000",
		TransactionReference:     "T12344C",
		ThirdPartyConversationID: "asv02e5958774f7ba228d83d0d689761",
		PurchasedItemsDesc:       "Test",
	}
	res, err := client.C2BPayment(context.Background(), a)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Business To Customer

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}

	c := mpesa.B2CPaymentRequest{
		Amount:                   "100",
		CustomerMSISDN:           "000000000001",
		Country:                  "TZN",
		Currency:                 "TZS",
		ServiceProviderCode:      "000000",
		TransactionReference:     "T12344C",
		ThirdPartyConversationID: "asv02e5958774f7ba228d83d0d689761",
		PaymentItemsDesc:       "Test",
	}

	res, err := client.B2CPayment(context.Background(), c)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)

}

```

### Business To Business

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	b := mpesa.B2BPaymentRequest{
		Amount:            "100",
		Country:           "TZN",
		Currency:          "TZS",
		PrimaryPartyCode: "000000",
		ReceiverPartyCode: "000001",
		ThirdPartyConversationID: "8a89835c71f15e99396",
		TransactionReference: "T12344C",
		PurchasedItemsDesc: "Test",
	}

	res, err := client.B2BPayment(context.Background(), b)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Payment Reversal

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	d :=  mpesa.ReversalRequest{
		TransactionID: "0000000000001",
		Country: "TZN",
		ServiceProviderCode: "000000",
		ReversalAmount: "100",
		ThirdPartyConversationID: "asv02e5958774f7ba228d83d0d689761",
	}

	res, err := client.Reversal(context.Background(), d)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Query Transaction status

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {
	
	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	e := mpesa.QueryTxStatusRequest{
		QueryReference:           "000000000000000000001",
		Country:                  "TZN",
		ServiceProviderCode:      "000000",
		ThirdPartyConversationID: "asv02e5958774f7ba228d83d0d689761",
	}

	res, err := client.QueryTxStatus(context.Background(), e)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Query Beneficiary Name

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {
	
	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	h := mpesa.QueryBenRequest{
		Country:                  "TZN",
		ServiceProviderCode:      "000000",
		ThirdPartyConversationID: "AAA6d1f939c1005v2de053v4912jbasdj1j2kk",
		CustomerMSISDN:           "000000000001",
		KycQueryType:             "Name",
	}

	res, err := client.QueryBeneficiaryName(context.Background(), h)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Query Direct Debit

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {
	
	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	i := mpesa.QueryDirectDBReq{
		QueryBalanceAmount:       true,
		BalanceAmount:            "100",
		Country:                  "TZN",
		CustomerMSISDN:           "255744553111",
		MsisdnToken:              "cvgwUBZ3lAO9ivwhWAFeng==",
		ServiceProviderCode:      "112244",
		ThirdPartyConversationID: "GPO3051656128",
		ThirdPartyReference:      "Test123",
		MandateID:                "15045",
		Currency:                 "TZS",
	}

	res, err := client.QueryDirectDebit(context.Background(), i)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Direct Debit Create

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {
	
	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	f := mpesa.DirectDBCreateReq{
		CustomerMSISDN:         "000000000001",
		Country:                "TZN",
		ServiceProviderCode:    "000000",
		ThirdPartyReference:    "3333",
		ThirdPartyConversationID: "asv02e5958774f7ba228d83d0d689761",
		AgreedTC:               "1",
		FirstPaymentDate:       "20160324",
		Frequency:              "06",
		StartRangeOfDays:       "01",
		EndRangeOfDays:         "22",
		ExpiryDate:             "20161126",
	}

	res, err := client.DirectDebitCreate(context.Background(), f)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Direct Debit Payment

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

	client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	g := mpesa.DebitDBPaymentReq{
		MsisdnToken:              "AbCd123=",
		CustomerMSISDN:           "000000000001",
		Country:                  "TZN",
		ServiceProviderCode:      "000000",
		ThirdPartyReference:      "5db410b459bd433ca8e5",
		ThirdPartyConversationID: "AAA6d1f939c1005v2de053v4912jbasdj1j2kk",
		Amount:                   "10",
		Currency:                 "TZS",
		MandateID:                "15045",
	}

	res, err := client.DirectDebitPayment(context.Background(), g)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}
```

### Cancel Direct Debit 
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Golang-Tanzania/mpesa"
)

func main() {

    client, err := mpesa.NewClient("your-api-key", mpesa.Sandbox, 24) 
	if err != nil {
		panic(err)
	}


	j := mpesa.CancelDirectDBReq{
		MsisdnToken:              "cvgwUBZ3lAO9ivwhWAFeng==",
		CustomerMSISDN:           "000000000001",
		Country:                  "TZN",
		ServiceProviderCode:      "000000",
		ThirdPartyReference:      "00000000000000000001",
		ThirdPartyConversationID: "AAA6d1f939c1005v2de053v4912jbasdj1j2kk",
		MandateID:                "15045",
	}

	res, err := client.CancelDirectDebit(context.Background(), j)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("res", res)
}

```

## Authors

This package is authored and maintained by [Hopertz](https://github.com/Hopertz) and [Mojo](https://github.com/AvicennaJr).
A list of all other contributors can be found [here](https://github.com/Golang-Tanzania/mpesa/graphs/contributors).

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.


## License

MIT License

Copyright (c) 2023 Golang Tanzania
