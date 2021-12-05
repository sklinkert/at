package heikinashi

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/indicator"
	"github.com/sklinkert/at/pkg/indicator/sma"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/at/pkg/volatility"
	"time"
)

type HeikinAshi struct {
	clog                   *log.Entry
	instrument             string
	closedHACandles        []*ohlc.OHLC
	ignoreInitialDirection bool
	initialDirection       *broker.BuyDirection
	currentDirection       *broker.BuyDirection
	volaTracker            *volatility.Volatility
	sma                    indicator.Indicator
	candlesReceived        bool
}

func New(instrument string) *HeikinAshi {
	clog := log.WithFields(log.Fields{"INSTRUMENT": instrument})

	ha := &HeikinAshi{
		clog:                   clog,
		instrument:             instrument,
		ignoreInitialDirection: true,
		volaTracker:            volatility.New(10, 30),
		sma:                    sma.New(41),
	}

	return ha
}

func (ha *HeikinAshi) GetCandleDuration() time.Duration {
	return time.Hour
}

func (ha *HeikinAshi) GetWarmUpCandleAmount() uint {
	return 1
}

func (ha *HeikinAshi) ProcessWarmUpCandle(_ *ohlc.OHLC) {}

func (ha *HeikinAshi) ProcessCandle(closedCandle *ohlc.OHLC, closedCandles []*ohlc.OHLC, currentTick tick.Tick, openOrders []broker.Order, openPositions []broker.Position, _ []broker.Position) (toOpen []broker.Order, toCloseOrderIDs []string, toClosePositions []broker.Position) {
	if ha.GetCandleDuration() == time.Hour*24 && closedCandle.Start.Weekday() == time.Sunday {
		return
	}

	if ha.candlesReceived {
		ha.volaTracker.AddOHLC(closedCandle)
		ha.sma.Insert(closedCandle)
	} else {
		ha.candlesReceived = true
		for i := range closedCandles {
			ha.volaTracker.AddOHLC(closedCandles[i])
			ha.sma.Insert(closedCandles[i])
		}
	}

	if len(closedCandles) < 3 {
		return
	}
	haPrevious := ohlc.ToHeikinAshi(closedCandles[len(closedCandles)-3], closedCandles[len(closedCandles)-2])
	haNow := ohlc.ToHeikinAshi(closedCandles[len(closedCandles)-2], closedCandles[len(closedCandles)-1])
	ha.closedHACandles = append(ha.closedHACandles, haNow)

	var direction broker.BuyDirection
	if isLongCandle(haNow) && isLongCandle(haPrevious) && haNow.Close.GreaterThan(haPrevious.Close) {
		direction = broker.BuyDirectionLong
	} else if isShortCandle(haNow) && isShortCandle(haPrevious) && haNow.Close.LessThan(haPrevious.Close) {
		direction = broker.BuyDirectionShort
	} else {
		// undecided
		return
	}

	defer func() {
		ha.currentDirection = &direction
	}()

	if ha.ignoreInitialDirection {
		if ha.initialDirection == nil {
			// Don't trade already running trend
			ha.initialDirection = &direction
			return
		}
		if *ha.initialDirection != direction {
			ha.ignoreInitialDirection = false
			ha.initialDirection = nil
		}
	}

	var havePositionInRightDirection = false
	for _, position := range openPositions {
		if position.BuyDirection == direction {
			havePositionInRightDirection = true
			continue
		}
		toClosePositions = append(toClosePositions, position)
	}
	if havePositionInRightDirection {
		return
	}

	// Open new positions only when the direction is changing
	if ha.currentDirection == nil || direction == *ha.currentDirection {
		return
	}

	if err := ha.checkCandleAmount(*ha.currentDirection, 2); err != nil {
		ha.clog.WithError(err).Info("checkCandleAmount() failed")
		return
	}

	//if err := ha.checkSMA(direction, &currentTick); err != nil {
	//	ha.clog.WithError(err).Debug("checkSMA() failed")
	//	return
	//}

	order, err := ha.createOrder(haNow, currentTick, 0.20, direction)
	if err == nil {
		toOpen = append(toOpen, order)
	}
	order, err = ha.createOrder(haNow, currentTick, 0.50, direction)
	if err == nil {
		toOpen = append(toOpen, order)
	}
	order, err = ha.createOrder(haNow, currentTick, 0.95, direction)
	if err == nil {
		toOpen = append(toOpen, order)
	}

	return
}

// checkSMA - Check if SMA trend is supporting HA direction
//func (ha *HeikinAshi) checkSMA(direction broker.BuyDirection, currentTick *tick.Tick) error {
//	smaValueFloat, err := ha.sma.Value()
//	if err != nil {
//		return err
//	}
//	smaValue := decimal.NewFromFloat(smaValueFloat[sma.Value])
//
//	switch direction {
//	case broker.BuyDirectionLong:
//		if currentTick.Price().LessThan(smaValue) {
//			return fmt.Errorf("no support from SMA; direction is long, SMA is short (%s > %s)",
//				currentTick.Price(), smaValue)
//		}
//	case broker.BuyDirectionShort:
//		if currentTick.Price().GreaterThan(smaValue) {
//			return fmt.Errorf("no support from SMA; direction is short, SMA is long (%s > %s)",
//				currentTick.Price(), smaValue)
//		}
//	default:
//		return broker.ErrUnknownBuyDirection
//	}
//	return nil
//}

