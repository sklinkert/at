package paperwallet

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
)

func (pw *Paperwallet) printOpenPositionSummary(position *broker.Position) {
	sizeDec := decimal.NewFromFloat(position.Size)
	entryAmount := position.BuyPrice.Mul(sizeDec)
	currentPrice := decimal.Decimal{}
	if position.BuyDirection == broker.BuyDirectionLong {
		currentPrice = pw.currentTick.Bid
	} else {
		currentPrice = pw.currentTick.Ask
	}
	nowAmount := currentPrice.Mul(sizeDec)
	profit := nowAmount.Sub(entryAmount)
	log.Infof("Position: Direction=%s Size=%f EntryLevel=%5s Now=%5s -> %5s",
		position.BuyDirection, position.Size, position.BuyPrice, currentPrice, profit)
}

func (pw *Paperwallet) PrintSummary() {
	openPositions, err := pw.GetOpenPositions()
	if len(openPositions) > 0 && err != nil {
		for _, position := range openPositions {
			pw.printOpenPositionSummary(&position)
		}
	}

	positions, err := pw.GetClosedPositions()
	if len(positions) == 0 || err != nil {
		return
	}

	avgTradingFee := pw.totalTradingFee.Div(decimal.NewFromFloat(float64(len(pw.closedPositions))))
	log.Infof("%25s: %s (%s avg)", "Total trading fee", pw.totalTradingFee.Round(2), avgTradingFee.Round(4))
	log.Infof("%25s: %s", "Initial balance", pw.GetInitialBalance())
	log.Infof("%25s: %s", "End Balance", pw.GetBalance())
}

func getTotalPerf(closedPositions map[string]broker.Position) (totalPerf float64) {
	for _, position := range closedPositions {
		totalPerf += position.PerformanceAbsolute(decimal.Zero, decimal.Zero)
	}
	return
}
