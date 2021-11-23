package adx

import (
	"github.com/falzm/golang-ring"
	"github.com/markcheno/go-talib"
	"github.com/sklinkert/at/pkg/indicator"
	"github.com/sklinkert/at/pkg/ohlc"
)

const Value = "ADX_VALUE"

type ADX struct {
	closePrices ring.Ring
	highPrices  ring.Ring
	lowPrices   ring.Ring
	size        int
}

// New creates a new instance.
// size is usually 14
func New(size int) *ADX {
	highPrices := ring.Ring{}
	closePrices := ring.Ring{}
	lowPrices := ring.Ring{}

	highPrices.SetCapacity(size * 2)
	lowPrices.SetCapacity(size * 2)
	closePrices.SetCapacity(size * 2)

	return &ADX{
		highPrices:  highPrices,
		lowPrices:   lowPrices,
		closePrices: closePrices,
		size:        size,
	}
}

func (v *ADX) Insert(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	closePrice, _ := o.Close.Float64()
	v.closePrices.Enqueue(closePrice)

	high, _ := o.High.Float64()
	v.highPrices.Enqueue(high)

	low, _ := o.Low.Float64()
	v.lowPrices.Enqueue(low)
}

func (v *ADX) Value() (map[string]float64, error) {
	closePrices := v.closePrices.Values()
	highPrices := v.highPrices.Values()
	lowPrices := v.lowPrices.Values()
	if len(closePrices) < v.size*2 || len(highPrices) < v.size*2 || len(lowPrices) < v.size*2 {
		return nil, indicator.ErrNotEnoughData
	}

	var m = map[string]float64{}
	adx := talib.Adx(highPrices, lowPrices, closePrices, v.size)
	if len(adx) > 0 {
		//fmt.Printf("ADX Output: %+v (%d)\n", adx, len(adx))
		m[Value] = adx[len(adx)-1]
	}
	return m, nil
}

func (v *ADX) ValueResultKeys() []string {
	return []string{Value}
}
