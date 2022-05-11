package scalper

import (
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

// scalper targets a small profit
// Entry: Do counter trade after a number of candles were in the same buy direction

type scalper struct {
	clog          *log.Entry
	openPositions []broker.Position
	openOrders    []broker.Order
	currentTick   tick.Tick
}

var (
	targetPercent   = decimal.NewFromFloat(0.12)
	stopLossPercent = decimal.NewFromFloat(0.25)
)

func New(instrument string) *scalper {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})

	mr := &scalper{
		clog: clog,
	}

	return mr
}

func (mr *scalper) OnPosition(openPositions []broker.Position, _ []broker.Position) {
	mr.openPositions = openPositions
}

func (mr *scalper) OnOrder(openOrders []broker.Order) {
	mr.openOrders = openOrders
}

func (mr *scalper) OnWarmUpCandle(_ *ohlc.OHLC) {}

func (mr *scalper) GetWarmUpCandleAmount() uint {
	return 1
}

func (mr *scalper) GetCandleDuration() time.Duration {
	return time.Minute * 5
}

func (mr *scalper) OnTick(currentTick tick.Tick) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	mr.currentTick = currentTick
	return
}

func (mr *scalper) OnCandle(closedCandles []*ohlc.OHLC) (toOpen, toClose []broker.Order, toClosePositions []broker.Position) {
	const candles = 10

	if len(mr.openPositions) > 0 {
		return
	}
	if len(closedCandles) < candles {
		return
	}

	closedCandle := closedCandles[len(closedCandles)-1]

	var buyDirection = getBuyDirection(closedCandle)
	for i := len(closedCandles) - candles; i < len(closedCandles)-1; i++ {
		candleDirection := getBuyDirection(closedCandles[i])
		if candleDirection == buyDirection {
			// Candles before closedCandle needs to have a different direction
			return
		}
	}

	newOrder, err := mr.createOrder(closedCandle, buyDirection, 1, "scalper")
	if err != nil {
		log.WithError(err).Errorf("createOrder() failed")
		return
	}
	toOpen = append(toOpen, newOrder)

	return
}

func getBuyDirection(candle *ohlc.OHLC) broker.BuyDirection {
	if candle.PerformanceInPercentage().GreaterThanOrEqual(decimal.Zero) {
		return broker.BuyDirectionLong
	} else {
		return broker.BuyDirectionShort
	}
}

func (mr *scalper) createOrder(openOHLC *ohlc.OHLC, direction broker.BuyDirection, size float64, orderName string) (broker.Order, error) {
	targetPrice, err := mr.calcTargetPrice(direction, mr.currentTick, targetPercent)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcTargetPrice() failed %v", err)
	}

	stopLossPrice, err := mr.calcStopLossPrice(direction, mr.currentTick, stopLossPercent)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcStopLossPrice() failed %v", err)
	}

	mr.clog.WithFields(log.Fields{
		"Direction":       direction.String(),
		"Time":            mr.currentTick.Datetime,
		"CurrentTick.Bid": mr.currentTick.Bid,
		"CurrentTick.Ask": mr.currentTick.Ask,
		"TargetPrice":     targetPrice,
		"StopLossPrice":   stopLossPrice,
		"OHLC.Age":        openOHLC.Age(mr.currentTick.Datetime).String(),
	}).Debug("Creating new order")

	return broker.NewMarketOrder(direction, size, openOHLC.Instrument, targetPrice, stopLossPrice), nil
}

func (mr *scalper) calcTargetPrice(direction broker.BuyDirection, tick tick.Tick, percentage decimal.Decimal) (decimal.Decimal, error) {
	switch direction {
	case broker.BuyDirectionLong:
		var currentPrice = tick.Ask
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(percentage)
		return currentPrice.Add(percentFrom).Round(6), nil
	case broker.BuyDirectionShort:
		var currentPrice = tick.Bid
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(percentage)
		return currentPrice.Sub(percentFrom).Round(6), nil
	default:
		return decimal.Zero, broker.ErrUnknownBuyDirection
	}
}

func (mr *scalper) calcStopLossPrice(direction broker.BuyDirection, tick tick.Tick, percentage decimal.Decimal) (decimal.Decimal, error) {
	switch direction {
	case broker.BuyDirectionLong:
		var currentPrice = tick.Ask
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(percentage)
		return currentPrice.Sub(percentFrom).Round(6), nil
	case broker.BuyDirectionShort:
		var currentPrice = tick.Bid
		percentFrom := currentPrice.Div(decimal.NewFromFloat(100)).Mul(percentage)
		return currentPrice.Add(percentFrom).Round(6), nil
	default:
		return decimal.Zero, broker.ErrUnknownBuyDirection
	}
}

func (mr *scalper) Name() string {
	return strategy.NameScalper
}

func (mr *scalper) String() string {
	return mr.Name()
}
