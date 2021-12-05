package paperwallet

import (
	"github.com/AMekss/assert"
	"github.com/go-test/deep"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/tick"
	"testing"
	"time"
)

func TestBuyAndSellLong(t *testing.T) {
	b := New()
	bid := decimal.NewFromFloat(0.9)
	ask := decimal.NewFromFloat(1.0)
	now := time.Now()
	b.currentTick = tick.New("", now, bid, ask)

	openPositions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t, 0, len(openPositions))
	targetPrice := decimal.NewFromFloat(2.0)
	stopLossPrice := decimal.NewFromFloat(0.5)
	order := broker.NewMarketOrder(broker.BuyDirectionLong, 1.00, "", targetPrice, stopLossPrice)

	_, err = b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(positions))

	pos := positions[0]
	assert.True(t, broker.BuyDirectionLong == pos.BuyDirection)
	assert.True(t, now == pos.BuyTime)
	assert.True(t, pos.SellPrice.Equals(decimal.NewFromFloat(0)))

	// Sell position
	bid = decimal.NewFromFloat(2.0)
	ask = decimal.NewFromFloat(3.0)
	now = now.Add(time.Minute)
	b.currentTick = tick.New("", now, bid, ask)

	err = b.Sell(pos)
	assert.NoError(t.Fatalf, err)

	// Update pos after Sell()
	closedPositions, err := b.GetClosedPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(closedPositions))
	pos = closedPositions[0]

	assertDecimal(t, bid, pos.SellPrice)

	perfAbs := pos.PerformanceAbsolute(decimal.NewFromFloat(0), decimal.NewFromFloat(0)) // must take SellPrice instead of given currentPrice
	assert.EqualFloat64(t, 1, perfAbs)

	perfPercent := pos.PerformanceInPercentage(decimal.NewFromFloat(0), decimal.NewFromFloat(0)) // must take SellPrice instead of given currentPrice
	assert.EqualFloat64(t, 100, perfPercent)
	assert.True(t, now == pos.SellTime)
}

func TestLimitOrderLong(t *testing.T) {
	b := New()
	ask := decimal.NewFromFloat(2.0)
	bid := decimal.NewFromFloat(2.0)
	now := time.Now()
	b.currentTick = tick.New("", now, bid, ask)

	targetPrice := decimal.NewFromFloat(2.5)
	stopLossPrice := decimal.NewFromFloat(0.5)
	limitPrice := decimal.NewFromFloat(1.95)
	size := 1.0
	order := broker.NewLimitOrder(broker.BuyDirectionLong, size, "", targetPrice, stopLossPrice, limitPrice)

	_, err := b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 0, len(positions))

	ask = limitPrice
	bid = decimal.NewFromFloat(1.90)
	tickData := tick.New("", now, bid, ask)
	b.SetCurrenctPrice(tickData)

	positions, err = b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(positions))
	pos := positions[0]
	assert.True(t, pos.BuyPrice.Equals(ask))
	assert.EqualFloat64(t, size, pos.Size)
	assert.True(t, broker.BuyDirectionLong == pos.BuyDirection)
}

func TestLimitOrderShort(t *testing.T) {
	b := New()
	ask := decimal.NewFromFloat(2.0)
	bid := decimal.NewFromFloat(2.0)
	now := time.Now()
	b.currentTick = tick.New("", now, bid, ask)

	targetPrice := decimal.NewFromFloat(1.5)
	stopLossPrice := decimal.NewFromFloat(2.5)
	limitPrice := decimal.NewFromFloat(2.05)
	size := 1.0
	order := broker.NewLimitOrder(broker.BuyDirectionShort, size, "", targetPrice, stopLossPrice, limitPrice)

	_, err := b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 0, len(positions))

	ask = decimal.NewFromFloat(2.10)
	bid = limitPrice
	tickData := tick.New("", now, bid, ask)
	b.SetCurrenctPrice(tickData)

	positions, err = b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(positions))
	pos := positions[0]
	assert.True(t, pos.BuyPrice.Equals(bid))
	assert.EqualFloat64(t, size, pos.Size)
	assert.True(t, broker.BuyDirectionShort == pos.BuyDirection)
}

