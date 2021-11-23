package trader

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"testing"
)

var lossPosition = broker.Position{
	BuyPrice:     decimal.NewFromFloat(2),
	SellPrice:    decimal.NewFromFloat(1),
	BuyDirection: broker.BuyDirectionLong,
}

var winPosition = broker.Position{
	BuyPrice:     decimal.NewFromFloat(1),
	SellPrice:    decimal.NewFromFloat(2),
	BuyDirection: broker.BuyDirectionLong,
}

func TestTrader_getMaxConsecutiveLossTrades(t *testing.T) {
	var tr = Trader{}
	var closedPositions = []broker.Position{
		winPosition,
		lossPosition,
		lossPosition,
		winPosition,
		winPosition,
		lossPosition,
	}
	assert.EqualInt(t, 2, int(tr.getMaxConsecutiveLossTrades(closedPositions)))
}
