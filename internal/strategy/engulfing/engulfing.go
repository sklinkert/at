package engulfing

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/helper"
	"github.com/sklinkert/at/pkg/indicator/sma"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

// Long: Buy if candle closes below the last 7 candles and is above SMA 200
// Short: Short if candle closes above the last 7 candles and is below SMA 200
// Source: https://www.youtube.com/watch?v=_9Bmxylp63Y

type Engulfing struct {
	clog       *log.Entry
	instrument string
	sma        *sma.SMA
	//previousLows  ring.Ring
	//previousHighs ring.Ring
	ohlcPeriod    time.Duration
	locEST        *time.Location
	openPositions []broker.Position
	openOrders    []broker.Order
}

const (
	targetInPercent      = 10.0
	stopLossInPercent    = 10.0
	smaCandles           = 200
	strategyLongEnabled  = true
	strategyShortEnabled = true
)

func New(instrument string, candleDuration time.Duration) *Engulfing {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument, "CANDLE": candleDuration})

	locEST, err := time.LoadLocation("EST")
	if err != nil {
		clog.WithError(err).Fatal("time zone EST missing")
	}

	return &Engulfing{
		clog:       clog,
		instrument: instrument,
		sma:        sma.New(smaCandles),
		ohlcPeriod: candleDuration,
		locEST:     locEST,
	}
}

func (d *Engulfing) OnPosition(openPositions []broker.Position, _ []broker.Position) {
	d.openPositions = openPositions
}

func (d *Engulfing) OnOrder(openOrders []broker.Order) {
	d.openOrders = openOrders
}

func (d *Engulfing) GetWarmUpCandleAmount() uint {
	return smaCandles
}

func (d *Engulfing) OnWarmUpCandle(closedCandle *ohlc.OHLC) {
	d.feedIndicator(closedCandle)
}

func (d *Engulfing) feedIndicator(closedCandle *ohlc.OHLC) {
	d.sma.Insert(closedCandle)
}

func (d *Engulfing) GetCandleDuration() time.Duration {
	return d.ohlcPeriod
}

func (d *Engulfing) noTradingPeriod(now time.Time) bool {
	now = now.Local()

	// Weekend
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return true
	}

	// Avoid trading during US stocks markets (NYSE + NASDAQ) opening and closing
	//var estTime = now.In(d.locEST)
	//if estTime.Hour() == 9 || estTime.Hour() == 16 { // 09:30 - 16:00 EST
	//	return true
	//}

	return false
}

func (d *Engulfing) OnTick(_ tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	return
}

func (d *Engulfing) OnCandle(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Order, toClosePositions []broker.Position) {
	var closedCandle = closedCandles[len(closedCandles)-1]
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

// Rules:
// 1. close(-1) > open(-1)
// 2. close (0) < open(0)
// 3. open(0) > close(-1)
// 4. close(0) < open(-1)
func (d *Engulfing) isBearishEngulfingCandle(closedCandles []*ohlc.OHLC) bool {
	if len(closedCandles) < 2 {
		return false
	}

	currentCandle := closedCandles[len(closedCandles)-1]
	previousCandle := closedCandles[len(closedCandles)-2]

	// Rule 1
	if !previousCandle.Close.GreaterThan(previousCandle.Open) {
		return false
	}

	// Rule 2
	if !currentCandle.Close.LessThan(currentCandle.Open) {
		return false
	}

	// Rule 3
	if !currentCandle.Open.GreaterThan(previousCandle.Close) {
		return false
	}

	// Rule 4
	if !currentCandle.Close.LessThan(previousCandle.Open) {
		return false
	}

	return true
}

// Rules:
// 1. close(0) > open(0)
// 2. close(-1) < open(-1)
// 3. open(-1) > close(0)
// 4. close(-1) < open(0)
func (d *Engulfing) isBullishEngulfingCandle(closedCandles []*ohlc.OHLC) bool {
	if len(closedCandles) < 2 {
		return false
	}

	currentCandle := closedCandles[len(closedCandles)-1]
	previousCandle := closedCandles[len(closedCandles)-2]

	// Rule 1
	if !currentCandle.Close.GreaterThan(currentCandle.Open) {
		return false
	}

	// Rule 2
	if !previousCandle.Close.LessThan(previousCandle.Open) {
		return false
	}

	// Rule 3
	if !previousCandle.Open.GreaterThan(currentCandle.Close) {
		return false
	}

	// Rule 4
	if !previousCandle.Close.LessThan(currentCandle.Open) {
		return false
	}

	return true
}

func (d *Engulfing) strategyShort(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Position) {
	var closedCandle = closedCandles[len(closedCandles)-1]
	var closePrice = helper.DecimalToFloat(closedCandle.Close)
	var lowPrice = helper.DecimalToFloat(closedCandle.Low)

	if len(closedCandles) < 2 {
		return
	}

	var previousCandle = closedCandles[len(closedCandles)-2]

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
		if closedCandle.Close.LessThan(previousCandle.Close) {
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

	if closePrice > smaPrice && lowPrice > smaPrice {
		d.clog.Debugf("close is above SMA: %f < %f", closePrice, smaPrice)
		return
	}

	if d.isBullishEngulfingCandle(closedCandles) {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}

	return
}

func (d *Engulfing) strategyLong(closedCandles []*ohlc.OHLC) (toOpen []broker.Order, toClose []broker.Position) {
	var closedCandle = closedCandles[len(closedCandles)-1]
	var closePrice = helper.DecimalToFloat(closedCandle.Close)
	var lowPrice = helper.DecimalToFloat(closedCandle.Low)

	if len(closedCandles) < 2 {
		return
	}

	var previousCandle = closedCandles[len(closedCandles)-2]

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
		if closedCandle.Close.GreaterThan(previousCandle.Close) {
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

	if d.isBearishEngulfingCandle(closedCandles) {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}

	return
}

func (d *Engulfing) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
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

func (d *Engulfing) Name() string {
	return strategy.NameEngulfing
}

func (d *Engulfing) String() string {
	return fmt.Sprintf("%s: Long=%t, Short=%t Target=%.2f%% StopLoss=%.2f%% SMA%d", d.Name(),
		strategyLongEnabled, strategyShortEnabled, targetInPercent, stopLossInPercent, smaCandles)
}
