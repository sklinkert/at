package stoch

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestStoch_Value(t *testing.T) {
	var stoch20 = New(14, 3)

	total := 0
	prices := 0
	now := time.Now()
	for i := 1; i < 100; i++ {
		o := ohlc.New("test", now, time.Minute, false)
		total += i
		prices++
		o.NewPrice(decimal.NewFromFloat(float64(i)), o.Start)
		o.ForceClose()
		stoch20.Insert(o)
	}

	stoch20Value, err := stoch20.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 100, stoch20Value[ValueK])

	assert.EqualFloat64(t, 100, stoch20Value[ValueD])
}

func TestStoch_Value_Down(t *testing.T) {
	var stoch20 = New(14, 3)

	total := 0
	prices := 0
	now := time.Now()
	for i := 100; i > 0; i-- {
		o := ohlc.New("test", now, time.Minute, false)
		total += i
		prices++
		o.NewPrice(decimal.NewFromFloat(float64(i)), o.Start)
		o.ForceClose()
		stoch20.Insert(o)
	}

	stoch20Value, err := stoch20.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 0, stoch20Value[ValueK])
}
