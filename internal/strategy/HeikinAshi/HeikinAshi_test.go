package heikinashi

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestHeikinAshi_Name(t *testing.T) {
	ha := New("test")
	assert.EqualStrings(t, strategy.NameHeikinAshi, ha.Name())
}

func getHACandlesLong(amount int) (candles []*ohlc.OHLC) {
	for i := 0; i < amount; i++ {
		now := time.Now()
		o := ohlc.New("test", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(float64(1)), o.Start)
		o.NewPrice(decimal.NewFromFloat(float64(2)), o.Start)
		o.ForceClose()
		candles = append(candles, o)
	}
	return candles
}

func getHACandlesShort(amount int) (candles []*ohlc.OHLC) {
	for i := 0; i < amount; i++ {
		now := time.Now()
		o := ohlc.New("test", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(float64(2)), o.Start)
		o.NewPrice(decimal.NewFromFloat(float64(1)), o.Start)
		o.ForceClose()
		candles = append(candles, o)
	}
	return candles
}

func TestHeikinAshi_checkCandleAmount(t *testing.T) {
	ha := New("test")
	err := ha.checkCandleAmount(broker.BuyDirectionLong, 0)
	assert.ErrorIncludesMessage(t, "not enough closed candles to check", err)

	// All candles in the wrong direction
	ha.closedHACandles = getHACandlesShort(6)
	err = ha.checkCandleAmount(broker.BuyDirectionLong, 0)
	assert.ErrorIncludesMessage(t, "not enough candles in the right direction", err)

	// All candles in the wrong direction with offset
	ha.closedHACandles = getHACandlesShort(6)
	err = ha.checkCandleAmount(broker.BuyDirectionLong, 2)
	assert.ErrorIncludesMessage(t, "not enough candles in the right direction", err)

	// All candles in the right direction with offset
	ha.closedHACandles = getHACandlesLong(4)
	ha.closedHACandles = append(ha.closedHACandles, getHACandlesShort(2)...)
	err = ha.checkCandleAmount(broker.BuyDirectionLong, 2)
	assert.NoError(t, err)
}
