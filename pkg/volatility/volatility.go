package volatility

import (
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/circularbuffer"
)

type Volatility struct {
	cb *circularbuffer.CircularBuffer
}

func New(minSize, maxSize int) *Volatility {
	return &Volatility{
		cb: circularbuffer.New(minSize, maxSize),
	}
}

func (v *Volatility) AddOHLC(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	volaFloat, _ := o.VolatilityInPercentage().Float64()
	v.cb.Insert(volaFloat)
}

func (v *Volatility) MedianVolatilityInPercentage() (float64, error) {
	return v.cb.Median()
}

func (v *Volatility) VolatilityInPercentageQuantile(quantile float64) (float64, error) {
	return v.cb.Quantile(quantile)
}
