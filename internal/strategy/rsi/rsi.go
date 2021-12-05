package rsi

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/helper"
	indicatorrsi "github.com/sklinkert/at/pkg/indicator/rsi"
	"github.com/sklinkert/at/pkg/indicator/sma"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

type RSI struct {
	clog           *log.Entry
	instrument     string
	rsi            *indicatorrsi.RSI
	sma            *sma.SMA
	candleDuration time.Duration
}

const (
	upperThreshold     = 75
	lowerThreshold     = 25
	targetInPercent    = 0.2
	stopLossInPercent  = 0.6
	rsiSize            = 14
	smaCandles         = 200
	maxAgeOpenPosition = time.Hour * 2
)

func New(instrument string, candleDuration time.Duration) *RSI {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})

	return &RSI{
		clog:           clog,
		instrument:     instrument,
		rsi:            indicatorrsi.New(rsiSize),
		sma:            sma.New(smaCandles),
		candleDuration: candleDuration,
	}
}

func (d *RSI) GetCandleDuration() time.Duration {
	return d.candleDuration
}

func (d *RSI) ProcessWarmUpCandle(closedCandle *ohlc.OHLC) {
	d.rsi.Insert(closedCandle)
	d.sma.Insert(closedCandle)

	rsiValue, err := d.getRSIValues()
	log.WithFields(log.Fields{
		"Start":   closedCandle.Start,
		"Close":   closedCandle.Close,
		"RSI":     rsiValue,
		"RSI_ERR": err,
	}).Info("Processing warmup candle")
}

func (d *RSI) GetWarmUpCandleAmount() uint {
	return rsiSize * 10
}

func (d *RSI) ProcessCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openOrders []broker.Order, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toCloseOrderIDs []string, toClosePositions []broker.Position) {
	d.rsi.Insert(closedCandle)
	d.sma.Insert(closedCandle)

	rsiValue, err := d.getRSIValues()
	log.WithFields(log.Fields{
		"RSI":     rsiValue,
		"RSI_ERR": err,
		"Close":   closedCandle.Close,
	}).Info("Processing Candle")

	// No night trading
	//if currentTick.Datetime.Hour() < 10 || currentTick.Datetime.Hour() > 20 {
	//	return
	//}

	for _, openPosition := range openPositions {
		//log.Infof("Have open position: %s", openPosition.String())

		if openPosition.Age(closedCandle.End) > maxAgeOpenPosition &&
			openPosition.PerformanceAbsolute(closedCandle.Close, closedCandle.Close) > 0 {
			toClosePositions = append(toClosePositions, openPosition)
			continue
		}

		switch openPosition.BuyDirection {
		case broker.BuyDirectionShort:
			if d.isRSILongSignal() {
				toClosePositions = append(toClosePositions, openPosition)

				// counter position
				if d.isSMALongSignal(closedCandle.Close) {
					toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
					toOpen = append(toOpen, toOpenNew)
				}
			}
		case broker.BuyDirectionLong:
			if d.isRSIShortSignal() {
				toClosePositions = append(toClosePositions, openPosition)

				// counter position
				if d.isSMAShortSignal(closedCandle.Close) {
					toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
					toOpen = append(toOpen, toOpenNew)
				}
			}
		}
	}
	if len(openPositions) > 0 {
		return toOpen, toCloseOrderIDs, toClosePositions
	}

	if d.isRSIShortSignal() && d.isSMAShortSignal(closedCandle.Close) {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionShort, 1.00)
		return []broker.Order{toOpenNew}, []string{}, []broker.Position{}
	}
	if d.isRSILongSignal() && d.isSMALongSignal(closedCandle.Close) {
		toOpenNew := d.prepareOrder(closedCandle, broker.BuyDirectionLong, 1.00)
		return []broker.Order{toOpenNew}, []string{}, []broker.Position{}
	}
	return
}

func (d *RSI) isSMALongSignal(closePrice decimal.Decimal) bool {
	return true

	//smaValue, err := d.sma.Value()
	//if err != nil {
	//	log.WithError(err).Warn("No SMA")
	//	return false
	//}
	//smaPrice := decimal.NewFromFloat(smaValue[sma.Value])
	//return closePrice.GreaterThan(smaPrice)
}

func (d *RSI) isSMAShortSignal(closePrice decimal.Decimal) bool {
	return true

	//smaValue, err := d.sma.Value()
	//if err != nil {
	//	log.WithError(err).Warn("No SMA")
	//	return false
	//}
	//smaPrice := decimal.NewFromFloat(smaValue[sma.Value])
	//return closePrice.LessThan(smaPrice)
}

func (d *RSI) isRSILongSignal() bool {
	var rsiValue, err = d.getRSIValues()
	return rsiValue <= lowerThreshold && err == nil
}

func (d *RSI) isRSIShortSignal() bool {
	var rsiValue, err = d.getRSIValues()
	return rsiValue >= upperThreshold && err == nil
}

func (d *RSI) getRSIValues() (rsiValue float64, err error) {
	rsiValueMap, err := d.rsi.Value()
	if err != nil {
		log.WithError(err).Warn("Cannot get value from indicator")
		return 0, err
	}
	rsiValue = rsiValueMap[indicatorrsi.Value]

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
	return strategy.NameRSI
}

func (d *RSI) String() string {
	return d.Name()
}
