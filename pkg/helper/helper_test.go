package helper

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"testing"
)

func TestCent2Pips(t *testing.T) {
	if !Cent2Pips(decimal.NewFromFloat(0.001)).Equals(decimal.NewFromFloat(10)) {
		t.Errorf("Cent2Pips does not work properly")
	}
}

func TestPips2Cent(t *testing.T) {
	if !Pips2Cent(decimal.NewFromFloat(10)).Equals(decimal.NewFromFloat(0.001)) {
		t.Errorf("Pips2Cent does not work properly")
	}
}

func TestPipHelper(t *testing.T) {
	n := decimal.NewFromFloat(1.00)
	assert.EqualStrings(t, n.String(), Pips2Cent(Cent2Pips(n)).String())
	n = decimal.NewFromFloat(1.87)
	assert.EqualStrings(t, n.String(), Pips2Cent(Cent2Pips(n)).String())
}

func TestCalcStopLossPriceByPercentage(t *testing.T) {
	price := decimal.NewFromFloat(100)
	stopLossPercentage := decimal.NewFromFloat(20)

	// Long
	stopPrice := CalcStopLossPriceByPercentage(price, stopLossPercentage, broker.BuyDirectionLong)
	stopPriceFloat, _ := stopPrice.Float64()
	assert.EqualFloat64(t, 80, stopPriceFloat)

	// Short
	stopPrice = CalcStopLossPriceByPercentage(price, stopLossPercentage, broker.BuyDirectionShort)
	stopPriceFloat, _ = stopPrice.Float64()
	assert.EqualFloat64(t, 120, stopPriceFloat)
}

func TestTargetPriceByPercentage(t *testing.T) {
	price := decimal.NewFromFloat(100)
	targetPercentage := decimal.NewFromFloat(20)

	// Long
	targetPrice := CalcTargetPriceByPercentage(price, targetPercentage, broker.BuyDirectionLong)
	targetPriceFloat, _ := targetPrice.Float64()
	assert.EqualFloat64(t, 120, targetPriceFloat)

	// Short
	targetPrice = CalcTargetPriceByPercentage(price, targetPercentage, broker.BuyDirectionShort)
	targetPriceFloat, _ = targetPrice.Float64()
	assert.EqualFloat64(t, 80, targetPriceFloat)
}

func TestDecimalToFloat(t *testing.T) {
	n := decimal.NewFromFloat(10.34)
	assert.EqualFloat64(t, 10.34, DecimalToFloat(n))
}

func TestFloatToDecimal(t *testing.T) {
	n := 10.34
	want := decimal.NewFromFloat(n)
	assert.True(t, want.Equals(FloatToDecimal(n)))
}
