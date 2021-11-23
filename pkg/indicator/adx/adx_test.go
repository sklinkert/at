package adx

import (
	"fmt"
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/indicator"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestADX_Bearish_Trend_Above_35(t *testing.T) {
	var adx1 = New(14)
	for i := 100; i > 0; i-- {
		adx1.Insert(generateCandle(float64(i)))
	}

	adxValue, err := adx1.Value()
	fmt.Printf("adx1 -> %f\n", adxValue[Value])
	assert.NoError(t.Fatalf, err)
	assert.True(t, adxValue[Value] > 35.0)
}

func TestADX_Bullish_Trend_Above_35(t *testing.T) {
	var adx1 = New(14)
	for i := 0; i < 100; i++ {
		adx1.Insert(generateCandle(float64(i)))
	}

	adxValue, err := adx1.Value()
	fmt.Printf("adx1 -> %f\n", adxValue[Value])
	assert.NoError(t.Fatalf, err)
	assert.True(t, adxValue[Value] > 35.0)
}

func TestADX_NotEnoughCandles(t *testing.T) {
	var adxIndicator = New(14)
	adxIndicator.Insert(generateCandle(1))
	adxIndicator.Insert(generateCandle(2))
	_, err := adxIndicator.Value()
	assert.EqualErrors(t, err, indicator.ErrNotEnoughData)
}

func generateCandle(price float64) *ohlc.OHLC {
	var o = ohlc.New("test", time.Now(), time.Minute, false)
	o.NewPrice(decimal.NewFromFloat(price), o.Start)
	o.ForceClose()
	return o
}
