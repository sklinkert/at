package stochrsi

import (
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/helper"
	"github.com/sklinkert/at/pkg/indicator/stochrsi"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

type RSI struct {
	clog          *log.Entry
	instrument    string
	rsi           *stochrsi.StochRSI
	openPositions []broker.Position
	openOrders    []broker.Order
}

const (
	ohlcPeriod        = time.Minute * 60
	upperThreshold    = 90
	lowerThreshold    = 10
	targetInPercent   = 4.0
	stopLossInPercent = 0.5
)

func New(instrument string) *RSI {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})

	return &RSI{
		clog:       clog,
		instrument: instrument,
		rsi:        stochrsi.New(5, 2, 14),
	}
}

func (d *RSI) OnPosition(openPositions []broker.Position, _ []broker.Position) {
	d.openPositions = openPositions
}

func (d *RSI) OnTick(_ tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	return
}

func (d *RSI) OnOrder(openOrders []broker.Order) {
	d.openOrders = openOrders
}

func (d *RSI) GetCandleDuration() time.Duration {
	return ohlcPeriod
}

func (d *RSI) OnWarmUpCandle(_ *ohlc.OHLC) {}

func (d *RSI) GetWarmUpCandleAmount() uint {
	return 1
}

func (d *RSI) OnCandle(closedCandles []*ohlc.OHLC) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	closedCandle := closedCandles[len(closedCandles)-1]

	d.rsi.Insert(closedCandle)
	if len(d.openPositions) > 0 {
		return
	}

	// No night trading
	if closedCandle.End.Hour() < 10 || closedCandle.End.Hour() > 20 {
		return
	}

	rsiValueMap, err := d.rsi.Value()
	if err != nil {
		log.WithError(err).Debug("Cannot get value from indicator")
		return
	}
	kValue := rsiValueMap[stochrsi.ValueK]
	dValue := rsiValueMap[stochrsi.ValueD]
	if dValue == 0 {
		return
	}
	if kValue > upperThreshold && dValue > upperThreshold {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []broker.Order{}, []broker.Position{}
	}
	if kValue < lowerThreshold && dValue < lowerThreshold {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []broker.Order{}, []broker.Position{}
	}
	return
}

func (d *RSI) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
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

func (d *RSI) Name() string {
	return strategy.NameStochRSI
}

func (d *RSI) String() string {
	return d.Name()
}
