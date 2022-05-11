package strategy

import (
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

const (
	NameDOJI       = "doji"
	NameHeikinAshi = "heikinashi"
	NameScalper    = "scalper"
	NameStochRSI   = "stochrsi"
	NameRSI        = "rsi"
	NameRSIADX     = "rsiadx"
	NameLowCandle  = "lowcandle"
	NameHarami     = "harami"
	NameSMA10      = "sma10"
	NameEngulfing  = "engulfing"
)

type Strategy interface {
	// Name returns the name of the strategy
	Name() string

	// OnCandle is processing a list of closed candles. Will be called right after a new candle has been closed.
	// closedCandles contains the 100 most recent candles.
	OnCandle(closedCandles []*ohlc.OHLC) (toOpen, toClose []broker.Order, toClosePositions []broker.Position)

	// OnTick is processing a new tick. Will be called right after a new tick has been received.
	OnTick(currentTick tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position)

	// OnPosition is processing a new position. Will be called right after a new position has been opened or closed.
	OnPosition(openPositions []broker.Position, closedPositions []broker.Position)

	// OnOrder is processing a new order. Will be called right after a new order has been opened or closed.
	OnOrder(openOrders []broker.Order)

	// OnWarmUpCandle sends a closed candle from database to strategy for e.g. warming up indicators.
	OnWarmUpCandle(closedCandle *ohlc.OHLC)

	// GetWarmUpCandleAmount tells the trader how many warmup candles are required from database. No guarantee that
	// the strategy will receive the requested amount.
	GetWarmUpCandleAmount() uint

	// GetCandleDuration - Returns the durations for all candles required by a strategy.
	GetCandleDuration() time.Duration

	// String explains the strategy settings, e.g. stop loss, target, etc.
	String() string
}
