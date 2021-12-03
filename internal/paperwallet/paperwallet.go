package paperwallet

import (
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/tick"
	"sync"
)

type Option func(paperwallet *Paperwallet)

type Paperwallet struct {
	initialBalance  decimal.Decimal
	balance         decimal.Decimal
	openPositions   map[string]broker.Position
	closedPositions map[string]broker.Position

	tradingFeePercent decimal.Decimal
	totalTradingFee   decimal.Decimal
	spreadInCents     decimal.Decimal
	slippageAbsolute  decimal.Decimal
	currentTick       tick.Tick
	trailingStops     map[string]TrailingStop
	sync.RWMutex
}

type TrailingStop struct {
	StopDistance        decimal.Decimal
	IncrementSizeInPips decimal.Decimal
}

func WithInitialBalance(balance decimal.Decimal) Option {
	return func(paperwallet *Paperwallet) {
		paperwallet.Lock()
		paperwallet.initialBalance = balance
		paperwallet.balance = balance
		paperwallet.Unlock()
	}
}

// WithSpread - additional bid/ask spread
func WithSpread(spreadInCents decimal.Decimal) Option {
	return func(paperwallet *Paperwallet) {
		paperwallet.spreadInCents = spreadInCents
	}
}

// WithTradingFeePercent - fee that is added to bid/ask
func WithTradingFeePercent(feePercent decimal.Decimal) Option {
	return func(paperwallet *Paperwallet) {
		paperwallet.tradingFeePercent = feePercent
	}
}

// WithSlippage - slippage that is added to the price when buying/selling.
// Makes it a bit more disadvantageous but also more realistic
func WithSlippage(slippageAbsolute decimal.Decimal) Option {
	return func(paperwallet *Paperwallet) {
		paperwallet.slippageAbsolute = slippageAbsolute
	}
}

func New(options ...Option) *Paperwallet {
	const defaultBalance = 1000

	pw := &Paperwallet{
		initialBalance:  decimal.NewFromFloat(defaultBalance),
		balance:         decimal.NewFromFloat(defaultBalance),
		openPositions:   map[string]broker.Position{},
		closedPositions: map[string]broker.Position{},
		trailingStops:   map[string]TrailingStop{},
	}

	for _, option := range options {
		option(pw)
	}

	return pw
}

func (pw *Paperwallet) GetInitialBalance() decimal.Decimal {
	return pw.initialBalance
}

func (pw *Paperwallet) GetBalance() decimal.Decimal {
	return pw.balance
}
