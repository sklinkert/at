package rsiadx

import (
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/eo"
	"github.com/sklinkert/at/pkg/helper"
	indicatoradx "github.com/sklinkert/at/pkg/indicator/adx"
	indicatorrsi "github.com/sklinkert/at/pkg/indicator/rsi"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

type RSIADX struct {
	clog           *log.Entry
	instrument     string
	rsi            *indicatorrsi.RSI
	adx            *indicatoradx.ADX
	candleDuration time.Duration
	eo             *eo.EnvironmentOverlay
	openPositions  []broker.Position
	openOrders     []broker.Order
}

const (
	orderPricePrecision = 1
	adxThreshold        = 35
	adxCandles          = 10
	rsiCandles          = 2
	targetInPercent     = 5.0
	stopLossInPercent   = 2.5
	maxAgeOpenPosition  = time.Hour * 2
)

func New(instrument string, candleDuration time.Duration) *RSIADX {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})

	return &RSIADX{
		clog:           clog,
		instrument:     instrument,
		rsi:            indicatorrsi.New(rsiCandles),
		adx:            indicatoradx.New(adxCandles),
		candleDuration: candleDuration,
		eo:             eo.New(),
	}
}

func (d *RSIADX) OnPosition(openPositions []broker.Position, _ []broker.Position) {
	d.openPositions = openPositions
}

func (d *RSIADX) OnOrder(openOrders []broker.Order) {
	d.openOrders = openOrders
}

func (d *RSIADX) OnTick(_ tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	return
}

func (d *RSIADX) GetCandleDuration() time.Duration {
	return d.candleDuration
}

func (d *RSIADX) OnWarmUpCandle(closedCandle *ohlc.OHLC) {
	d.rsi.Insert(closedCandle)
	d.adx.Insert(closedCandle)
	d.eo.AddCandle(closedCandle)

	rsiValue, rsiErr := d.getRSI()
	adxValue, adxErr := d.getADX()
	log.WithFields(log.Fields{
		"Start":   closedCandle.Start,
		"Close":   closedCandle.Close,
		"RSI":     rsiValue,
		"RSI_ERR": rsiErr,
		"ADX":     adxValue,
		"ADX_ERR": adxErr,
	}).Info("Processing warmup candle")
}

func (d *RSIADX) GetWarmUpCandleAmount() uint {
	return adxCandles * 2
}

func (d *RSIADX) OnCandle(closedCandles []*ohlc.OHLC) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	closedCandle := closedCandles[len(closedCandles)-1]

	d.rsi.Insert(closedCandle)
	d.adx.Insert(closedCandle)
	d.eo.AddCandle(closedCandle)

	if len(d.openPositions) > 0 {
		return d.checkOpenPositions(closedCandle, closedCandles, d.openPositions)
	}

	if !d.isStrongADXTrend() {
		return
	}

	if d.isRSIShortSignal() {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []broker.Order{}, []broker.Position{}
	} else if d.isRSILongSignal() {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []broker.Order{}, []broker.Position{}
	}
	return
}

func (d *RSIADX) checkOpenPositions(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, openPositions []broker.Position) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	var previousCandle = closedCandles[len(closedCandles)-2]

	for _, openPosition := range openPositions {
		if openPosition.Age(closedCandle.End) > maxAgeOpenPosition &&
			openPosition.PerformanceAbsolute(closedCandle.Close, closedCandle.Close) > 0 {
			toClosePositions = append(toClosePositions, openPosition)
			continue
		}

		switch openPosition.BuyDirection {
		case broker.BuyDirectionLong:
			if closedCandle.Close.GreaterThan(previousCandle.High) {
				toClosePositions = append(toClosePositions, openPosition)
			}
		case broker.BuyDirectionShort:
			if closedCandle.Close.LessThan(previousCandle.Low) {
				toClosePositions = append(toClosePositions, openPosition)
			}
		}
	}
	return
}

func (d *RSIADX) isRSILongSignal() bool {
	var rsiValue, err = d.getRSI()
	var _, rsiLowerThreshold = d.eo.RSI()
	return rsiValue <= rsiLowerThreshold && err == nil
}

func (d *RSIADX) isRSIShortSignal() bool {
	var rsiValue, err = d.getRSI()
	var rsiUpperThreshold, _ = d.eo.RSI()
	return rsiValue >= rsiUpperThreshold && err == nil
}

func (d *RSIADX) isStrongADXTrend() bool {
	var adxValue, err = d.getADX()
	return adxValue >= adxThreshold && err == nil
}

func (d *RSIADX) getRSI() (rsiValue float64, err error) {
	rsiValueMap, err := d.rsi.Value()
	if err != nil {
		log.WithError(err).Warn("Cannot get value from indicator")
		return 0, err
	}
	rsiValue = rsiValueMap[indicatorrsi.Value]
	return
}

func (d *RSIADX) getADX() (adxValue float64, err error) {
	adxValueMap, err := d.adx.Value()
	if err != nil {
		log.WithError(err).Warn("Cannot get value from indicator")
		return 0, err
	}
	adxValue = adxValueMap[indicatoradx.Value]
	return
}

func (d *RSIADX) prepareOrder(closedCandle *ohlc.OHLC, direction broker.BuyDirection, size float64) broker.Order {
	var (
		targetPrice   = helper.CalcTargetPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(targetInPercent), direction).Round(orderPricePrecision)
		stopLossPrice = helper.CalcStopLossPriceByPercentage(closedCandle.Close, helper.FloatToDecimal(stopLossInPercent), direction).Round(orderPricePrecision)
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

func (d *RSIADX) Name() string {
	return strategy.NameRSIADX
}

func (d *RSIADX) String() string {
	return d.Name()
}
