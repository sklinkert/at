package doji

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/helper"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/at/pkg/volatility"
	"time"
)

type Doji struct {
	clog           *log.Entry
	instrument     string
	volaTracker    *volatility.Volatility
	openCandle     *ohlc.OHLC
	closedCandles  []*ohlc.OHLC
	previousCandle *ohlc.OHLC
}

const (
	ohlcPeriod        = time.Minute * 60
	trackersMinPeriod = time.Hour * 24 * 15 // 15d
	trackersMaxPeriod = time.Hour * 24 * 30 // 30d
	orderNoteDOJI     = "doji"
	orderNoteCounter  = "counter"
)

var decZero = decimal.NewFromFloat(0)
var targetInPercent = decimal.NewFromFloat(0.045)
var dec2Pip = helper.Pips2Cent(decimal.NewFromFloat(2))

func New(instrument string) *Doji {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})
	trackerMin := int((trackersMinPeriod / ohlcPeriod).Nanoseconds())
	trackerMax := int((trackersMaxPeriod / ohlcPeriod).Nanoseconds())

	return &Doji{
		clog:        clog,
		instrument:  instrument,
		volaTracker: volatility.New(trackerMin, trackerMax),
	}
}

func (d *Doji) ProcessWarmUpCandle(_ *ohlc.OHLC) {}

func (d *Doji) sendTickToOHLC(currentTick tick.Tick) bool {
	var isOpen = d.openCandle.NewPrice(currentTick.Price(), currentTick.Datetime)
	if isOpen {
		return isOpen
	}

	// OHLC is closed
	d.previousCandle = d.openCandle
	if d.openCandle.Duration == ohlcPeriod {
		d.closedCandles = append(d.closedCandles, d.openCandle)
	}

	d.volaTracker.AddOHLC(d.openCandle)

	// Replace closed OHLC from openOHLCs list
	d.openCandle = ohlc.New(d.instrument, currentTick.Datetime, ohlcPeriod, true)
	d.sendTickToOHLC(currentTick)
	return false
}

func (d *Doji) GetCandleDuration() time.Duration {
	return time.Hour * 24
}

func (d *Doji) ProcessCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toClose []broker.Position) {
	return
}

func (d *Doji) GetWarmUpCandleAmount() uint {
	return 1
}

func (d *Doji) ProcessTick(currentTick tick.Tick, openPositions []broker.Position, closedPositions []broker.Position) (toOpen []broker.Order, toClose []broker.Position) {
	if d.openCandle == nil {
		// Initialize candle
		d.openCandle = ohlc.New(d.instrument, currentTick.Datetime, ohlcPeriod, true)
	}

	if !d.sendTickToOHLC(currentTick) {
		toOpenOld, toCloseOld := d.checkClosedPositions(d.openCandle, currentTick, openPositions, closedPositions)
		toOpen = mergeOrders(toOpen, toOpenOld)
		toClose = mergePositions(toClose, toCloseOld)
	}
	if d.openCandle.HasGaps() {
		return
	}
	if !isDOJI(d.previousCandle) {
		return
	}
	if len(openPositions) > 0 {
		return
	}
	if d.hadPositionInCandle(d.openCandle, closedPositions) {
		return
	}

	// No night trading
	if currentTick.Datetime.Hour() < 10 || currentTick.Datetime.Hour() > 20 {
		return
	}

	// Check for long signal
	if currentTick.Bid.GreaterThan(d.previousCandle.High.Add(dec2Pip)) {
		toOpenNew, err := d.createOrder(d.openCandle, currentTick, targetInPercent, broker.BuyDirectionLong, 1.00, orderNoteDOJI)
		if err == nil {
			toOpen = mergeOrders(toOpen, []broker.Order{toOpenNew})
		}
		return
	}

	// Check for short signal
	if currentTick.Ask.LessThan(d.previousCandle.Low.Sub(dec2Pip)) {
		toOpenNew, err := d.createOrder(d.openCandle, currentTick, targetInPercent, broker.BuyDirectionShort, 1.00, orderNoteDOJI)
		if err == nil {
			toOpen = mergeOrders(toOpen, []broker.Order{toOpenNew})
		}
		return
	}
	return
}

func mergeOrders(orders1, orders2 []broker.Order) []broker.Order {
	return append(orders1, orders2...)
}

func mergePositions(positions1, positions2 []broker.Position) []broker.Position {
	return append(positions1, positions2...)
}

