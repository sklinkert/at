package broker

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"testing"
)

func TestPerformanceInPercentage(t *testing.T) {
	currentPrice := decimal.NewFromFloat(2.0)

	// Long
	position := Position{
		BuyPrice:     decimal.NewFromFloat(1.0),
		BuyDirection: BuyDirectionLong,
	}
	perf := position.PerformanceInPercentage(currentPrice, currentPrice)
	assert.EqualFloat64(t, 100, perf)

	// Short
	currentPrice = decimal.NewFromFloat(1.0)
	position = Position{
		BuyPrice:     decimal.NewFromFloat(0.0),
		BuyDirection: BuyDirectionShort,
	}
	perf = position.PerformanceInPercentage(currentPrice, currentPrice)
	assert.EqualFloat64(t, -100, perf)
}

func TestPerformanceAbsolute(t *testing.T) {
	currentPrice := decimal.NewFromFloat(2.0)

	// Long
	position := Position{
		BuyPrice:     decimal.NewFromFloat(1.0),
		BuyDirection: BuyDirectionLong,
		Size:         1.00,
	}
	perf := position.PerformanceAbsolute(currentPrice, currentPrice)
	assert.EqualFloat64(t, 1.0, perf)

	// Short
	position = Position{
		BuyPrice:     decimal.NewFromFloat(1.0),
		BuyDirection: BuyDirectionShort,
		Size:         1.00,
	}
	perf = position.PerformanceAbsolute(currentPrice, currentPrice)
	assert.EqualFloat64(t, -1.0, perf)
}
