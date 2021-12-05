package paperwallet

import (
	"errors"
	"github.com/sklinkert/at/internal/broker"
)

var ErrOrderNotFound = errors.New("order not found")

func (pw *Paperwallet) GetOpenOrders() []broker.Order {
	var openOrders []broker.Order
	for orderID, order := range pw.openOrders {
		order.ID = orderID
		openOrders = append(openOrders, order)
	}
	return openOrders
}

func (pw *Paperwallet) CancelOrder(orderID string) error {
	pw.Lock()
	defer pw.Unlock()

	_, exists := pw.openOrders[orderID]
	if !exists {
		return ErrOrderNotFound
	}
	delete(pw.openOrders, orderID)

	return nil
}

func (pw *Paperwallet) checkOpenOrders() {
	for orderID, order := range pw.openOrders {
		if order.Direction == broker.BuyDirectionLong {
			if pw.currentTick.Ask.LessThanOrEqual(order.Limit) {
				pw.openPosition(orderID, order)
				continue
			}
		} else if order.Direction == broker.BuyDirectionShort {
			if pw.currentTick.Bid.GreaterThanOrEqual(order.Limit) {
				pw.openPosition(orderID, order)
				continue
			}
		}
	}
}
