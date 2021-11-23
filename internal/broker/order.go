package broker

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

// Order is a request to open a new position
type Order struct {
	Direction                       BuyDirection
	Size                            float64
	Instrument                      string
	CurrencyCode                    string
	TargetPrice                     decimal.Decimal // optional
	StopLossPrice                   decimal.Decimal
	TrailingStopDistanceInPips      decimal.Decimal
	TrailingStopIncrementSizeInPips decimal.Decimal
	Note                            string
	CandleStart                     time.Time
}

// NewOrder creates a new order from given parameters
func NewOrder(direction BuyDirection, size float64, instrument string, targetPrice, stopLossPrice decimal.Decimal, note string) Order {
	return newOrderImpl(direction, size, instrument, targetPrice, stopLossPrice, decimal.Zero, decimal.Zero, note)
}

func NewOrderWithTrailingStop(direction BuyDirection, size float64, instrument string, targetPrice, trailingStopDistanceInPips, trailingStopIncrementSizeInPips decimal.Decimal, note string) Order {
	return newOrderImpl(direction, size, instrument, targetPrice, decimal.Zero, trailingStopDistanceInPips, trailingStopIncrementSizeInPips, note)
}

func newOrderImpl(direction BuyDirection, size float64, instrument string, targetPrice, stopLossPrice, trailingStopDistanceInPips, trailingStopIncrementSizeInPips decimal.Decimal, note string) Order {
	return Order{
		Direction:                       direction,
		Size:                            size,
		Instrument:                      instrument,
		TargetPrice:                     targetPrice,
		StopLossPrice:                   stopLossPrice,
		Note:                            note,
		TrailingStopDistanceInPips:      trailingStopDistanceInPips,
		TrailingStopIncrementSizeInPips: trailingStopIncrementSizeInPips,
	}
}

// MergeOrders merges two order slices
func MergeOrders(orders1, orders2 []Order) []Order {
	return append(orders1, orders2...)
}

// HasTargetPrice checks if the optional target price has been set
func (order *Order) HasTargetPrice() bool {
	return !order.TargetPrice.IsZero()
}

// Valid checks if the given order contains malformed data
func (order *Order) Valid() error {
	if order.Direction != BuyDirectionShort && order.Direction != BuyDirectionLong {
		return fmt.Errorf("unknown order direction %v", order.Direction)
	}
	if !order.StopLossPrice.IsZero() && !order.TrailingStopDistanceInPips.IsZero() {
		return fmt.Errorf("cannot set both, stop level and trailing stop")
	}
	if !order.StopLossPrice.IsZero() && !order.TrailingStopIncrementSizeInPips.IsZero() {
		return fmt.Errorf("cannot set both, stop level and trailing stop")
	}
	if !order.TrailingStopDistanceInPips.IsZero() && order.TrailingStopIncrementSizeInPips.IsZero() {
		return fmt.Errorf("need trailing stop increment size")
	}
	if order.Size <= 0 {
		return fmt.Errorf("cannot be <= 0")
	}
	return nil
}
