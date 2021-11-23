package round

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestRoundnum_Value(t *testing.T) {
	var now = time.Now()
	var rn = New()

	var testCases = []struct {
		price                  float64
		lowerRoundNumberWeak   float64
		lowerRoundNumberStrong float64
		upperRoundNumberWeak   float64
		upperRoundNumberStrong float64
	}{
		{
			price:                  0.23561,
			lowerRoundNumberWeak:   0.23,
			lowerRoundNumberStrong: 0.20,
			upperRoundNumberWeak:   0.24,
			upperRoundNumberStrong: 0.30,
		},
		{
			price:                  9.5,
			lowerRoundNumberWeak:   9.0,
			lowerRoundNumberStrong: 1.0,
			upperRoundNumberWeak:   10.0,
			upperRoundNumberStrong: 10.0,
		},
		{
			price:                  95,
			lowerRoundNumberWeak:   90,
			lowerRoundNumberStrong: 10,
			upperRoundNumberWeak:   100,
			upperRoundNumberStrong: 100,
		},
		{
			price:                  278,
			lowerRoundNumberWeak:   200,
			lowerRoundNumberStrong: 100,
			upperRoundNumberWeak:   300,
			upperRoundNumberStrong: 1000,
		},
		{
			price:                  1210,
			lowerRoundNumberWeak:   1200,
			lowerRoundNumberStrong: 1000,
			upperRoundNumberWeak:   1300,
			upperRoundNumberStrong: 10000,
		},
	}

	for _, tc := range testCases {
		o := ohlc.New("test", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(tc.price), o.Start)
		o.ForceClose()
		rn.Insert(o)

		rnValue, err := rn.Value()
		assert.NoError(t.Fatalf, err)
		assert.EqualInt(t.Fatalf, 4, len(rnValue))
		assert.EqualFloat64(t, tc.lowerRoundNumberWeak, rnValue[LowerRoundNumberWeak])
		assert.EqualFloat64(t, tc.lowerRoundNumberStrong, rnValue[LowerRoundNumberStrong])
		assert.EqualFloat64(t, tc.upperRoundNumberWeak, rnValue[UpperRoundNumberWeak])
		assert.EqualFloat64(t, tc.upperRoundNumberStrong, rnValue[UpperRoundNumberStrong])
	}

}