func (ha *HeikinAshi) createOrder(haCandle *ohlc.OHLC, currentTick tick.Tick, volaQuantileForTarget float64, direction broker.BuyDirection) (broker.Order, error) {
	const size = 1

	volaTargetFloat, err := ha.volaTracker.VolatilityInPercentageQuantile(volaQuantileForTarget)
	if err != nil {
		return broker.Order{}, err
	}
	volaTarget := decimal.NewFromFloat(volaTargetFloat).Abs().Mul(decimal.NewFromFloat(2))

	targetPrice, err := ha.calcTargetPrice(direction, currentTick, volaTarget)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcTargetPrice() failed %v", err)
	}

	stopLossPrice, err := ha.calcStopLossPrice(direction, currentTick)
	if err != nil {
		return broker.Order{}, fmt.Errorf("calcStopLossPrice() failed %v", err)
	}

	ha.clog.WithFields(log.Fields{
		"Direction":       direction.String(),
		"Time":            currentTick.Datetime,
		"CurrentTick.Bid": currentTick.Bid,
		"CurrentTick.Ask": currentTick.Ask,
		"VolaTarget":      volaTarget,
		"TargetPrice":     targetPrice,
		"StopLossPrice":   stopLossPrice,
		"OHLC.Age":        haCandle.Age(currentTick.Datetime).String(),
	}).Debug("Creating new order")

	return broker.NewMarketOrder(direction, size, haCandle.Instrument, targetPrice, stopLossPrice), nil
}

func isShortCandle(candle *ohlc.OHLC) bool {
	return candle.Close.LessThan(candle.Open)
}

func isLongCandle(candle *ohlc.OHLC) bool {
	return candle.Close.GreaterThan(candle.Open)
}

func (ha *HeikinAshi) checkCandleAmount(direction broker.BuyDirection, offset int) error {
	const candlesToCheck = 4
	max := candlesToCheck + offset
	lenCandles := len(ha.closedHACandles)

	if lenCandles < max {
		return errors.New("not enough closed candles to check")
	}
	candles := ha.closedHACandles[lenCandles-max : lenCandles-offset]
	if len(candles) != candlesToCheck {
		return fmt.Errorf("unexecpted amount of candles: %d", len(candles))
	}

	candlesInDirection := 0
	for _, candle := range candles {
		var candleDirection broker.BuyDirection
		if candle.PerformanceInPercentage().GreaterThanOrEqual(decimal.Zero) {
			candleDirection = broker.BuyDirectionLong
		} else {
			candleDirection = broker.BuyDirectionShort
		}
		if candleDirection == direction {
			candlesInDirection++
		}
	}

	if candlesInDirection < candlesToCheck {
		return fmt.Errorf("not enough candles in the right direction (%s), need %d, found %d",
			direction.String(), candlesToCheck, candlesInDirection)
	}
	return nil
}

//func isDOJI(candle *ohlc.OHLC) bool {
//	if candle == nil || !candle.Closed() {
//		return false
//	}
//	perfPercentage := candle.PerformanceInPercentage().Abs()
//	if perfPercentage.LessThanOrEqual(decimal.NewFromFloat(0.25)) {
//		return true
//	}
//	return false
//}

//func isBigCandle(candle *ohlc.OHLC) bool {
//	if candle == nil || !candle.Closed() {
//		return false
//	}
//	perfPercentage := candle.VolatilityInPercentage().Abs()
//	if perfPercentage.GreaterThanOrEqual(decimal.NewFromFloat(1)) {
//		return true
//	}
//	return false
//}

//func (ha *HeikinAshi) hadPositionInCandle(candle *ohlc.OHLC, closedPositions []broker.Position) bool {
//	for _, position := range closedPositions {
//		if position.BuyTime.After(candle.Start) && position.BuyTime.Before(candle.End) {
//			return true
//		}
//		if position.SellTime.After(candle.Start) && position.SellTime.Before(candle.End) {
//			return true
//		}
//	}
//	return false
//}

func (ha *HeikinAshi) calcTargetPrice(direction broker.BuyDirection, tick tick.Tick, perfMarginInPercentage decimal.Decimal) (decimal.Decimal, error) {
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
		return decimal.Zero, errors.New("unknown direction")
	}
}

func (ha *HeikinAshi) calcStopLossPrice(direction broker.BuyDirection, tick tick.Tick) (decimal.Decimal, error) {
	volaMaxFloat, err := ha.volaTracker.VolatilityInPercentageQuantile(0.95)
	if err != nil {
		return decimal.Zero, err
	}
	volaMax := decimal.NewFromFloat(volaMaxFloat).Abs()

	switch direction {
	case broker.BuyDirectionLong:
		return tick.Price().Sub(volaMax), nil
	case broker.BuyDirectionShort:
		return tick.Price().Add(volaMax), nil
	default:
		return decimal.Zero, errors.New("unknown direction")
	}
}

func (ha *HeikinAshi) Name() string {
	return strategy.NameHeikinAshi
}

func (ha *HeikinAshi) String() string {
	return strategy.NameHeikinAshi
}
