package lowcandle

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

// Long: Buy if candle closes below the last 7 candles and is above SMA 200
// Short: Short if candle closes above the last 7 candles and is below SMA 200
// Source: https://www.youtube.com/watch?v=_9Bmxylp63Y

type LowCandle struct {
	clog          *log.Entry
	instrument    string
	sma           *sma.SMA
	previousLows  *circularbuffer.CircularBuffer
	previousHighs *circularbuffer.CircularBuffer
	ohlcPeriod    time.Duration
	locEST        *time.Location
	openPositions []broker.Position
	openOrders    []broker.Order
}

const (
	targetInPercent      = 4.0
	stopLossInPercent    = 0.5
	smaCandles           = 200
	strategyLongEnabled  = true
	strategyShortEnabled = true
)

func New(instrument string, candleDuration time.Duration) *LowCandle {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument, "CANDLE": candleDuration})

	locEST, err := time.LoadLocation("EST")
	if err != nil {
		clog.WithError(err).Fatal("time zone EST missing")
	}

	return &LowCandle{
		clog:          clog,
		instrument:    instrument,
		sma:           sma.New(smaCandles),
		previousLows:  circularbuffer.New(7, 7),
		previousHighs: circularbuffer.New(7, 7),
		ohlcPeriod:    candleDuration,
		locEST:        locEST,
	}
}

func (d *LowCandle) OnPosition(openPositions []broker.Position, _ []broker.Position) {
	d.openPositions = openPositions
}

func (d *LowCandle) OnOrder(openOrders []broker.Order) {
	d.openOrders = openOrders
}

func (d *LowCandle) GetWarmUpCandleAmount() uint {
	return smaCandles
}

func (d *LowCandle) OnWarmUpCandle(closedCandle *ohlc.OHLC) {
	d.feedIndicator(closedCandle)
}

func (d *LowCandle) feedIndicator(closedCandle *ohlc.OHLC) {
	var high = helper.DecimalToFloat(closedCandle.High)
	var low = helper.DecimalToFloat(closedCandle.Low)
	d.sma.Insert(closedCandle)
	d.previousLows.Insert(low)
	d.previousHighs.Insert(high)
}

func (d *LowCandle) GetCandleDuration() time.Duration {
	return d.ohlcPeriod
}

func (d *LowCandle) noTradingPeriod(now time.Time) bool {
	now = now.Local()

	// Weekend
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return true
	}

	// Avoid trading during night hours to improve backtesting matching.
	if now.Hour() > 20 && now.Hour() < 4 {
		return false
	}

	// Avoid trading during US stocks markets (NYSE + NASDAQ) opening and closing
	var estTime = now.In(d.locEST)
	if estTime.Hour() == 9 || estTime.Hour() == 16 { // 09:30 - 16:00 EST
		return true
	}

	return false
}

func (d *LowCandle) OnTick(_ tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	return
}

func (d *LowCandle) OnCandle(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Order, toClosePositions []broker.Position) {
	closedCandle := closedCandles[len(closedCandles)-1]
	defer d.feedIndicator(closedCandle)

	if strategyLongEnabled {
		toOpenLong, toCloseLong := d.strategyLong(closedCandles)
		toOpen = append(toOpen, toOpenLong...)
		toClosePositions = append(toClosePositions, toCloseLong...)
	}
	if strategyShortEnabled {
		toOpenShort, toCloseShort := d.strategyShort(closedCandles)
		toOpen = append(toOpen, toOpenShort...)
		toClosePositions = append(toClosePositions, toCloseShort...)
	}
	return
}

func (d *LowCandle) strategyLong(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Position) {
	var closedCandle = closedCandles[len(closedCandles)-1]
	var closePrice = helper.DecimalToFloat(closedCandle.Close)
	var lowPrice = helper.DecimalToFloat(closedCandle.Low)

	smaValue, err := d.sma.Value()
	if err != nil {
		log.WithError(err).Warn("No SMA")
		return
	}
	smaPrice := smaValue[sma.Value]

	for _, openPosition := range d.openPositions {
		if openPosition.BuyDirection != broker.BuyDirectionLong {
			continue
		}
		previousCandlesHigh, err := d.previousHighs.Max()
		if err != nil {
			return
		}
		if helper.DecimalToFloat(closedCandle.Close) > previousCandlesHigh {
			toClose = d.openPositions
			return
		}
	}
	if len(d.openPositions) > 0 {
		return
	}

	if d.noTradingPeriod(closedCandle.End) {
		d.clog.Info("No trading period")
		return
	}

	if closePrice < smaPrice && lowPrice < smaPrice {
		d.clog.Debugf("close is below SMA: %f < %f", closePrice, smaPrice)
		return
	}

	previousCandlesLow, err := d.previousLows.Min()
	if err != nil {
		d.clog.WithError(err).Warn("no previous low")
		return
	}

	if closePrice < previousCandlesLow {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}
	d.clog.Debugf("long: closePrice >= previousCandlesLow : %f > %f", closePrice, previousCandlesLow)
	return
}

func (d *LowCandle) strategyShort(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Position) {
	var closedCandle = closedCandles[len(closedCandles)-1]
	var closePrice = helper.DecimalToFloat(closedCandle.Close)
	var highPrice = helper.DecimalToFloat(closedCandle.High)

	defer d.feedIndicator(closedCandle)

	smaValue, err := d.sma.Value()
	if err != nil {
		log.WithError(err).Warn("No SMA")
		return
	}
	smaPrice := smaValue[sma.Value]

	for _, openPosition := range d.openPositions {
		if openPosition.BuyDirection != broker.BuyDirectionShort {
			continue
		}
		previousCandlesLow, err := d.previousLows.Min()
		if err != nil {
			return
		}
		if helper.DecimalToFloat(closedCandle.Close) < previousCandlesLow {
			toClose = append(toClose, openPosition)
			return
		}
	}
	if len(d.openPositions) > 0 {
		return
	}

	if d.noTradingPeriod(closedCandle.End) {
		d.clog.Info("No trading period")
		return
	}

	if closePrice > smaPrice && highPrice > smaPrice {
		d.clog.Debugf("close is below SMA: %f < %f", closePrice, smaPrice)
		return
	}

	previousCandlesHigh, err := d.previousHighs.Max()
	if err != nil {
		d.clog.WithError(err).Warn("no previous low")
		return
	}

	if closePrice > previousCandlesHigh {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}
	d.clog.Debugf("short: closePrice <= previousCandlesHigh : %f > %f", closePrice, previousCandlesHigh)
	return
}

func (d *LowCandle) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
	var (
		targetPrice   = helper.CalcTargetPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(targetInPercent), direction)
		stopLossPrice = helper.CalcStopLossPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(stopLossInPercent), direction)
	)

	d.clog.WithFields(log.Fields{
		"Direction": direction.String(),
		"Time":      closedCandle.End,
		"Close":     closedCandle.Close,
		"Target":    targetInPercent,
		"StopLoss":  stopLossPrice,
	}).Debug("Prepare new order")

	return broker.NewMarketOrder(direction, size, d.instrument, targetPrice, stopLossPrice)
}

func (d *LowCandle) Name() string {
	return strategy.NameLowCandle
}

func (d *LowCandle) String() string {
	return fmt.Sprintf("%s: Long=%t, Short=%t Target=%.2f%% StopLoss=%.2f%% SMA%d", d.Name(),
		strategyLongEnabled, strategyShortEnabled, targetInPercent, stopLossInPercent, smaCandles)
}
