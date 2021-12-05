package broker

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"time"
)

type Position struct {
	PerformanceRecordID uint // foreign key
	Reference           string
	Instrument          string
	BuyPrice            decimal.Decimal
	BuyTime             time.Time
	BuyDirection        BuyDirection
	SellPrice           decimal.Decimal
	SellTime            time.Time
	TargetPrice         decimal.Decimal
	StopLossPrice       decimal.Decimal
	Size                float64
	OHLCAgeOnBuy        time.Duration
	CandleBuyTime       time.Time
	CandleSellTime      time.Time

	// Backtesting
	MaxSurge                  float64 // Pips
	MaxDrawdown               float64 // Pips
	TodayPerformanceInPercent decimal.Decimal
	GapToSMA                  decimal.Decimal
}

// Duration - returns duration of position
func (p *Position) Duration() time.Duration {
	return p.SellTime.Sub(p.BuyTime)
}

// MergePositions merges two position slices
func MergePositions(positions1, positions2 []Position) []Position {
	return append(positions1, positions2...)
}

func (p *Position) PerformanceAbsolute(bid, ask decimal.Decimal) float64 {
	var abs decimal.Decimal
	if p.SellPrice.IsZero() {
		//if bid.IsZero() || ask.IsZero() {
		//	log.Panicf("bid/ask must not be empty when position.SellPrice is empty as well: %+v", p)
		//}
		switch p.BuyDirection {
		case BuyDirectionLong:
			abs = bid.Sub(p.BuyPrice)
		case BuyDirectionShort:
			abs = p.BuyPrice.Sub(ask)
		}
	} else {
		switch p.BuyDirection {
		case BuyDirectionLong:
			abs = p.SellPrice.Sub(p.BuyPrice)
		case BuyDirectionShort:
			abs = p.BuyPrice.Sub(p.SellPrice)
		}
	}
	abs = abs.Mul(decimal.NewFromFloat(p.Size))
	absFloat, _ := abs.Float64()
	return absFloat
}

// PerformanceInPercentagePretty returns performance for closed positions, round to 2 decimal places
func (p *Position) PerformanceInPercentagePretty() float64 {
	perf := p.PerformanceInPercentage(decimal.Zero, decimal.Zero)
	return math.Round(perf*100) / 100
}

func (p *Position) PerformanceInPercentage(bid, ask decimal.Decimal) float64 {
	var percentage decimal.Decimal
	if p.SellPrice.IsZero() {
		switch p.BuyDirection {
		case BuyDirectionLong:
			percentage = bid.Sub(p.BuyPrice).Div(p.BuyPrice)
		case BuyDirectionShort:
			if ask.IsZero() {
				return 0
			}
			percentage = p.BuyPrice.Sub(ask).Div(ask)
		}
		percentage = percentage.Mul(decimal.NewFromFloat(100))
	} else {
		switch p.BuyDirection {
		case BuyDirectionLong:
			percentage = p.SellPrice.Sub(p.BuyPrice).Div(p.BuyPrice)
		case BuyDirectionShort:
			percentage = p.BuyPrice.Sub(p.SellPrice).Div(p.SellPrice)
		}
		percentage = percentage.Mul(decimal.NewFromFloat(100))
	}
	percentageFloat, _ := percentage.Float64()
	return percentageFloat
}

func (p *Position) Age(now time.Time) time.Duration {
	return now.Sub(p.BuyTime)
}

func (p *Position) String() string {
	return fmt.Sprintf("%s/%s: Direction=%s BuyLevel=%s BuyTime=%s Size=%.2f",
		p.Instrument, p.Reference, p.BuyDirection, p.BuyPrice, p.BuyTime, p.Size)
}
