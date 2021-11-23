package tick

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestTick_Spread(t *testing.T) {
	var bid = decimal.NewFromFloat(1.00)
	var ask = decimal.NewFromFloat(1.50)
	var tick = New("EURUSD", time.Now(), bid, ask)
	var spread, _ = tick.Spread().Float64()
	assert.EqualFloat64(t, 0.50, spread)
}

func TestTick_SpreadInPercent(t *testing.T) {
	var bid = decimal.NewFromFloat(0.80)
	var ask = decimal.NewFromFloat(1.50)
	var tick = New("EURUSD", time.Now(), bid, ask)
	var spread, _ = tick.SpreadInPercent().Float64()
	assert.EqualFloat64(t, 87.50, spread)
}
