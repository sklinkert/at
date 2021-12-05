package trader

import (
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/helper"
	"sort"
)

// getTotalPerformanceInPips return total performance of closed positions
//func (tr *Trader) getTotalPerformanceInPips() (decimal.Decimal, error) {
//	var positions, err = tr.broker.GetClosedPositions()
//	if err != nil {
//		return decZero, err
//	}
//	var totalPerfPips decimal.Decimal
//	for _, position := range positions {
//		targetInPips := helper.Cent2Pips(position.TargetPrice.Sub(position.BuyPrice)).Round(2)
//		stopLossInPips := helper.Cent2Pips(position.BuyPrice.Sub(position.StopLossPrice)).Round(2)
//		if position.BuyDirection == broker.BuyDirectionShort {
//			targetInPips = targetInPips.Neg()
//			stopLossInPips = stopLossInPips.Neg()
//		}
//
//		perfAbs := position.PerformanceAbsolute(position.SellPrice, position.SellPrice)
//		perfPips := helper.Cent2Pips(decimal.NewFromFloat(perfAbs))
//		totalPerfPips = totalPerfPips.Add(perfPips)
//	}
//	return totalPerfPips, nil
//}

var dec100 = decimal.NewFromFloat(100)

// distanceInPercentage - distance between price1 and price2 in %
func distanceInPercentage(price1, price2 decimal.Decimal) decimal.Decimal {
	if price1.IsZero() {
		return decimal.Zero
	}
	return price2.Sub(price1).Div(price1).Mul(dec100)
}

//func (tr *Trader) gapToSMAInPercent(currentTick tick.Tick) (decimal.Decimal, error) {
//	price := currentTick.Ask.Add(currentTick.Bid).Div(decimal.NewFromFloat(2))
//	sma, err := tr.sma.Average()
//	if err != nil {
//		return decZero, err
//	}
//	return distanceInPercentage(decimal.NewFromFloat(sma), price), nil
//}

func (tr *Trader) printPositionPerformanceByNotes() {
	closedPositions, _ := tr.GetClosedPositions()

	var perfPositionsByNote = map[string]float64{}
	for _, position := range closedPositions {
		perfInPips := helper.Cent2Pips(decimal.NewFromFloat(position.PerformanceAbsolute(decimal.Zero, decimal.Zero)))
		perfInPipsFloat, _ := perfInPips.Float64()
		key := fmt.Sprintf("%d-%s", position.BuyTime.Year(), position.Reference)
		perfPositionsByNote[key] += perfInPipsFloat
	}
	var sortedKeys []string
	for key := range perfPositionsByNote {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for _, note := range sortedKeys {
		totalProfit := perfPositionsByNote[note]
		log.Infof("%25s: %s %.2f pips", "Total profit for positions with note", note, totalProfit)
	}
}
