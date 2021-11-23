package stochrsi

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestStochRSI_Value(t *testing.T) {
	var rsi20 = New(5, 2, 14)

	total := 0
	prices := 0
	now := time.Now()
	for i := 1; i < 100; i++ {
		o := ohlc.New("test", now, time.Minute, false)
		total += i
		prices++
		o.NewPrice(decimal.NewFromFloat(float64(i)), o.Start)
		o.ForceClose()
		rsi20.Insert(o)
	}

	rsi20Value, err := rsi20.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 100, rsi20Value[ValueK])

	assert.EqualFloat64(t, 100, rsi20Value[ValueD])
}

func TestStochRSI_Value_Down(t *testing.T) {
	var rsi20 = New(5, 2, 14)

	total := 0
	prices := 0
	now := time.Now()
	for i := 100; i > 0; i-- {
		o := ohlc.New("test", now, time.Minute, false)
		total += i
		prices++
		o.NewPrice(decimal.NewFromFloat(float64(i)), o.Start)
		o.ForceClose()
		rsi20.Insert(o)
	}

	rsi20Value, err := rsi20.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 0, rsi20Value[ValueK])
}
