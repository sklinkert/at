package indicator

import (
	"errors"
	"github.com/sklinkert/at/pkg/ohlc"
)

type Indicator interface {
	Insert(o *ohlc.OHLC)
	ValueResultKeys() []string
	Value() (map[string]float64, error)
}

var ErrNotEnoughData = errors.New("not enough data to calculate indicator")
