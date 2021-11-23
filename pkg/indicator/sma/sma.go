package sma

import (
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/circularbuffer"
)

const Value = "SMA_VALUE"

type SMA struct {
	cb *circularbuffer.CircularBuffer
}

func New(size int) *SMA {
	return &SMA{
		cb: circularbuffer.New(size, size),
	}
}

func (v *SMA) Insert(o *ohlc.OHLC) {
	if !o.Closed() {
		return
	}

	close, _ := o.Close.Float64()
	v.cb.Insert(close)
}

func (v *SMA) Value() (map[string]float64, error) {
	var err error
	var m = map[string]float64{}

	m[Value], err = v.cb.Average()
	return m, err
}

func (v *SMA) ValueResultKeys() []string {
	return []string{Value}
}
