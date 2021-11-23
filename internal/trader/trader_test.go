package trader

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/tick"
	"testing"
	"time"
)

func TestTrader_flashCrashCheck(t *testing.T) {
	tick1 := tick.New("", time.Now(), decimal.NewFromFloat(100.0), decimal.NewFromFloat(100.0))
	tick2 := tick.New("", time.Now(), decimal.NewFromFloat(100.1), decimal.NewFromFloat(100.1))
	assert.NoError(t, flashCrashCheck(tick1, tick2))

	tick1 = tick.New("", time.Now(), decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.00))
	tick2 = tick.New("", time.Now(), decimal.NewFromFloat(2.0), decimal.NewFromFloat(2.00))
	assert.True(t, flashCrashCheck(tick1, tick2) != nil)
}

func Test__distanceInPercentage(t *testing.T) {
	price1 := decimal.NewFromFloat(10)
	price2 := decimal.NewFromFloat(12)
	assert.EqualStrings(t, "20", distanceInPercentage(price1, price2).String())

	price1 = decimal.NewFromFloat(10)
	price2 = decimal.NewFromFloat(8)
	assert.EqualStrings(t, "-20", distanceInPercentage(price1, price2).String())

	price1 = decimal.NewFromFloat(1)
	price2 = decimal.NewFromFloat(2)
	assert.EqualStrings(t, "100", distanceInPercentage(price1, price2).String())
}
