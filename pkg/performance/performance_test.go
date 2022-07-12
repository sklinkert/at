package performance

import (
	"fmt"
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
	"testing"
	"time"
)

func TestAddOHLC(t *testing.T) {
	v := New(5, 10)
	now := time.Now()
	for i := 1; i < 12; i++ {
		o := ohlc.New("test", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(1.0), o.Start)
		o.NewPrice(decimal.NewFromFloat(float64(i)+1), o.Start)
		fmt.Printf("ADD: %d -> %s\n", i, o.PerformanceInPercentage())
		o.ForceClose()
		v.AddOHLC(o)
	}

	wantArray := []float64{11, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	isArray, err := v.cb.GetAll()
	assert.NoError(t.Fatalf, err)

	for i := range wantArray {
		want := wantArray[i] * 100
		if want != isArray[i] {
			t.Errorf("TestAddOHLC: wantArray differs: index=%d want=%.1f got=%.1f", i, want, isArray[i])
		}
	}
}

func TestMedianPerformanceInPercentage(t *testing.T) {
	v := New(10, 10)
	now := time.Now()
	for i := 0; i < 10; i++ {
		o := ohlc.New("EURUSD", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(1.0), o.Start)
		o.NewPrice(decimal.NewFromFloat(float64(i)+1.0), o.Start)
		o.ForceClose()
		v.AddOHLC(o)
	}

	perf, err := v.MedianPerformanceInPercentage()
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, 400, perf)
}

func TestPerformanceInPercentageQuantile(t *testing.T) {
	v := New(1000, 1000)
	now := time.Now()
	for i := 0; i < 1000; i++ {
		o := ohlc.New("EURUSD", now, time.Minute, false)
		o.NewPrice(decimal.NewFromFloat(1.0), o.Start)
		o.NewPrice(decimal.NewFromFloat(float64(i)+1.0), o.End)
		assert.True(t, o.Closed())
		v.AddOHLC(o)
	}

	isArray, err := v.cb.GetAll()
	assert.NoError(t.Fatalf, err)

	perf, err := v.PerformanceInPercentageQuantile(0)
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, isArray[0], perf)

	perf, err = v.PerformanceInPercentageQuantile(1)
	assert.NoError(t.Fatalf, err)
	assert.EqualFloat64(t, isArray[len(isArray)-1], perf)
}