func TestBuyAndSellShort(t *testing.T) {
	b := New()
	bid := decimal.NewFromFloat(2.0)
	ask := decimal.NewFromFloat(0.9)
	now := time.Now()
	b.currentTick = tick.New("", now, bid, ask)

	openPositions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t, 0, len(openPositions))
	targetPrice := decimal.NewFromFloat(0.5)
	stopLossPrice := decimal.NewFromFloat(2.0)
	order := broker.NewMarketOrder(broker.BuyDirectionShort, 1.00, "", targetPrice, stopLossPrice)

	_, err = b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(positions))

	pos := positions[0]
	assert.True(t, broker.BuyDirectionShort == pos.BuyDirection)
	assert.True(t, now == pos.BuyTime)
	assert.True(t, pos.SellPrice.Equals(decimal.NewFromFloat(0)))

	// Sell position
	bid = decimal.NewFromFloat(1.77)
	ask = decimal.NewFromFloat(1.0)
	now = now.Add(time.Minute)
	b.currentTick = tick.New("", now, bid, ask)

	err = b.Sell(pos)
	assert.NoError(t.Fatalf, err)

	// Update pos after Sell()
	closedPositions, err := b.GetClosedPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 1, len(closedPositions))
	pos = closedPositions[0]

	assertDecimal(t, ask, pos.SellPrice)

	perfAbs := pos.PerformanceAbsolute(decimal.NewFromFloat(0), decimal.NewFromFloat(0)) // must take SellPrice instead of given currentPrice
	assert.EqualFloat64(t, 1, perfAbs)

	perfPercent := pos.PerformanceInPercentage(decimal.NewFromFloat(0), decimal.NewFromFloat(0)) // must take SellPrice instead of given currentPrice
	assert.EqualFloat64(t, 100, perfPercent)
	assert.True(t, now == pos.SellTime)
}

func TestBacktest_GetOpenPositions(t *testing.T) {
	b := New()
	order := broker.NewMarketOrder(broker.BuyDirectionLong, 1.00, "", decimal.Zero, decimal.Zero)
	_, err := b.Buy(order)
	assert.NoError(t.Fatalf, err)
	_, err = b.Buy(order)
	assert.NoError(t.Fatalf, err)
	_, err = b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 3, len(positions))
	pos := positions[0]

	err = b.Sell(pos)
	assert.NoError(t.Fatalf, err)
	openPositions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t, 2, len(openPositions))
}

func TestBacktest_GetClosedPositions(t *testing.T) {
	b := New()
	order := broker.NewMarketOrder(broker.BuyDirectionLong, 1.00, "", decimal.Zero, decimal.Zero)
	_, _ = b.Buy(order)
	_, _ = b.Buy(order)
	_, err := b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 3, len(positions))
	pos := positions[0]

	err = b.Sell(pos)
	assert.NoError(t.Fatalf, err)
	openPositions, err := b.GetClosedPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t, 1, len(openPositions))
}

func TestBacktest_GetOpenPosition(t *testing.T) {
	b := New()
	order := broker.NewMarketOrder(broker.BuyDirectionLong, 1.00, "", decimal.Zero, decimal.Zero)
	_, err := b.Buy(order)
	assert.NoError(t.Fatalf, err)

	_, err = b.Buy(order)
	assert.NoError(t.Fatalf, err)

	positions, err := b.GetOpenPositions()
	assert.NoError(t.Fatalf, err)
	assert.EqualInt(t.Fatalf, 2, len(positions))
	pos := positions[0]

	openPosition, err := b.GetOpenPosition(pos.Reference)
	assert.NoError(t.Fatalf, err)
	if diff := deep.Equal(pos, openPosition); diff != nil {
		t.Error(diff)
	}

	err = b.Sell(pos)
	assert.NoError(t.Fatalf, err)
	_, err = b.GetOpenPosition(pos.Reference)
	assert.EqualErrors(t, broker.ErrPositionNotFound, err)
}

func Test_getBuyPriceByDirection(t *testing.T) {
	bid := decimal.NewFromFloat(1.0)
	ask := decimal.NewFromFloat(1.1)
	b := New()
	b.currentTick = tick.New("test", time.Now(), bid, ask)

	price := b.getBuyPriceByDirection(broker.BuyDirectionLong)
	assertDecimal(t, ask, price)

	price = b.getBuyPriceByDirection(broker.BuyDirectionShort)
	assertDecimal(t, bid, price)
}

func Test_getSellPriceByDirection(t *testing.T) {
	bid := decimal.NewFromFloat(1.0)
	ask := decimal.NewFromFloat(1.1)
	b := New()
	b.currentTick = tick.New("test", time.Now(), bid, ask)

	price := b.getSellPriceByDirection(broker.BuyDirectionLong, true)
	assertDecimal(t, bid, price)

	price = b.getSellPriceByDirection(broker.BuyDirectionShort, true)
	assertDecimal(t, ask, price)

	// without slippage
	price = b.getSellPriceByDirection(broker.BuyDirectionLong, false)
	assertDecimal(t, bid, price)

	price = b.getSellPriceByDirection(broker.BuyDirectionShort, false)
	assertDecimal(t, ask, price)
}

func assertDecimal(t *testing.T, want, got decimal.Decimal) {
	wantFloat, _ := want.Float64()
	gotFloat, _ := got.Float64()
	assert.EqualFloat64(t, wantFloat, gotFloat)
}

func Test_closeAllOpenPositions(t *testing.T) {
	b := New()
	order := broker.NewMarketOrder(broker.BuyDirectionLong, 1.00, "", decimal.Zero, decimal.Zero)
	_, _ = b.Buy(order)
	_, _ = b.Buy(order)
	_, _ = b.Buy(order)
	assert.EqualInt(t, 3, len(b.openPositions))
	b.CloseAllOpenPositions()
	assert.EqualInt(t, 0, len(b.openPositions))
}

