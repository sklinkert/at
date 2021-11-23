package stochrsi

import (
	"github.com/markcheno/go-talib"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/circularbuffer"
)

const ValueK = "StochRSI_VALUE_K"
const ValueD = "StochRSI_VALUE_D"

type StochRSI struct {
	cb           *circularbuffer.CircularBuffer
	fastKPeriod  int
	fastDPeriod  int
	inTimePeriod int
}

func New(fastKPeriod, fastDPeriod, size int) *StochRSI {
	return &StochRSI{
		cb:           circularbuffer.New(size*3, size*3),
		inTimePeriod: size,
		fastKPeriod:  fastKPeriod,
		fastDPeriod:  fastDPeriod,
	}
}

func (v *StochRSI) Insert(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	closePrice, _ := o.Close.Float64()
	v.cb.Insert(closePrice)
}

func (v *StochRSI) Value() (map[string]float64, error) {
	var m = map[string]float64{}

	closePrices, err := v.cb.GetAll()
	if err != nil {
		return nil, err
	}

	k, d := talib.StochRsi(closePrices, v.inTimePeriod, v.fastKPeriod, v.fastDPeriod, talib.SMA)
	if len(k) > 0 {
		m[ValueK] = k[len(k)-1]
	}
	if len(d) > 0 {
		m[ValueD] = k[len(d)-1]
	}

	return m, err
}

func (v *StochRSI) ValueResultKeys() []string {
	return []string{ValueK, ValueD}
}
