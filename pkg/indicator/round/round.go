package round

import (
	"errors"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/ohlc"
)

const (
	LowerRoundNumberWeak   = "LowerRoundNumberWeak"
	LowerRoundNumberStrong = "LowerRoundNumberStrong"
	UpperRoundNumberWeak   = "UpperRoundNumberWeak"
	UpperRoundNumberStrong = "UpperRoundNumberStrong"
)

var ten = decimal.NewFromFloat(10)
var hundred = decimal.NewFromFloat(100)

type Number struct {
	latestCandle *ohlc.OHLC
}

func New() *Number {
	return &Number{}
}

func (rn *Number) Insert(o *ohlc.OHLC) {
	rn.latestCandle = o
}

func floor(number, multiplier decimal.Decimal) decimal.Decimal {
	return number.Mul(multiplier).Floor().Div(multiplier)
}

func ceil(number, multiplier decimal.Decimal) decimal.Decimal {
	return number.Mul(multiplier).Ceil().Div(multiplier)
}

func (rn *Number) Value() (map[string]float64, error) {
	if rn.latestCandle == nil {
		return nil, errors.New("price data is missing")
	}

	var m = map[string]float64{}
	var unit decimal.Decimal
	var multiplier decimal.Decimal

	if rn.latestCandle.Close.LessThan(decimal.NewFromFloat(1.00)) {
		m[LowerRoundNumberWeak], _ = floor(rn.latestCandle.Close, hundred).Float64()
		m[LowerRoundNumberStrong], _ = floor(rn.latestCandle.Close, ten).Float64()
		m[UpperRoundNumberWeak], _ = ceil(rn.latestCandle.Close, hundred).Float64()
		m[UpperRoundNumberStrong], _ = ceil(rn.latestCandle.Close, ten).Float64()
		return m, nil
	} else if rn.latestCandle.Close.LessThan(decimal.NewFromFloat(10.00)) {
		unit = decimal.NewFromFloat(1)
		multiplier = decimal.NewFromFloat(1)
	} else if rn.latestCandle.Close.LessThan(decimal.NewFromFloat(100.00)) {
		unit = decimal.NewFromFloat(10)
		multiplier = decimal.NewFromFloat(0.1)
	} else if rn.latestCandle.Close.LessThan(decimal.NewFromFloat(1000.00)) {
		unit = decimal.NewFromFloat(100)
		multiplier = decimal.NewFromFloat(0.01)
	} else if rn.latestCandle.Close.LessThan(decimal.NewFromFloat(10000.00)) {
		unit = decimal.NewFromFloat(1000)
		multiplier = decimal.NewFromFloat(0.01)
	} else {
		return nil, errors.New("not supported: price is too high")
	}

	{
		var lowerRoundNumberWeak = rn.latestCandle.Close.Mul(multiplier).Floor().Div(multiplier)
		m[LowerRoundNumberWeak], _ = lowerRoundNumberWeak.Float64()
	}

	{
		m[LowerRoundNumberStrong], _ = unit.Float64()
	}

	{
		var upperRoundNumberWeak = rn.latestCandle.Close.Mul(multiplier).Ceil().Div(multiplier)
		m[UpperRoundNumberWeak], _ = upperRoundNumberWeak.Float64()
	}

	{
		var upperRoundNumberStrong = unit.Mul(ten)
		m[UpperRoundNumberStrong], _ = upperRoundNumberStrong.Float64()
	}

	return m, nil
}

func (rn *Number) ValueResultKeys() []string {
	return []string{UpperRoundNumberStrong, UpperRoundNumberWeak, LowerRoundNumberStrong, LowerRoundNumberWeak}
}