func TestBacktest_getTotalPerf(t *testing.T) {
	b := New()
	b.closedPositions = map[string]broker.Position{
		"1": {
			Size:      1,
			BuyPrice:  decimal.NewFromFloat(100),
			SellPrice: decimal.NewFromFloat(110),
		},
		"2": {
			Size:      2,
			BuyPrice:  decimal.NewFromFloat(1),
			SellPrice: decimal.NewFromFloat(20),
		},
	}
	assert.EqualFloat64(t, 48, getTotalPerf(b.closedPositions))
}

func TestBacktest_getTotalLossPositions(t *testing.T) {
	b := New()
	b.closedPositions = map[string]broker.Position{
		"1": {
			Size:      1,
			BuyPrice:  decimal.NewFromFloat(100),
			SellPrice: decimal.NewFromFloat(90),
		},
		"2": {
			Size:      2,
			BuyPrice:  decimal.NewFromFloat(1),
			SellPrice: decimal.NewFromFloat(20),
		},
	}
	assert.EqualInt(t, 1, getTotalLossPositions(b.closedPositions))
}

//func TestBuyCheckTargetAndStopLoss(t *testing.T) {
//	var testCases = []struct {
//		bid                  decimal.Decimal
//		ask                  decimal.Decimal
//		direction            broker.BuyDirection
//		stopPrice            decimal.Decimal
//		targetPrice          decimal.Decimal
//		expectedErrorMessage string
//	}{
//		{
//			bid:                  decimal.NewFromFloat(1.00), // buy price
//			ask:                  decimal.NewFromFloat(1.50), // sell price
//			direction:            broker.BuyDirectionLong,
//			stopPrice:            decimal.NewFromFloat(0.80),
//			targetPrice:          decimal.NewFromFloat(1.30),
//			expectedErrorMessage: "target is below current price",
//		},
//		{
//			bid:                  decimal.NewFromFloat(1.00), // buy price
//			ask:                  decimal.NewFromFloat(1.50), // sell price
//			direction:            broker.BuyDirectionLong,
//			stopPrice:            decimal.NewFromFloat(1.80),
//			targetPrice:          decimal.NewFromFloat(2.00),
//			expectedErrorMessage: "current price is below stop loss",
//		},
//		{
//			bid:                  decimal.NewFromFloat(1.00), // buy price
//			ask:                  decimal.NewFromFloat(1.50), // sell price
//			direction:            broker.BuyDirectionShort,
//			stopPrice:            decimal.NewFromFloat(2.00),
//			targetPrice:          decimal.NewFromFloat(1.20),
//			expectedErrorMessage: "target is above current price",
//		},
//		{
//			bid:                  decimal.NewFromFloat(1.00), // buy price
//			ask:                  decimal.NewFromFloat(1.50), // sell price
//			direction:            broker.BuyDirectionShort,
//			stopPrice:            decimal.NewFromFloat(0.80),
//			targetPrice:          decimal.NewFromFloat(0.50),
//			expectedErrorMessage: "current price is above stop loss",
//		},
//	}
//
//	for _, testCase := range testCases {
//		currentTick := tick.New("", time.Now(), testCase.bid, testCase.ask)
//		b := New()
//		b.currentTick = currentTick
//		order := broker.NewMarketOrder(testCase.direction, 1.00, "", testCase.targetPrice, testCase.stopPrice, "")
//		_, err := b.Buy(order)
//		assert.ErrorIncludesMessage(t, testCase.expectedErrorMessage, err)
//	}
//}

func Test_getAbsoluteTradingFee(t *testing.T) {
	b := New()
	b.tradingFeePercent = decimal.NewFromFloat(0.26)
	price := decimal.NewFromFloat(43.50)
	tradingFee := decimal.NewFromFloat(0.1131)
	assertDecimal(t, tradingFee, b.getAbsoluteTradingFee(price))

	// Buy: Long
	b.currentTick = tick.Tick{
		Ask: price,
	}
	wantLong := tradingFee.Add(price)
	assertDecimal(t, wantLong, b.getBuyPriceByDirection(broker.BuyDirectionLong))

	// Buy: Short
	b.currentTick = tick.Tick{
		Bid: price,
	}
	wantShort := price.Sub(tradingFee)
	assertDecimal(t, wantShort, b.getBuyPriceByDirection(broker.BuyDirectionShort))

	// Sell: Long
	b.currentTick = tick.Tick{
		Bid: price,
	}
	wantLong = price.Sub(tradingFee)
	assertDecimal(t, wantLong, b.getSellPriceByDirection(broker.BuyDirectionLong, true))

	// Sell: Short
	b.currentTick = tick.Tick{
		Ask: price,
	}
	wantShort = price.Add(tradingFee)
	assertDecimal(t, wantShort, b.getSellPriceByDirection(broker.BuyDirectionShort, true))
}