func (d *Doji) checkClosedPositions(openOHLC *ohlc.OHLC, currentTick tick.Tick, openPositions []broker.Position, closedPositions []broker.Position) (toOpen []broker.Order, toClose []broker.Position) {
	if len(openPositions) > 0 {
		return
	}
	for _, closedPosition := range closedPositions {
		if !(closedPosition.SellTime.After(d.previousCandle.Start) && closedPosition.SellTime.Before(d.previousCandle.End)) {
			// position not closed in previousCandle candle
			continue
		}
		if closedPosition.Note == orderNoteDOJI && closedPosition.PerformanceInPercentage(decZero, decZero) < 0 {
			switch closedPosition.BuyDirection {
			case broker.BuyDirectionLong:
				log.Infof("Placing counter trade: short")
				order, err := d.createOrder(openOHLC, currentTick, targetInPercent, broker.BuyDirectionShort, 1, orderNoteCounter)
				if err == nil {
					toOpen = append(toOpen, order)
				}
			case broker.BuyDirectionShort:
				log.Infof("Placing counter trade: long")
				order, err := d.createOrder(openOHLC, currentTick, targetInPercent, broker.BuyDirectionLong, 1, orderNoteCounter)
				if err == nil {
					toOpen = append(toOpen, order)
				}
			}
			return
		}
	}
	return
}

func (d *Doji) createOrder(openOHLC *ohlc.OHLC, currentTick tick.Tick, perfMargin decimal.Decimal, direction broker.BuyDirection, size float64, note string) (broker.Order, error) {
	targetPrice, err := d.calcTargetPrice(direction, currentTick, perfMargin)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcTargetPrice() failed %v", err)
	}

	stopLossPrice, err := d.calcStopLossPrice(direction, currentTick)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcStopLossPrice() failed %v", err)
	}

	d.clog.WithFields(log.Fields{
		"Direction":       direction.String(),
		"Time":            currentTick.Datetime,
		"CurrentTick.Bid": currentTick.Bid,
		"CurrentTick.Ask": currentTick.Ask,
		"PerfMargin":      perfMargin,
		"TargetPrice":     targetPrice,
		"StopLossPrice":   stopLossPrice,
		"OHLC.Age":        openOHLC.Age(currentTick.Datetime).String(),
		//"Today.Perf":      d.today.PerformanceInPercentage().Round(2),
	}).Debug("Creating new order")

	return broker.NewOrder(direction, size, openOHLC.Instrument, targetPrice, stopLossPrice, note), nil
}

func isDOJI(candle *ohlc.OHLC) bool {
	if candle == nil || !candle.Closed() {
		return false
	}
	perfPercentage := candle.PerformanceInPercentage().Abs()
	return perfPercentage.LessThanOrEqual(decimal.NewFromFloat(0.025))
}

func (d *Doji) hadPositionInCandle(candle *ohlc.OHLC, closedPositions []broker.Position) bool {
	for _, position := range closedPositions {
		if position.BuyTime.After(candle.Start) && position.BuyTime.Before(candle.End) {
			return true
		}
		if position.SellTime.After(candle.Start) && position.SellTime.Before(candle.End) {
			return true
		}
	}
	return false
}

func (d *Doji) calcTargetPrice(direction broker.BuyDirection, tick tick.Tick, perfMarginInPercentage decimal.Decimal) (decimal.Decimal, error) {
	switch direction {
	case broker.BuyDirectionLong:
		var currentPrice = tick.Ask
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(perfMarginInPercentage)
		return currentPrice.Add(percentFrom).Round(6), nil
	case broker.BuyDirectionShort:
		var currentPrice = tick.Bid
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(perfMarginInPercentage)
		return currentPrice.Sub(percentFrom).Round(6), nil
	default:
		return decZero, errors.New("unknown direction")
	}
	//switch direction {
	//case broker.BuyDirectionLong:
	//	return tick.Price().Add(tick.Price().Sub(d.previousCandle.Low)), nil
	//case broker.BuyDirectionShort:
	//	return tick.Price().Sub(d.previousCandle.High.Sub(tick.Price())), nil
	//default:
	//	return decZero, errors.New("unknown direction")
	//}
}

func (d *Doji) calcStopLossPrice(direction broker.BuyDirection, tick tick.Tick) (decimal.Decimal, error) {
	if d.previousCandle == nil {
		return decZero, errors.New("d.previousCandle is empty")
	}
	//targetInPercent := helper.Pips2Cent(decimal.NewFromFloat(10))
	margin := decZero
	switch direction {
	case broker.BuyDirectionLong:
		return d.previousCandle.Low.Sub(margin), nil
	case broker.BuyDirectionShort:
		return d.previousCandle.High.Add(margin), nil
	default:
		return decZero, errors.New("unknown direction")
	}
}

func (d *Doji) Name() string {
	return strategy.NameDOJI
}

func (d *Doji) String() string {
	return ""
}
