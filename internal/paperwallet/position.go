package paperwallet

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/tick"
	"sort"
)

// Buy open new position with target and stop loss
func (pw *Paperwallet) Buy(order broker.Order) (orderID string, err error) {
	pw.Lock()
	defer pw.Unlock()

	if err := order.Valid(); err != nil {
		log.WithError(err).Fatalf("Order is not valid: %+v", order)
	}
	if err := pw.buyCheckTargetAndStopLoss(order); err != nil {
		return "", err
	}

	orderID = uuid.New().String()
	if order.Type == broker.OrderTypeMarket {
		pw.openPosition(orderID, order)
	} else if order.Type == broker.OrderTypeLimit {
		pw.openOrders[orderID] = order
	}

	return orderID, nil
}

func (pw *Paperwallet) openPosition(orderID string, order broker.Order) broker.Position {
	positionRef := uuid.New().String()
	_, exists := pw.openPositions[positionRef]
	if exists {
		log.Fatalf("Buy: Position %q already exists", positionRef)
	}

	position := broker.Position{
		Reference:     positionRef,
		Instrument:    order.Instrument,
		BuyPrice:      pw.getBuyPriceByDirection(order.Direction),
		BuyTime:       pw.currentTick.Datetime,
		BuyDirection:  order.Direction,
		TargetPrice:   order.TargetPrice,
		StopLossPrice: order.StopLossPrice,
		Size:          order.Size,
	}
	pw.openPositions[position.Reference] = position
	delete(pw.openOrders, orderID)

	log.WithFields(log.Fields{
		"BuyTime":   pw.openPositions[positionRef].BuyTime,
		"Reference": pw.openPositions[positionRef].Reference,
		"Size":      order.Size,
	}).Debug("New position")

	return position
}

func (pw *Paperwallet) getSellPriceByDirection(direction broker.BuyDirection, slippage bool) decimal.Decimal {
	switch direction {
	case broker.BuyDirectionLong:
		if slippage {
			tradingFee := pw.getAbsoluteTradingFee(pw.currentTick.Bid)
			pw.totalTradingFee = pw.totalTradingFee.Add(tradingFee)
			return pw.currentTick.Bid.Sub(pw.slippageAbsolute).Sub(tradingFee)
		}
		return pw.currentTick.Bid
	case broker.BuyDirectionShort:
		if slippage {
			tradingFee := pw.getAbsoluteTradingFee(pw.currentTick.Ask)
			pw.totalTradingFee = pw.totalTradingFee.Add(tradingFee)
			return pw.currentTick.Ask.Add(pw.slippageAbsolute).Add(tradingFee)
		}
		return pw.currentTick.Ask
	default:
		log.Fatal("unsupported direction", direction)
	}

	// Never reached
	return decimal.Zero
}

var dec100 = decimal.NewFromFloat(100)

func (pw *Paperwallet) getAbsoluteTradingFee(price decimal.Decimal) decimal.Decimal {
	return price.Div(dec100).Mul(pw.tradingFeePercent)
}

func (pw *Paperwallet) sell(position broker.Position, optionalSellPrice decimal.Decimal, slippage bool, reason string) error {
	position, exists := pw.openPositions[position.Reference]
	if !exists {
		return broker.ErrPositionNotFound
	}

	if optionalSellPrice.IsZero() {
		position.SellPrice = pw.getSellPriceByDirection(position.BuyDirection, slippage)
	} else {
		position.SellPrice = optionalSellPrice
	}
	position.SellTime = pw.currentTick.Datetime

	_, exists = pw.closedPositions[position.Reference]
	if exists {
		log.Fatalf("sell: position.Reference already exists in pw.closedPositions: %q", position.Reference)
	}
	pw.closedPositions[position.Reference] = position
	pw.updateBalance(&position)
	delete(pw.openPositions, position.Reference)

	log.WithFields(log.Fields{
		"Reason":             reason,
		"BuyTime":            position.BuyTime.Local(),
		"SellTime":           position.SellTime.Local(),
		"Reference":          position.Reference,
		"Target":             position.TargetPrice,
		"StopLoss":           position.StopLossPrice,
		"TotalLossPositions": getTotalLossPositions(pw.closedPositions),
		"TotalPerformance":   decimal.NewFromFloat(getTotalPerf(pw.closedPositions)).Round(1),
		"OpenPositions":      len(pw.openPositions),
	}).Info("Position closed")

	return nil
}

func (pw *Paperwallet) updateBalance(closedPosition *broker.Position) {
	profit := decimal.NewFromFloat(closedPosition.PerformanceAbsolute(decimal.Zero, decimal.Zero))
	pw.balance = pw.balance.Add(profit)
}

// Sell closes the given open position
func (pw *Paperwallet) Sell(position broker.Position) error {
	pw.Lock()
	defer pw.Unlock()

	return pw.sell(position, decimal.Decimal{}, true, "Initiated by trader")
}

func (pw *Paperwallet) buyCheckTargetAndStopLoss(order broker.Order) error {
	switch order.Direction {
	case broker.BuyDirectionLong:
		//if order.TargetPrice.LessThan(pw.currentTick.Ask) {
		//	return fmt.Errorf("target is below current price: %s < %s", order.TargetPrice, pw.currentTick.Ask)
		//}
		if pw.currentTick.Ask.LessThan(order.StopLossPrice) {
			return fmt.Errorf("current price is below stop loss: %s < %s", pw.currentTick.Ask, order.StopLossPrice)
		}
	case broker.BuyDirectionShort:
		//if order.TargetPrice.GreaterThan(pw.currentTick.Bid) {
		//	return fmt.Errorf("target is above current price: %s > %s", order.TargetPrice, pw.currentTick.Bid)
		//}
		if pw.currentTick.Bid.GreaterThan(order.StopLossPrice) {
			return fmt.Errorf("current price is above stop loss: %s > %s", pw.currentTick.Bid, order.StopLossPrice)
		}
	default:
		log.Fatalf("unsupported order direction %s", order.Direction)
	}
	return nil
}

