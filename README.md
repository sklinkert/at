# Automated Trader (at)

[![tests](https://github.com/sklinkert/at/actions/workflows/ci.yaml/badge.svg)](https://github.com/sklinkert/at/actions/workflows/ci.yaml)

**Purpose**: Framework for building automated trading strategies in three steps:

1. Build your own strategy.
2. Verify it with the backtest module.
3. Connect it to a real broker API to make some money.

Every broker API can be implemented. Works for stocks, forex, cryptocurrencies, etc.

**Supported brokers**:

|  Broker    |  Demo account | Paperwallet trading |  Real trading   | Backtesting (historical prices) |
| ---- | ---- | ---- | ---- | ---- |
| IG.com   |  ✅    |   ❌   | ✅ | ✅
|  Coinbase    |   ❌   |  ✅    | ❌ | ✅
|  FTX    |   ❌   |  ✅    | ❌ | ❌


**Disclaimer**: The developers are not liable for any losses arising from the buy or sell of securities. All included strategies are examples and in no case ready trading systems.

## Installation

```sh
go get github.com/sklinkert/at
```

Example backtesting run:

```shell
INSTRUMENT="BTC-USD" STRATEGY="rsi" CANDLE_DURATION=24h YEAR_FROM=2021 MONTH_FROM=4 YEAR_TO=2021 MONTH_TO=12 PRICE_SOURCE="COINBASE" go run -ldflags="-w -s -X main.GitRev=123" ./cmd/backtesting/main.go
```

Uses ETHUSD candles from Coinbase via api.pattern-trading.com

## Overview

![Overview](docs/overview.png)

## Packages

### broker

Implements concrete broker API. [paperwallet](https://github.com/sklinkert/at/tree/master/internal/paperwallet) can be used if a broker does not offer testing/sandbox accounts for trading without real money. 

### trader

The nerve center of the program. It connects the broker with strategies.

### indicator

Implements various trading indicator like Simple Moving Average (SMA) or RSI. Indicators should implement [this interface](https://github.com/sklinkert/at/blob/master/pkg/indicator/indicator.go).

### strategy

Implements various trading strategies. Each strategy needs to implement [this interface](https://github.com/sklinkert/at/blob/master/internal/strategy/strategy.go#L26) thus the trader is able to connect.

General flow: 
1. Trader sends candlesticks ([what is a candlestick?](https://www.investopedia.com/terms/c/candlestick.asp)) and the current price (tick) to the strategy
2. Optional: The strategy feeds indicators and retrieves the latest indicator values
3. The strategy decides if trader should open a new position and which position should be closed.
4. Trader executes new orders or closes open positions via broker APIs.

### environment overlays (eo) 

EOs can help to adjust a strategy according to the market volatility in order to reduce risk. E.g. you might want to buy when RSI indicator is below 25. However, if the market is getting dangerous due to high volatility you might only buy if RSI falls below 10 to ensure you buy only when indicator signals are stronger.

## backtesting

The `backtest` module can apply historical price simulations to your strategy.

You can configure tradings fees and spreads to simulate trading like with real money.

![Terminal output](docs/backtest-result.png)

It can also print a chart with a beautiful equity curve:

![Terminal output](docs/backtest-equity-curve.png)

You can use Coinbase and **histdata.com** prices for backtestings. Check [cmd/import-histdata](https://github.com/sklinkert/at/tree/master/cmd/import-histdata) for more.

## Contribution

Feel free to send PRs. I'm always happy to discuss and improve the code.

For hosting I can recommend DigitalOcean:

[![DigitalOcean Referral Badge](https://web-platforms.sfo2.digitaloceanspaces.com/WWW/Badge%203.svg)](https://www.digitalocean.com/?refcode=4a328aa341e2&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge)

## Donations

In case you want to show some love:

- BTC: 3BNnZUfw9qnLVnza9FvWF6n7tEXfWYVVy2
- ETH: 0xd30638F4fD54aeDB458d30504DD1cF2ce7563D36
- XMR: 45At7ezTicAejiLWTAfb28NNXnciH1M67VRrxLRHgfFyimHuPNP7MqbiUgYwwdTzXjbGFwCMsoMoH1Cvv7jPqKKANuaMpjo
