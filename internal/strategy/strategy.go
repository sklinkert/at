package strategy

import (
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

const (
	NameMeanReversion = "meanreversion"
	NameDOJI          = "doji"
	NameHeikinAshi    = "heikinashi"
	NameMock          = "mock"
	NameScalper       = "scalper"
	NameStochRSI      = "stochrsi"
	NameRSI           = "rsi"
	NameRSIADX        = "rsiadx"
	NameLowCandle     = "lowcandle"
	NameHarami        = "harami"
	NameSMA10         = "sma10"
	NameEngulfing     = "engulfing"
	NameHighestHigh   = "marektstructure"
)

type Strategy interface {
	// ProcessWarmUpCandle sends a closed candle from database to strategy for warming up indicators etc.
	ProcessWarmUpCandle(closedCandle *ohlc.OHLC)

	// GetWarmUpCandleAmount tells the trader how many warmup candles are required from database
	GetWarmUpCandleAmount() uint

	// ProcessCandle send the candle right after it has been closed.
	// closedCandles contains the 100 most recent candles.
	ProcessCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openPositions []broker.Position, closedPositions []broker.Position) (toOpen []broker.Order, toClose []broker.Position)

	// GetCandleDuration - Returns the durations for all candles required by a strategy.
	GetCandleDuration() time.Duration

	// Name returns the name of the strategy
	Name() string

	// String explains the strategy settings, e.g. stop loss, target, etc.
	String() string
}
