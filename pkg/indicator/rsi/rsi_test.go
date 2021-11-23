package rsi

import (
	"fmt"
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/indicator"
	"github.com/sklinkert/at/pkg/ohlc"
	"math/rand"
	"testing"
	"time"
)

func TestRSI_Value(t *testing.T) {
	var prices = []float64{14.5, 18.45, 12.75, 15.35, 13.05, 16.10, 12.20, 11.65, 13.25, 15.30, 14.85, 16.15, 19.05, 21.45, 17.55}
	var rsi14 = New(14)

	for i := len(prices) - 1; i >= 0; i-- {
		rsi14.Insert(generateCandle(prices[i]))
	}

	rsiValue, err := rsi14.Value()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 45.83901773533423, rsiValue[Value])
}

func TestRSI_Value_Shift(t *testing.T) {
	var prices1 = []float64{15947.1, 15952.1, 15953.6, 15952.1, 15953.6, 15955.6, 15952.6, 15954.1, 15952.1, 15962.1, 15960.1, 15960, 15959.8, 15959, 15959.9}
	var prices2 = []float64{15948, 15947.1, 15952.1, 15953.6, 15952.1, 15953.6, 15955.6, 15952.6, 15954.1, 15952.1, 15962.1, 15960.1, 15960, 15959.8, 15959, 15959.9}

	var rsi1 = New(14)
	for _, price := range prices1 {
		rsi1.Insert(generateCandle(price))
	}
	rsiValue, err := rsi1.Value()
	fmt.Printf("rsi1 -> %f\n", rsiValue[Value])
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 69.99999999999888, rsiValue[Value])

	var rsi2 = New(14)
	for _, price := range prices2 {
		rsi2.Insert(generateCandle(price))
	}
	rsiValue, err = rsi2.Value()
	fmt.Printf("rsi2 -> %f\n", rsiValue[Value])
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 68.15212319178684, rsiValue[Value])
}

func randFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}

func TestRSI_Random(t *testing.T) {
	var prices1 = randFloats(10, 100, 14*3)

	var rsi1 = New(14)
	for _, price := range prices1 {
		rsi1.Insert(generateCandle(price))
	}
	rsiValue, err := rsi1.Value()
	fmt.Printf("rsi1 -> %f\n", rsiValue[Value])
	assert.NoError(t.Fatalf, err)

	prices2 := randFloats(10, 100, 14*10)
	prices2 = append(prices2, prices1...)
	rsi1 = New(14)
	for _, price := range prices2 {
		rsi1.Insert(generateCandle(price))
	}
	rsi2Value, err := rsi1.Value()
	assert.NoError(t.Fatalf, err)
	fmt.Printf("rsi2 -> %f\n", rsi2Value[Value])

	diff := rsi2Value[Value] - rsiValue[Value]
	fmt.Printf("diff: %f\n", diff)
	assert.True(t, diff < 1)
}

func TestRSI_NotEnoughCandles(t *testing.T) {
	var rsiIndicator = New(14)
	rsiIndicator.Insert(generateCandle(1))
	rsiIndicator.Insert(generateCandle(2))
	_, err := rsiIndicator.Value()
	assert.EqualErrors(t, err, indicator.ErrNotEnoughData)
}

func generateCandle(price float64) *ohlc.OHLC {
	var o = ohlc.New("test", time.Now(), time.Minute, false)
	o.NewPrice(decimal.NewFromFloat(price), o.Start)
	o.ForceClose()
	return o
}