func (pw *Paperwallet) getBuyPriceByDirection(direction broker.BuyDirection) decimal.Decimal {
	switch direction {
	case broker.BuyDirectionLong:
		var tradingFee = pw.getAbsoluteTradingFee(pw.currentTick.Ask)
		pw.totalTradingFee = pw.totalTradingFee.Add(tradingFee)
		return pw.currentTick.Ask.Add(pw.slippageAbsolute).Add(tradingFee)
	case broker.BuyDirectionShort:
		var tradingFee = pw.getAbsoluteTradingFee(pw.currentTick.Bid)
		pw.totalTradingFee = pw.totalTradingFee.Add(tradingFee)
		return pw.currentTick.Bid.Sub(pw.slippageAbsolute).Sub(tradingFee)
	default:
		log.Fatal("unsupported direction", direction)
	}

	// Never reached
	return decimal.Zero
}

func (pw *Paperwallet) checkOpenPositionsTarget(position broker.Position) (positionSold bool) {
	if position.TargetPrice.Equal(decimal.Zero) {
		return true
	}

	if position.BuyDirection == broker.BuyDirectionLong {
		if pw.currentTick.Bid.GreaterThanOrEqual(position.TargetPrice) {
			_ = pw.sell(position, position.TargetPrice, false, "Target hit")
			return true
		}
	} else {
		if pw.currentTick.Ask.LessThanOrEqual(position.TargetPrice) {
			_ = pw.sell(position, position.TargetPrice, false, "Target hit")
			return true
		}
	}
	return false
}

func (pw *Paperwallet) checkOpenPositionsStopLoss(position broker.Position) (positionSold bool) {
	if position.BuyDirection == broker.BuyDirectionLong {
		if pw.currentTick.Bid.LessThanOrEqual(position.StopLossPrice) {
			_ = pw.sell(position, position.StopLossPrice, false, "Stop loss hit")
			return true
		}
	} else {
		if pw.currentTick.Ask.GreaterThanOrEqual(position.StopLossPrice) {
			_ = pw.sell(position, position.StopLossPrice, false, "Stop loss hit")
			return true
		}
	}
	return false
}

func (pw *Paperwallet) SetCurrenctPrice(currentTick tick.Tick) {
	pw.Lock()
	pw.currentTick = currentTick
	pw.checkOpenOrders()
	pw.checkOpenPositions()
	pw.Unlock()
}

func (pw *Paperwallet) checkOpenPositions() {
	for ref, position := range pw.openPositions {
		var perfPips = position.PerformanceAbsolute(pw.currentTick.Bid, pw.currentTick.Ask) * 10000
		if perfPips > position.MaxSurge {
			position.MaxSurge = perfPips
			pw.openPositions[ref] = position
		} else if perfPips < position.MaxDrawdown {
			position.MaxDrawdown = perfPips
			pw.openPositions[ref] = position
		}

		if pw.checkOpenPositionsTarget(position) {
			continue
		}
		if pw.checkOpenPositionsStopLoss(position) {
			continue
		}
	}
}

func (pw *Paperwallet) CloseAllOpenPositions() {
	positions, err := pw.GetOpenPositions()
	if err != nil {
		log.WithError(err).Error("CloseAllOpenPositions failed")
		return
	}
	for _, position := range positions {
		_ = pw.Sell(position)
	}
}

func getTotalLossPositions(closedPositions map[string]broker.Position) (lossPositions int) {
	for _, position := range closedPositions {
		if position.PerformanceAbsolute(decimal.Zero, decimal.Zero) < 0 {
			lossPositions++
		}
	}
	return
}

type sortedPositions []broker.Position

func (p sortedPositions) Len() int {
	return len(p)
}

func (p sortedPositions) Less(i, j int) bool {
	return p[i].BuyTime.Before(p[j].BuyTime)
}

func (p sortedPositions) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// GetClosedPositions returns all closed positions
func (pw *Paperwallet) GetClosedPositions() ([]broker.Position, error) {
	pw.RLock()
	defer pw.RUnlock()

	var positions sortedPositions
	for i := range pw.closedPositions {
		positions = append(positions, pw.closedPositions[i])
	}
	sort.Sort(positions)
	return positions, nil
}

// GetOpenPositionsByInstrument returns all open positions for given instrument name
func (pw *Paperwallet) GetOpenPositionsByInstrument(_ string) ([]broker.Position, error) {
	return pw.GetOpenPositions()
}

// GetOpenPositions returns all open positions
func (pw *Paperwallet) GetOpenPositions() ([]broker.Position, error) {
	pw.RLock()
	defer pw.RUnlock()

	var positions []broker.Position
	for i := range pw.openPositions {
		positions = append(positions, pw.openPositions[i])
	}
	return positions, nil
}

// GetOpenPosition returns the position for the given reference
func (pw *Paperwallet) GetOpenPosition(positionRef string) (broker.Position, error) {
	pw.RLock()
	defer pw.RUnlock()

	position, exists := pw.openPositions[positionRef]
	if !exists {
		return broker.Position{}, broker.ErrPositionNotFound
	}
	return position, nil
}
