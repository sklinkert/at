package backtest

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/ohlc"
	"io"
	"net/http"
	"strings"
	"time"
)

type Candlestick struct {
	Isin      string
	Period    string
	Exchange  string
	OpenTime  time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	CloseTime time.Time
}

func (b *Backtest) splitInstrument() (exchange, isin string) {
	var parts = strings.Split(b.instrument, ".")
	return parts[0], parts[1]
}

func (b *Backtest) retrieveCandlesFromPatternTrading(receiver chan ohlc.OHLC) {
	defer close(receiver)

	exchange, isin := b.splitInstrument()
	var url = fmt.Sprintf("https://api.pattern-trading.com/api/v1/candlesticks/%s/%s/%s", exchange, isin, b.candlePeriod)

	var httpClient = http.Client{}
	resp, err := httpClient.Get(url)
	if err != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var candlesticks []Candlestick
	err = json.Unmarshal(body, &candlesticks)
	if err != nil {
		return
	}

	for _, candlestick := range candlesticks {
		candle := ohlc.New(b.instrument, candlestick.OpenTime, b.candlePeriod, false)
		candle.NewPrice(candlestick.Open, candlestick.OpenTime)
		candle.NewPrice(candlestick.Low, candlestick.OpenTime)
		candle.NewPrice(candlestick.High, candlestick.OpenTime)
		candle.NewPrice(candlestick.Close, candlestick.CloseTime)
		candle.ForceClose()
		receiver <- *candle
	}

	log.Infof("Processed %d candles from Pattern-Trading.com", len(candlesticks))
}
