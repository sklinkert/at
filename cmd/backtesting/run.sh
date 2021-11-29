#!/bin/bash

INSTRUMENT="BTC-USD" STRATEGY="rsi" CANDLE_DURATION=24h YEAR_FROM=2021 MONTH_FROM=4 YEAR_TO=2021 MONTH_TO=12 PRICE_SOURCE="COINBASE" go run -ldflags="-w -s -X main.GitRev=123" ./cmd/backtesting/main.go