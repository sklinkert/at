package paperwallet

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
)

func (pw *Paperwallet) PrintSummary() {
	positions, _ := pw.GetClosedPositions()
	if len(positions) == 0 {
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
