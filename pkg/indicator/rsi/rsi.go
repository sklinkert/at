package rsi

import (
	"github.com/falzm/golang-ring"
	"github.com/markcheno/go-talib"
	"github.com/sklinkert/at/pkg/indicator"
	"github.com/sklinkert/at/pkg/ohlc"
)

const Value = "RSI_VALUE"

type RSI struct {
	cb   ring.Ring
	size int
}

// New creates a new instance.
// size is usually 14
func New(size int) *RSI {
	cb := ring.Ring{}

	// The talib code seems to be doing a simple moving average for the initial n values,
	// and then do 1/n exponential smoothing thereafter. This is the standard Wilder's RSI.
	// I believe the calculations shold start at the beginning of the data and not using a sliding window which would
	// be problematic due to the simple moving average for the 1st n values.
	// So basically, in order for talib RSI to be calculated as 'accurate' as possible, the number of price points
	// I pass into the function should greatly exceed the number of price points needed to initialize the indicator
	// until you reach an n value where further increasing number of price points has a negligible effect on
	// the RSI value.
	// https://www.reddit.com/r/algotrading/comments/kmgmtt/cant_validate_rsi_indicator_values_from_talib_vs/
	cb.SetCapacity(size * 10)

	return &RSI{
		cb:   cb,
		size: size,
	}
}

func (v *RSI) Insert(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	close, _ := o.Close.Float64()
	v.cb.Enqueue(close)
}

func (v *RSI) Value() (map[string]float64, error) {
	var m = map[string]float64{}

	closePrices := v.cb.Values()
	if len(closePrices) < v.size+1 {
		return nil, indicator.ErrNotEnoughData
	}
	//fmt.Printf("RSI Input: %+v (%d)\n", closePrices, len(closePrices))

	rsi := talib.Rsi(closePrices, v.size)
	if len(rsi) > 0 {
		//fmt.Printf("RSI Output: %+v (%d)\n", rsi, len(rsi))
		m[Value] = rsi[len(rsi)-1]
	}
	return m, nil
}

func (v *RSI) ValueResultKeys() []string {
	return []string{Value}
}
