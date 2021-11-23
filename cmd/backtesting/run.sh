#!/bin/bash

INSTRUMENT="CB.ETHUSD" STRATEGY="rsiadx" CANDLE_DURATION=15m YEAR_FROM=2014 YEAR_TO=2022 PRICE_SOURCE="PATTERN_TRADING" go run  -ldflags="-w -s -X main.GitRev=123" ./cmd/backtesting/main.go