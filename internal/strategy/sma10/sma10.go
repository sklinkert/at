package sma10

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

type SMA struct {
	clog       *log.Entry
	instrument string
	sma        *sma.SMA
	sma10      *sma.SMA
	ohlcPeriod time.Duration
	locEST     *time.Location
}

const (
	targetInPercent      = 2.0
	stopLossInPercent    = 0.5
	smaCandles           = 200
	strategyLongEnabled  = true
	strategyShortEnabled = true
)

func New(instrument string, candleDuration time.Duration) *SMA {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument, "CANDLE": candleDuration})

	locEST, err := time.LoadLocation("EST")
	if err != nil {
		clog.WithError(err).Fatal("time zone EST missing")
	}

	return &SMA{
		clog:       clog,
		instrument: instrument,
		sma:        sma.New(smaCandles),
		sma10:      sma.New(10),
		ohlcPeriod: candleDuration,
		locEST:     locEST,
	}
}

func (d *SMA) GetWarmUpCandleAmount() uint {
	return smaCandles
}

func (d *SMA) OnWarmUpCandle(closedCandle *ohlc.OHLC) {
	d.feedIndicator(closedCandle)
}

func (d *SMA) feedIndicator(closedCandle *ohlc.OHLC) {
	d.sma.Insert(closedCandle)
	d.sma10.Insert(closedCandle)
}

func (d *SMA) GetCandleDuration() time.Duration {
	return d.ohlcPeriod
}

func (d *SMA) noTradingPeriod(now time.Time) bool {
	now = now.Local()

	// Weekend
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return true
	}

	// Avoid trading during US stocks markets (NYSE + NASDAQ) opening and closing
	var estTime = now.In(d.locEST)
	if estTime.Hour() == 9 || estTime.Hour() == 16 { // 09:30 - 16:00 EST
		return true
	}

	return false
}

func (d *SMA) OnCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openOrders []broker.Order, openPositions []broker.Position, closedPositions []broker.Position) (toOpen []broker.Order, toCloseOrderIDs []string, toClosePositions []broker.Position) {
	defer d.feedIndicator(closedCandle)

	if strategyLongEnabled {
		toOpenLong, toCloseLong := d.strategyLong(closedCandle, closedCandles, currentTick, openPositions, closedPositions)
		toOpen = append(toOpen, toOpenLong...)
		toClosePositions = append(toClosePositions, toCloseLong...)
	}
	if strategyShortEnabled {
		toOpenShort, toCloseShort := d.strategyShort(closedCandle, closedCandles, currentTick, openPositions, closedPositions)
		toOpen = append(toOpen, toOpenShort...)
		toClosePositions = append(toClosePositions, toCloseShort...)
	}
	return
}

func (d *SMA) strategyLong(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toClose []broker.Position) {
	var closePrice = helper.DecimalToFloat(closedCandle.Close)

	smaValue, err := d.sma.Value()
	if err != nil {
		log.WithError(err).Warn("No SMA")
		return
	}
	sma200Price := smaValue[sma.Value]

	sma10, err := d.sma10.Value()
	if err != nil {
		return
	}
	sma10Price := sma10[sma.Value]

	for _, openPosition := range openPositions {
		if openPosition.BuyDirection != broker.BuyDirectionLong {
			continue
		}
		if helper.DecimalToFloat(closedCandle.Close) > sma10Price {
			toClose = openPositions
			return
		}
	}
	if len(openPositions) > 0 {
		return
	}

	if d.noTradingPeriod(currentTick.Datetime) {
		d.clog.Info("No trading period")
		return
	}

	if closePrice < sma200Price {
		d.clog.Debugf("close is below SMA: %f < %f", closePrice, sma200Price)
		return
	}

	if closePrice < sma10Price {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}
	d.clog.Debugf("long: closePrice >= SMA10 : %f >= %f", closePrice, sma10Price)
	return
}

func (d *SMA) strategyShort(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toClose []broker.Position) {
	var closePrice = helper.DecimalToFloat(closedCandle.Close)

	smaValue, err := d.sma.Value()
	if err != nil {
		log.WithError(err).Warn("No SMA")
		return
	}
	sma200Price := smaValue[sma.Value]

	sma10, err := d.sma10.Value()
	if err != nil {
		return
	}
	sma10Price := sma10[sma.Value]

	for _, openPosition := range openPositions {
		if openPosition.BuyDirection != broker.BuyDirectionShort {
			continue
		}
		if helper.DecimalToFloat(closedCandle.Close) < sma10Price {
			toClose = openPositions
			return
		}
	}
	if len(openPositions) > 0 {
		return
	}

	if d.noTradingPeriod(currentTick.Datetime) {
		d.clog.Info("No trading period")
		return
	}

	if closePrice > sma200Price {
		d.clog.Debugf("close is below SMA: %f < %f", closePrice, sma200Price)
		return
	}

	if closePrice > sma10Price {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []broker.Position{}
	}
	d.clog.Debugf("short: closePrice <= SMA10 : %f <= %f", closePrice, sma10Price)
	return
}

func (d *SMA) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
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

func (d *SMA) Name() string {
	return strategy.NameSMA10
}

func (d *SMA) String() string {
	return fmt.Sprintf("%s: Long=%t, Short=%t Target=%.2f%% StopLoss=%.2f%% SMA%d", d.Name(),
		strategyLongEnabled, strategyShortEnabled, targetInPercent, stopLossInPercent, smaCandles)
}
