package backtest

import (
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/ohlc"
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

func (b *Backtest) retrieveCandlesFromCoinbase(receiver chan ohlc.OHLC) {
	defer close(receiver)

	client := coinbasepro.NewClient()
	params := coinbasepro.GetHistoricRatesParams{
		Start:       b.periodFrom,
		End:         b.periodTo,
		Granularity: int(b.candlePeriod.Seconds()),
	}
	historicRates, err := client.GetHistoricRates(b.instrument, params)
	if err != nil {
		log.WithError(err).Fatalf("cannot fetch historic rates from coinbase (params: %+v)", params)
	}

	for _, historicRate := range historicRates {
		openPrice := decimal.NewFromFloat(historicRate.Open)
		highPrice := decimal.NewFromFloat(historicRate.High)
		lowPrice := decimal.NewFromFloat(historicRate.Low)
		closePrice := decimal.NewFromFloat(historicRate.Close)
		candle := ohlc.New(b.instrument, historicRate.Time, b.candlePeriod, false)

		candle.NewPrice(openPrice, historicRate.Time)
		candle.NewPrice(highPrice, historicRate.Time)
		candle.NewPrice(lowPrice, historicRate.Time)
		candle.NewPrice(closePrice, historicRate.Time)
		candle.ForceClose()

		receiver <- *candle
	}

	log.Infof("Processed %d candles", len(historicRates))
}
