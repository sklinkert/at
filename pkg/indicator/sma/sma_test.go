package sma

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestSMA_Value(t *testing.T) {
	var sma20 = New(21)

	total := 0
	prices := 0
	now := time.Now()
	for i := 1; i < 22; i++ {
		o := ohlc.New("test", now, time.Minute, false)
		total += i
		prices++
		o.NewPrice(decimal.NewFromFloat(float64(i)), o.Start)
		o.ForceClose()
		sma20.Insert(o)
		if i < 20 {
			_, err := sma20.Value()
			assert.ErrorIncludesMessage(t, "not enough", err)
		}
	}

	sma20Value, err := sma20.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, float64(total/prices), sma20Value[Value])
}
