package harami

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/helper"
	"github.com/sklinkert/at/pkg/indicator/sma"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/circularbuffer"
	"time"
)

// Buy if current closed candle is a bullish harami candle and market is still above SMA 200
// Source: https://www.youtube.com/watch?v=_9Bmxylp63Y

type Harami struct {
	clog          *log.Entry
	instrument    string
	sma           *sma.SMA
	previousLows  *circularbuffer.CircularBuffer
	previousHighs *circularbuffer.CircularBuffer
	ohlcPeriod    time.Duration
}

const (
	targetInPercent   = 100.0
	stopLossInPercent = 0.5
	smaCandles        = 200
)

func New(instrument string, candleDuration time.Duration) *Harami {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument, "CANDLE": candleDuration})

	return &Harami{
		clog:          clog,
		instrument:    instrument,
		sma:           sma.New(smaCandles),
		previousLows:  circularbuffer.New(7, 7),
		previousHighs: circularbuffer.New(7, 7),
		ohlcPeriod:    candleDuration,
	}
}

func (h *Harami) GetWarmUpCandleAmount() uint {
	return smaCandles
}

func (h *Harami) ProcessWarmUpCandle(closedCandle *ohlc.OHLC) {
	h.feedIndicator(closedCandle)
}

func (h *Harami) feedIndicator(closedCandle *ohlc.OHLC) {
	var high = helper.DecimalToFloat(closedCandle.High)
	var low = helper.DecimalToFloat(closedCandle.Low)
	h.sma.Insert(closedCandle)
	h.previousLows.Insert(low)
	h.previousHighs.Insert(high)
}

func (h *Harami) GetCandleDuration() time.Duration {
	return h.ohlcPeriod
}

func (h *Harami) isBearishCandle(candle *ohlc.OHLC) bool {
	return candle.Close.LessThan(candle.Open)
}

func (h *Harami) isBullishCandle(candle *ohlc.OHLC) bool {
	return candle.Close.GreaterThan(candle.Open)
}

func (h *Harami) isHaramiLong(firstCandle, secondCandle *ohlc.OHLC) bool {
	if h.isBearishCandle(firstCandle) && h.isBullishCandle(secondCandle) &&
		secondCandle.High.LessThan(firstCandle.Open) &&
		secondCandle.Low.GreaterThan(firstCandle.Close) {
		return true
	}
	return false
}

func (h *Harami) ProcessCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, _ tick.Tick, openOrders []broker.Order, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toCloseOrderIDs []string, toClosePositions []broker.Position) {
	var closePrice = helper.DecimalToFloat(closedCandle.Close)

	defer h.feedIndicator(closedCandle)

	smaValue, err := h.sma.Value()
	if err != nil {
		log.WithError(err).Warn("No SMA")
		return
	}
	smaPrice := smaValue[sma.Value]

	if len(openPositions) > 0 {
		if closePrice < smaPrice {
			toClosePositions = openPositions
			return
		}

		previousCandlesHigh, err := h.previousHighs.Max()
		if err != nil {
			return
		}
		if helper.DecimalToFloat(closedCandle.Close) > previousCandlesHigh {
			toClosePositions = openPositions
			return
		}
		return
	}

	latestCandle := closedCandles[len(closedCandles)-1]
	secondLatestCandle := closedCandles[len(closedCandles)-2]
	if h.isHaramiLong(secondLatestCandle, latestCandle) {
		toOpenNew := h.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []string{}, []broker.Position{}
	}

	h.clog.Debugf("No harami long candle found: %s", closedCandle)
	return
}

func (h *Harami) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
	var (
		targetPrice   = helper.CalcTargetPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(targetInPercent), direction)
		stopLossPrice = helper.CalcStopLossPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(stopLossInPercent), direction)
	)

	h.clog.WithFields(log.Fields{
		"Direction": direction.String(),
		"Time":      closedCandle.End,
		"Close":     closedCandle.Close,
		"Target":    targetInPercent,
		"StopLoss":  stopLossPrice,
	}).Debug("Prepare new order")

	return broker.NewMarketOrder(direction, size, h.instrument, targetPrice, stopLossPrice)
}

func (h *Harami) Name() string {
	return strategy.NameHarami
}

func (h *Harami) String() string {
	return fmt.Sprintf("%s: Target=%.2f StopLoss=%.2f", h.Name(), targetInPercent, stopLossInPercent)
}
