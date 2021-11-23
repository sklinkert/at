package backtest

import (
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/paperwallet"
	"github.com/sklinkert/igmarkets"
	"sync"
	"time"
)

// Backtest contains all required data for running a backtesting.
type Backtest struct {
	instrument      string
	periodFrom      time.Time
	periodTo        time.Time
	quotesSource    QuotesSource
	brokerIGMarkets *igmarkets.IGMarkets
	candlePeriod    time.Duration
	paperwallet     *paperwallet.Paperwallet

	// Price data sqlite file
	priceDBFile           string
	priceDBCandleDuration time.Duration

	// Read raw data from CSV files
	tickDataFiles []string
	sync.RWMutex
}

type Option func(backtest *Backtest)

func WithPriceDBFile(dbFile string, candleDuration time.Duration) Option {
	return func(backtest *Backtest) {
		backtest.priceDBFile = dbFile
		backtest.priceDBCandleDuration = candleDuration
	}
}

func WithTickDataFiles(files []string) Option {
	return func(backtest *Backtest) {
		backtest.tickDataFiles = files
	}
}

func WithQuotesSource(quotesSource QuotesSource) Option {
	return func(backtest *Backtest) {
		backtest.quotesSource = quotesSource
	}
}

func WithQuotesSourceIGMarkets(igBroker *igmarkets.IGMarkets) Option {
	return func(backtest *Backtest) {
		backtest.brokerIGMarkets = igBroker
	}
}

func WithCandlePeriod(period time.Duration) Option {
	return func(backtest *Backtest) {
		backtest.candlePeriod = period
	}
}

// New creates new backtesting instance
func New(instrument string, periodFrom, periodTo time.Time, paperwallet *paperwallet.Paperwallet, options ...Option) *Backtest {
	var b = &Backtest{
		instrument:  instrument,
		periodFrom:  periodFrom,
		periodTo:    periodTo,
		paperwallet: paperwallet,
	}

	for _, option := range options {
		option(b)
	}

	return b
}

// Buy open new position with target and stop loss
func (b *Backtest) Buy(order broker.Order) (broker.Position, error) {
	b.Lock()
	defer b.Unlock()

	position, err := b.paperwallet.Buy(order)

	log.WithFields(log.Fields{
		"Error":     err,
		"BuyTime":   position.BuyTime,
		"Reference": position.Reference,
		"Size":      order.Size,
	}).Debug("New position")

	return position, err
}

// Sell closes the given open position
func (b *Backtest) Sell(position broker.Position) error {
	b.Lock()
	defer b.Unlock()

	return b.paperwallet.Sell(position)
}
