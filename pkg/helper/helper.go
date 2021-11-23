package helper

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
)

// Pips2Cent - Convert pips to cent
func Pips2Cent(n decimal.Decimal) decimal.Decimal {
	return n.Div(decimal.NewFromFloat(10000))
}

// Cent2Pips - Convert cent to pips
func Cent2Pips(n decimal.Decimal) decimal.Decimal {
	return n.Mul(decimal.NewFromFloat(10000))
}

func FloatToDecimal(n float64) decimal.Decimal {
	return decimal.NewFromFloat(n)
}

func DecimalToFloat(n decimal.Decimal) float64 {
	f, _ := n.Float64()
	return f
}

func CalcStopLossPriceByPercentage(price, percentage decimal.Decimal, orderDirection broker.BuyDirection) decimal.Decimal {
	percentFrom := price.Div(decimal.NewFromFloat(100)).Mul(percentage)

	switch orderDirection {
	case broker.BuyDirectionLong:
		return price.Sub(percentFrom).Round(6)
	case broker.BuyDirectionShort:
		return price.Add(percentFrom).Round(6)
	default:
		log.Panicf("Unexpected order direction %q", orderDirection.String())
	}

	// Never reached
	return decimal.Zero
}

func CalcTargetPriceByPercentage(price, percentage decimal.Decimal, orderDirection broker.BuyDirection) decimal.Decimal {
	percentFrom := price.Div(decimal.NewFromFloat(100)).Mul(percentage)

	switch orderDirection {
	case broker.BuyDirectionLong:
		return price.Add(percentFrom).Round(6)
	case broker.BuyDirectionShort:
		return price.Sub(percentFrom).Round(6)
	default:
		log.Panicf("Unexpected order direction %q", orderDirection.String())
	}

	// Never reached
	return decimal.Zero
}

func SlippageAbsolute(expectedPrice, realPrice decimal.Decimal) decimal.Decimal {
	return realPrice.Sub(expectedPrice)
}

func GetMedian(n []float64) float64 {
	return GetPercentile(n, 50)
}

func GetPercentile(n []float64, percentile int) float64 {
	var pos = int(float64(len(n)) / float64(100) * float64(percentile))
	if pos < 1 {
		pos = 1
	}
	return n[pos-1]
}
