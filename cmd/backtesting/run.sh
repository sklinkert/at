#!/bin/bash

#INSTRUMENT="BTC-USD" STRATEGY="rsi" CANDLE_DURATION=24h YEAR_FROM=2021 MONTH_FROM=4 YEAR_TO=2021 MONTH_TO=12 PRICE_SOURCE="COINBASE" go run -ldflags="-w -s -X main.GitRev=123" ./cmd/backtesting/main.go

INSTRUMENT="SPXUSD" STRATEGY="rsi" CANDLE_DURATION=24h YEAR_FROM=2020 MONTH_FROM=1 YEAR_TO=2022 MONTH_TO=12 PRICE_DB_FILE=./data/SPXUSD.db go run -ldflags="-w -s -X main.GitRev=123" ./cmd/backtesting/main.go
