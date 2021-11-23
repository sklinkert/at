package eo

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func Test_riskLevelHigh(t *testing.T) {
	overlay := New()

	for i := 0; i < 100; i++ {
		candle := generateCandle(float64(i))
		overlay.AddCandle(candle)
	}
	level := overlay.riskLevel()
	assert.EqualInt(t, int(RExtreme), int(level))
}

func Test_riskLevelLow(t *testing.T) {
	overlay := New()

	for i := 100; i > 0; i-- {
		candle := generateCandle(float64(i))
		overlay.AddCandle(candle)
	}
	level := overlay.riskLevel()
	assert.EqualInt(t, int(RLow), int(level))
}

func generateCandle(diff float64) *ohlc.OHLC {
	now := time.Now()
	candle := ohlc.New("test", now, time.Minute, false)
	candle.NewPrice(decimal.NewFromFloat(10), now)
	candle.NewPrice(decimal.NewFromFloat(10+diff), now)
	candle.ForceClose()
	return candle
}
