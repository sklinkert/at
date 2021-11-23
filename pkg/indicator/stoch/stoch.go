package stoch

import (
	"github.com/markcheno/go-talib"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/circularbuffer"
)

const ValueK = "Stoch_VALUE_K"
const ValueD = "Stoch_VALUE_D"

type Stoch struct {
	closePrices *circularbuffer.CircularBuffer
	highPrices  *circularbuffer.CircularBuffer
	lowPrices   *circularbuffer.CircularBuffer
	fastKPeriod int
	fastDPeriod int
}

func New(fastKPeriod, fastDPeriod int) *Stoch {
	return &Stoch{
		closePrices: circularbuffer.New(1, fastKPeriod*fastDPeriod),
		highPrices:  circularbuffer.New(1, fastKPeriod*fastDPeriod),
		lowPrices:   circularbuffer.New(1, fastKPeriod*fastDPeriod),
		fastKPeriod: fastKPeriod,
		fastDPeriod: fastDPeriod,
	}
}

func (v *Stoch) Insert(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	closePrice, _ := o.Close.Float64()
	v.closePrices.Insert(closePrice)

	highPrice, _ := o.High.Float64()
	v.highPrices.Insert(highPrice)

	lowPrice, _ := o.Low.Float64()
	v.lowPrices.Insert(lowPrice)
}

func (v *Stoch) Value() (map[string]float64, error) {
	var m = map[string]float64{}

	closePrices, err := v.closePrices.GetAll()
	if err != nil {
		return nil, err
	}
	highPrices, err := v.highPrices.GetAll()
	if err != nil {
		return nil, err
	}
	lowPrices, err := v.lowPrices.GetAll()
	if err != nil {
		return nil, err
	}

	k, d := talib.StochF(highPrices, lowPrices, closePrices, v.fastKPeriod, v.fastDPeriod, talib.SMA)
	if len(k) > 0 {
		m[ValueK] = k[len(k)-1]
	}
	if len(d) > 0 {
		m[ValueD] = k[len(d)-1]
	}

	return m, err
}

func (v *Stoch) ValueResultKeys() []string {
	return []string{ValueK, ValueD}
}
