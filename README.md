# Automated Trader (at)

[![CI](https://github.com/sklinkert/at/actions/workflows/ci.yaml/badge.svg)](https://github.com/sklinkert/at/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sklinkert/at)](https://goreportcard.com/report/github.com/sklinkert/at)
[![Go Reference](https://pkg.go.dev/badge/github.com/sklinkert/at.svg)](https://pkg.go.dev/github.com/sklinkert/at)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sklinkert/at)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go framework for building, backtesting, and running automated trading strategies against real broker APIs. It works for stocks, forex, and crypto, and any broker can be plugged in behind a small interface.

**This is a framework, not a ready-made trading bot.** The bundled strategies are illustrative examples, not profitable systems.

```
 ticks / candles          decision            orders
 ───────────────▶  Strategy ────────▶  Trader ───────▶  Broker
    (Trader)         (your logic)      (this repo)     (IG, Coinbase, …)
```

## Why use it

- **Plug-in architecture** — brokers, strategies, and indicators are all small interfaces. Implement one, wire it in, done.
- **Backtest before you risk money** — replay historical prices with configurable spreads and trading fees, then print a performance summary and equity curve.
- **Paperwallet** — simulate fills for brokers without a sandbox, so you can dry-run against live prices.

## Quick start

Runs a full backtest with **no external data source or account** — a small sample EUR/USD dataset ships in `examples/sample-data/`.

```sh
# 1. Import the sample tick data into a local SQLite DB (creates ./data/EURUSD.db)
IMPORT_HISTDATA_CSV_FILES="examples/sample-data/EURUSD-2021-01.csv" \
INSTRUMENT="EURUSD" \
  go run ./cmd/import-histdata

# 2. Backtest the RSI strategy against it
PRICE_SOURCE="LOCAL_DB" \
PRICE_DB_FILE="./data/EURUSD.db" \
INSTRUMENT="EURUSD" \
STRATEGY="rsi" \
CANDLE_DURATION="1m" \
YEAR_FROM=2021 MONTH_FROM=1 YEAR_TO=2021 MONTH_TO=1 \
  go run ./cmd/backtesting
```

The backtest prints a trade-by-trade log and a summary (positions, win rate, performance in pips), writes `results/backtesting_result.csv`, and serves an interactive chart at `http://localhost:8080/chart`.

See [`.env.example`](.env.example) for every configuration variable and its default.

## Supported brokers

| Broker   | Demo account | Paperwallet trading | Real trading | Backtesting |
| -------- | :----------: | :-----------------: | :----------: | :---------: |
| IG.com   |      ✅      |         ❌          |      ✅      |     ✅      |
| Coinbase |      ❌      |         ✅          |      ❌      |     ✅      |

## Architecture

The `Trader` is the nerve center: it receives prices from a broker, forwards them to your strategy, and executes the orders the strategy returns.

1. The trader sends closed candles ([what is a candlestick?](https://www.investopedia.com/terms/c/candlestick.asp)) and the current tick to the strategy.
2. The strategy optionally feeds indicators and reads their latest values.
3. The strategy decides which positions to open and which to close.
4. The trader executes those orders and closes positions through the broker API.

### Packages

- **`internal/broker`** — the [`Broker`](internal/broker/broker.go) interface plus IG and Coinbase implementations. [`internal/paperwallet`](internal/paperwallet) simulates fills for brokers without a sandbox.
- **`internal/strategy`** — the [`Strategy`](internal/strategy/strategy.go) interface and example strategies: `rsi`, `rsiadx`, `sma10`, `stochrsi`, `doji`, `engulfing`, `harami`, `lowcandle`, `scalper`, `heikinashi`.
- **`pkg/indicator`** — the [`Indicator`](pkg/indicator/indicator.go) interface and implementations (SMA, RSI, ADX, Stoch, StochRSI) wrapping [go-talib](https://github.com/markcheno/go-talib).
- **`pkg/eo`** — environment overlays that adapt a strategy to market volatility (e.g. require a stronger signal in dangerous conditions).
- **`pkg/chart`** — renders the equity curve and price chart as HTML.

New to the code? Start with the [`Strategy`](internal/strategy/strategy.go) interface, then read [`internal/strategy/rsi`](internal/strategy/rsi) as a worked example.

![Overview](docs/overview.png)

## Backtesting

The backtest module replays historical prices through your strategy. Configure trading fees and spreads to approximate real conditions.

![Terminal output](docs/backtest-result.png)

It can also render an equity curve:

![Equity curve](docs/backtest-equity-curve.png)

Price data comes from a local SQLite database. Import your own [histdata.com](https://www.histdata.com/) CSV files with [`cmd/import-histdata`](cmd/import-histdata) — the same tool used in the quick start.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for how to build, test, lint, and add a new strategy or indicator.

## Disclaimer

The developers are not liable for any losses arising from buying or selling securities. All included strategies are examples and are in no case ready trading systems. Trade at your own risk.

## License

[MIT](LICENSE)
