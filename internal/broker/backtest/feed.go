package backtest

import (
	"context"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/igmarkets"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type QuotesSource int

const (
	QuotesSourceSqlite = iota + 1
	QuotesSourceYahooFinance
	QuotesSourceIGMarkets
	QuotesSourcePatternTrading
)

func (b *Backtest) retrieveCandlesFromIGMarkets(receiver chan ohlc.OHLC) {
	defer close(receiver)

	var ctx = context.Background()

	if err := b.brokerIGMarkets.Login(ctx); err != nil {
		log.WithError(err).Fatal("login failed")
	}

	priceResponse, err := b.brokerIGMarkets.GetPriceHistory(ctx, b.instrument, igmarkets.ResolutionHour, 100, b.periodFrom, b.periodTo)
	if err != nil {
		log.WithError(err).Fatalf("failed to fetch price history for %q from IG Markets", b.instrument)
	}
	log.Infof("prices fetched: %d", len(priceResponse.Prices))

	for _, price := range priceResponse.Prices {
		open := bidAskToTick(b.instrument, price.SnapshotTimeUTCParsed, price.OpenPrice.Bid, price.OpenPrice.Ask)
		high := bidAskToTick(b.instrument, price.SnapshotTimeUTCParsed, price.HighPrice.Bid, price.HighPrice.Ask)
		low := bidAskToTick(b.instrument, price.SnapshotTimeUTCParsed, price.LowPrice.Bid, price.LowPrice.Ask)
		close := bidAskToTick(b.instrument, price.SnapshotTimeUTCParsed, price.ClosePrice.Bid, price.ClosePrice.Ask)

		var candle = ohlc.OHLC{
			Instrument: b.instrument,
			Open:       open.Price(),
			High:       high.Price(),
			Low:        low.Price(),
			Close:      close.Price(),
			Start:      price.SnapshotTimeUTCParsed,
			End:        price.SnapshotTimeUTCParsed.Add(b.priceDBCandleDuration),
		}
		candle.ForceClose()
		log.Infof("Candle: %+v", candle)
		receiver <- candle
	}
}

func bidAskToTick(instrument string, datetime time.Time, bid, ask float64) tick.Tick {
	return tick.New(instrument, datetime, decimal.NewFromFloat(bid), decimal.NewFromFloat(ask))
}

func (b *Backtest) retrieveCandlesFromYahooFinance(receiver chan ohlc.OHLC) {
	defer close(receiver)

	params := &chart.Params{
		Symbol:   b.instrument,
		Interval: datetime.OneDay,
		Start:    datetime.New(&b.periodFrom),
		End:      datetime.New(&b.periodTo),
	}

	log.Infof("Fetching quotes from Yahoo Finance for %q with period %s - %s",
		b.instrument, b.periodFrom, b.periodTo)
	iter := chart.Get(params)

	for iter.Next() {
		openTime := time.Unix(int64(iter.Bar().Timestamp), 0)
		bar := iter.Bar()

		candle := ohlc.OHLC{
			Instrument: b.instrument,
			Open:       bar.Open,
			High:       bar.High,
			Low:        bar.Low,
			Close:      bar.Close,
			Start:      openTime,
			End:        openTime.Add(b.priceDBCandleDuration),
		}
		candle.ForceClose()
		log.Infof("Candle: %+v", candle)

		receiver <- candle
	}
	if err := iter.Err(); err != nil {
		log.WithError(err).Fatal("getting quotes from yahoo failed")
	}
}

func (b *Backtest) retrieveCandlesFromSQLite(receiver chan ohlc.OHLC) {
	defer close(receiver)

	db, err := gorm.Open(sqlite.Open(b.priceDBFile), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatalf("failed to connect database %q", b.priceDBFile)
	}

	// Speed up read performance
	db.Exec("PRAGMA locking_mode = EXCLUSIVE;")

	const pageSize = 80000
	var offset int
	for {
		var candles []ohlc.OHLC
		if err := db.
			Offset(offset).
			Limit(pageSize).
			Order("start").
			Where("duration = ? AND start BETWEEN ? AND ?", b.priceDBCandleDuration, b.periodFrom, b.periodTo).
			Find(&candles).Error; err != nil {
			log.WithError(err).Error("db.Find(&candles) failed")
			return
		}
		if len(candles) == 0 {
			log.Info("No more candles fetched")
			return
		}
		for _, candle := range candles {
			receiver <- candle
		}
		offset += pageSize
	}
}

func (b *Backtest) ListenToPriceFeed(traderChan chan tick.Tick) {
	var c = make(chan ohlc.OHLC)

	switch b.quotesSource {
	case QuotesSourceSqlite:
		go b.retrieveCandlesFromSQLite(c)
	case QuotesSourceYahooFinance:
		go b.retrieveCandlesFromYahooFinance(c)
	case QuotesSourceIGMarkets:
		go b.retrieveCandlesFromIGMarkets(c)
	case QuotesSourcePatternTrading:
		go b.retrieveCandlesFromPatternTrading(c)
	default:
		log.Fatalf("Unknown quotes source: %d", b.quotesSource)
	}

	for candle := range c {
		for _, currentTick := range candle.ToTicks() {
			b.paperwallet.SetCurrenctPrice(currentTick)
			traderChan <- currentTick
		}
	}
	b.paperwallet.CloseAllOpenPositions()
	b.writeCSV()
	b.paperwallet.PrintSummary()
}
