package broker

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"testing"
)

func TestMergeOrders(t *testing.T) {
	order1 := NewMarketOrder(BuyDirectionShort, 1.00, "", decimal.Zero, decimal.Zero)
	order2 := NewMarketOrder(BuyDirectionShort, 1.00, "", decimal.Zero, decimal.Zero)
	orders1 := []Order{order1}
	orders2 := []Order{order2}

	orders := MergeOrders(orders1, orders2)
	assert.EqualInt(t, 2, len(orders))
}
