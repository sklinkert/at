# igmarkets - Unofficial IG Markets Trading API for Golang

This is an **unofficial** API for [IG Markets Trading REST API](https://labs.ig.com/rest-trading-api-reference). The StreamingAPI is not part of this project.

**Disclaimer**: This library is not associated with IG Markets Limited or any of its affiliates or subsidiaries. If you use this library, you should contact them to make sure they are okay with how you intend to use it. Use this lib at your own risk.

Reference: https://labs.ig.com/rest-trading-api-reference

## Currently supported endpoints

### Lightstreamer

- Create session, add subscription(control), bind session

### Account

- GET /accounts
- GET /accounts/preferences

### Session

- POST /session (version 2 + 3)

### Markets

- GET /markets/{epic}
- GET /markets?searchTerm=...

### Client sentiment

- GET /clientsentiment/{marketID}

### Positions

- POST /positions/otc
- PUT /positions/otc/{dealId}
- GET /positions
- DELETE /positions
- GET /confirms/{dealReference}

### Workingorders
- GET /workingorders
- POST /workingorders/otc
- DELETE /workingorders/otc/{dealId}

### Prices

- GET /prices/{epic}/{resolution}/{startDate}/{endDate}

### Watchlists
- POST /watchlists/ (Create watchlist)
- GET /watchlists/{watchlistid}
- DELETE /watchlists/{watchlistid} (Delete watchlist)

- GET /watchlists (Get all watchlists)
- PUT /watchlists/{watchlistid} (Add epic)
- DELETE /watchlists/{watchlistid}/{epic} (Delete epic)

### History

- GET /history/activity
- GET /history/transactions

## Example

```go
package main

import (
	    "context"
        "fmt"
        "github.com/sklinkert/igmarkets"
        "time"
)

var ig *igmarkets.IGMarkets

func main() {
	    var ctx = context.Background()
        ig = igmarkets.New(igmarkets.DemoAPIURL, "APIKEY", "ACCOUNTID", "USERNAME/IDENTIFIER", "PASSWORD")
        if err := ig.Login(ctx); err != nil {
                fmt.Println("Unable to login into IG account", err)
        }

        // Get current open ask, open bid, close ask, close bid, high ask, high bid, low ask, and low bid
        prices, _ := ig.GetPrice(ctx, "CS.D.EURUSD.CFD.IP")

        fmt.Println(prices)

        // Place a new order
        order := igmarkets.OTCOrderRequest{
                Epic:           "CS.D.EURUSD.CFD.IP",
                OrderType:      "MARKET",
                CurrencyCode:   "USD",
                Direction:      "BUY",
                Size:           1.0,
                Expiry:         "-",
                StopDistance:   "10", // Pips
                LimitDistance:  "5",  // Pips
                GuaranteedStop: true,
                ForceOpen:      true,
        }
        dealRef, err := ig.PlaceOTCOrder(ctx, order)
        if err != nil {
                fmt.Println("Unable to place order:", err)
                return
        }
        fmt.Println("New order placed with dealRef", dealRef)

        // Check order status
        confirmation, err := ig.GetDealConfirmation(ctx, dealRef.DealReference)
        if err != nil {
                fmt.Println("Cannot get deal confirmation for:", dealRef, err)
                return
        }

        fmt.Println("Order dealRef", dealRef)
        fmt.Println("DealStatus", confirmation.DealStatus) // "ACCEPTED"
        fmt.Println("Profit", confirmation.Profit, confirmation.ProfitCurrency)
        fmt.Println("Status", confirmation.Status) // "OPEN"
        fmt.Println("Reason", confirmation.Reason)
        fmt.Println("Level", confirmation.Level) // Buy price

        // List transactions
        transactionResponse, err := ig.GetTransactions(ctx, "ALL", time.Now().AddDate(0, 0, -30).UTC()) // last 30 days
        if err != nil {
                fmt.Println("Unable to get transactions: ", err)
        }
        for _, transaction := range transactionResponse.Transactions {
                fmt.Println("Found new transaction")
                fmt.Println("Epic:", transaction.InstrumentName)
                fmt.Println("Type:", transaction.TransactionType)
                fmt.Println("OpenDate:", transaction.OpenDateUtc)
                fmt.Println("CloseDate:", transaction.DateUTC)
                fmt.Println("OpenLevel:", transaction.OpenLevel)
                fmt.Println("CloseLevel:", transaction.CloseLevel)
                fmt.Println("Profit/Loss:", transaction.ProfitAndLoss)
	}

        // Example of getting client sentiment
        sentiment, _ := ig.GetClientSentiment(ctx, "F-US") //Ford
        fmt.Println("Sentiment example:", sentiment)
}
```

More examples can be found [here](https://github.com/sklinkert/igmarkets/tree/master/examples).

### LightStreamer API Subscription Example

```go
    var ctx = context.Background()
	for {
		tickChan := make(chan igmarkets.LightStreamerTick)
    err := igHandle.OpenLightStreamerSubscription(ctx, []string{"CS.D.BITCOIN.CFD.IP"}, tickChan)
		if err != nil {
      log.WithError(err).Error("OpenLightStreamerSubscription() failed")
		}

		for tick := range tickChan {
			log.Infof("tick: %+v", tick)
		}

		log.Infof("Server closed stream, restarting...")
  }
```

Output:

```
INFO[0003] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:15 +0000 UTC Bid:18230.35 Ask:18266.35} 
INFO[0003] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:15 +0000 UTC Bid:18230.45 Ask:18266.45} 
INFO[0003] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:16 +0000 UTC Bid:18231.14 Ask:18267.14} 
INFO[0003] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:16 +0000 UTC Bid:18231.04 Ask:18267.04} 
INFO[0004] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:16 +0000 UTC Bid:18231.53 Ask:18267.53} 
INFO[0004] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:16 +0000 UTC Bid:18231.35 Ask:18267.35} 
INFO[0004] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:17 +0000 UTC Bid:18230.64 Ask:18266.64} 
INFO[0004] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:17 +0000 UTC Bid:18231.08 Ask:18267.08} 
INFO[0005] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:17 +0000 UTC Bid:18231.36 Ask:18267.36} 
INFO[0005] tick: {Epic:CS.D.BITCOIN.CFD.IP Time:2020-11-22 14:14:17 +0000 UTC Bid:18230.93 Ask:18266.93} 
```





## TODOs

- Write basic tests

Feel free to send PRs.
