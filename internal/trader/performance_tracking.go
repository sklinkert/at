package trader

import (
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/helper"
	"gorm.io/gorm"
	"time"
)

type PerformanceRecord struct {
	gorm.Model
	BacktestingID              string
	StrategyName               string
	Strategy                   string
	Instrument                 string
	CandleDuration             time.Duration
	TargetInPips               float64
	StopLossInPips             float64
	PerformanceTrigger         float64
	TotalPerformanceInPips     float64
	AVGPerformanceInPips       float64
	MaxAggregateDrawdownInPips float64
	MaxLossInPips              float64
	MaxLossInPercent           float64
	MaxWinInPercent            float64
	MaxWinInPips               float64
	TradesWinRationInPercent   float64
	Trades                     int
	TradesWin                  int
	TradesLoss                 int
	TradesLossLong             int
	TradesLossShort            int
	TradesLong                 int
	TradesShort                int
	MaxConsecutiveTradesLoss   uint
	MaxConcurrentPositions     int
	GitRev                     string
	Duration                   string
	FirstTrade                 time.Time
	LastTrade                  time.Time
	AVGTradeDurationInSeconds  float64
	TotalExposureInPercent     float64
	ChartHTML                  string
	BacktestingConfigJSON      string
	ClosedPositions            []broker.Position `gorm:"-"`
	TotalTimeInMarket          time.Duration
	AVGTimeInMarket            time.Duration
}

const pipsFactor = 10000.0

func (tr *Trader) GetPerformanceRecords() ([]PerformanceRecord, error) {
	var records []PerformanceRecord
	if err := tr.gormDB.Model(&PerformanceRecord{}).Where("backtesting_id IS NOT NULL").Order("created_at DESC").Find(&records).Error; err != nil {
		return records, err
	}
	return records, nil
}

func (tr *Trader) GetPerformanceRecordByID(backtestingID string) (PerformanceRecord, error) {
	var record PerformanceRecord
	if err := tr.gormDB.Model(&PerformanceRecord{}).Where("backtesting_id = ?", backtestingID).First(&record).Error; err != nil {
		return record, err
	}
	return record, nil
}

func (tr *Trader) totalTimeInMarket(closedPositions []broker.Position) (timeInMarket time.Duration) {
	for _, position := range closedPositions {
		timeInMarket += position.Duration()
	}
	return
}

func (tr *Trader) avgTimeInMarket(closedPositions []broker.Position) (avgTimeInMarket time.Duration) {
	totalTime := tr.totalTimeInMarket(closedPositions)
	positions := int64(len(closedPositions))
	if positions > 0 {
		avgTimeInMarket = totalTime / time.Duration(positions)
	}
	return
}

func (tr *Trader) candleDuration() (duration time.Duration) {
	if len(tr.closedCandles) > 0 {
		return tr.closedCandles[0].Duration
	}
	return
}

func (tr *Trader) tradesCounter(closedPositions []broker.Position, direction broker.BuyDirection) int {
	var trades int
	for _, position := range closedPositions {
		if position.BuyDirection == direction {
			trades++
		}
	}
	return trades
}

func (tr *Trader) maxWinInPips(closedPositions []broker.Position) float64 {
	var maxWin float64
	for _, position := range closedPositions {
		perf := position.PerformanceAbsolute(decimal.Decimal{}, decimal.Decimal{})
		if perf > maxWin {
			maxWin = perf
		}
	}
	return maxWin * pipsFactor
}

func (tr *Trader) maxWinInPercent(closedPositions []broker.Position) float64 {
	var maxWin float64
	for _, position := range closedPositions {
		perf := position.PerformanceInPercentage(decimal.Decimal{}, decimal.Decimal{})
		if perf > maxWin {
			maxWin = perf
		}
	}
	return maxWin
}

func (tr *Trader) maxLossInPips(closedPositions []broker.Position) float64 {
	var maxLoss float64
	for _, position := range closedPositions {
		perf := position.PerformanceAbsolute(decimal.Decimal{}, decimal.Decimal{})
		if perf < maxLoss {
			maxLoss = perf
		}
	}
	return maxLoss * pipsFactor
}

func (tr *Trader) maxLossInPercent(closedPositions []broker.Position) float64 {
	var maxLoss float64
	for _, position := range closedPositions {
		perf := position.PerformanceInPercentage(decimal.Decimal{}, decimal.Decimal{})
		if perf < maxLoss {
			maxLoss = perf
		}
	}
	return maxLoss
}

func (tr *Trader) tradesLossCounter(closedPositions []broker.Position, direction broker.BuyDirection) int {
	var trades int
	for _, position := range closedPositions {
		if position.BuyDirection == direction {
			perf := position.PerformanceInPercentage(decimal.Decimal{}, decimal.Decimal{})
			if perf < 0 {
				trades++
			}
		}
	}
	return trades
}

func (tr *Trader) tradesWinCounter(closedPositions []broker.Position, direction broker.BuyDirection) int {
	var trades int
	for _, position := range closedPositions {
		if position.BuyDirection == direction {
			perf := position.PerformanceInPercentage(decimal.Decimal{}, decimal.Decimal{})
			if perf >= 0 {
				trades++
			}
		}
	}
	return trades
}

func (tr *Trader) getMaxConsecutiveLossTrades(closedPositions []broker.Position) uint {
	var maxConsecutiveTradesLoss, currentTradesLoss uint
	for _, position := range closedPositions {
		perf := position.PerformanceInPercentage(decimal.Decimal{}, decimal.Decimal{})
		if perf < 0 {
			currentTradesLoss++
		} else {
			if currentTradesLoss > maxConsecutiveTradesLoss {
				maxConsecutiveTradesLoss = currentTradesLoss
				currentTradesLoss = 0
			}
		}
	}

	if currentTradesLoss > maxConsecutiveTradesLoss {
		maxConsecutiveTradesLoss = currentTradesLoss
	}
	return maxConsecutiveTradesLoss
}

func (tr *Trader) totalPerfInPips(closedPositions []broker.Position) decimal.Decimal {
	var totalPerfInPips decimal.Decimal
	for _, position := range closedPositions {
		perf := helper.Cent2Pips(decimal.NewFromFloat(position.PerformanceAbsolute(decimal.Decimal{}, decimal.Decimal{})))
		totalPerfInPips = totalPerfInPips.Add(perf)
	}
	return totalPerfInPips
}

func (tr *Trader) GetPerformanceRecord(chartHTML string) (*PerformanceRecord, error) {
	closedPositions, err := tr.GetClosedPositions()
	if err != nil {
		log.WithError(err).Fatal("unable to get closed positions")
		return nil, err
	}

	if len(closedPositions) == 0 {
		return nil, nil
	}

	var totalPerfInPips = tr.totalPerfInPips(closedPositions)
	avgPerfInPips := totalPerfInPips.Div(decimal.NewFromFloat(float64(len(closedPositions))))
	avgPerfInPipsFloat, _ := avgPerfInPips.Float64()
	totalPerfInPipsFloat, _ := totalPerfInPips.Float64()
	maxAggregatedDrawdownFloat, _ := tr.MaxAggregatedDrawdownInPips.Float64()

	perf := &PerformanceRecord{
		BacktestingID:              tr.ID(),
		Instrument:                 tr.Instrument,
		StrategyName:               tr.strategy.Name(),
		Strategy:                   tr.strategy.String(),
		CandleDuration:             tr.candleDuration(),
		ChartHTML:                  chartHTML,
		TotalPerformanceInPips:     totalPerfInPipsFloat,
		AVGPerformanceInPips:       avgPerfInPipsFloat,
		Trades:                     len(closedPositions),
		TradesWin:                  tr.tradesWinCounter(closedPositions, broker.BuyDirectionLong) + tr.tradesWinCounter(closedPositions, broker.BuyDirectionShort),
		TradesLoss:                 tr.tradesLossCounter(closedPositions, broker.BuyDirectionLong) + tr.tradesLossCounter(closedPositions, broker.BuyDirectionShort),
		TradesLong:                 tr.tradesCounter(closedPositions, broker.BuyDirectionLong),
		TradesShort:                tr.tradesCounter(closedPositions, broker.BuyDirectionShort),
		TradesLossLong:             tr.tradesLossCounter(closedPositions, broker.BuyDirectionLong),
		TradesLossShort:            tr.tradesLossCounter(closedPositions, broker.BuyDirectionShort),
		MaxLossInPips:              tr.maxLossInPips(closedPositions),
		MaxLossInPercent:           tr.maxLossInPercent(closedPositions),
		MaxWinInPercent:            tr.maxWinInPercent(closedPositions),
		MaxWinInPips:               tr.maxWinInPips(closedPositions),
		MaxAggregateDrawdownInPips: maxAggregatedDrawdownFloat,
		MaxConcurrentPositions:     tr.maxConcurrentPositions,
		GitRev:                     tr.gitRev,
		FirstTrade:                 closedPositions[0].BuyTime,
		LastTrade:                  closedPositions[len(closedPositions)-1].BuyTime,
		Duration:                   time.Since(tr.StartTime).String(),
		MaxConsecutiveTradesLoss:   tr.getMaxConsecutiveLossTrades(closedPositions),
		ClosedPositions:            closedPositions,
		TotalTimeInMarket:          tr.totalTimeInMarket(closedPositions),
		AVGTimeInMarket:            tr.avgTimeInMarket(closedPositions),
		AVGTradeDurationInSeconds:  tr.totalTimeInMarket(closedPositions).Seconds() / float64(len(closedPositions)),
	}
	perf.TradesWinRationInPercent = float64(perf.TradesWin) * 100 / float64(perf.Trades)
	perf.TotalExposureInPercent = tr.totalExposureInPercent(perf.TotalTimeInMarket, perf.FirstTrade, perf.LastTrade)

	if (perf.TradesWin + perf.TradesLoss) != perf.Trades {
		return nil, fmt.Errorf("TradesWin(%d) + TradesLoss(%d) != Trades(%d)", perf.TradesWin, perf.TradesLoss, perf.Trades)
	}
	if (perf.TradesLong + perf.TradesShort) != perf.Trades {
		return nil, fmt.Errorf("TradesLong(%d) + TradesShort(%d) != Trades(%d)", perf.TradesLong, perf.TradesShort, perf.Trades)
	}

	return perf, nil
}

func (tr *Trader) totalExposureInPercent(totalTimeInMarket time.Duration, firstPrice, lastPrice time.Time) float64 {
	var totalTime = lastPrice.Sub(firstPrice)
	return float64(totalTimeInMarket) * 100 / float64(totalTime)
}

func (tr *Trader) Summary() {
	pr, err := tr.GetPerformanceRecord("")
	if err != nil {
		log.WithError(err).Error("Cannot get performance record")
		return
	}

	log.Infof("%25s: %s", "Instrument", pr.Instrument)
	log.Infof("%25s: %s", "Strategy", pr.Strategy)
	log.Infof("%25s: %s", "Candle duration", pr.CandleDuration)
	log.Infof("%25s: %s -> %s", "Period", pr.FirstTrade.Format("02.01.2006"), pr.LastTrade.Format("02.01.2006"))
	log.Infof("%25s: %d (%d long, %d short)", "Total positions", pr.Trades, pr.TradesLong, pr.TradesShort)
	log.Infof("%25s: %s (%.2f%%)", "Total time in market", pr.TotalTimeInMarket, pr.TotalExposureInPercent)
	log.Infof("%25s: %s", "AVG time in market", pr.AVGTimeInMarket)
	log.Infof("%25s: %d (%.2f%%)", "Profit positions", pr.TradesWin, pr.TradesWinRationInPercent)
	log.Infof("%25s: %d", "Loss positions", pr.TradesLoss)
	log.Infof("%25s: %d", "Loss positions long", pr.TradesLossLong)
	log.Infof("%25s: %d", "Loss positions short", pr.TradesLossShort)
	log.Infof("%25s: %.2f%% %.2f (%.2f pips)", "Max win", pr.MaxWinInPercent, pr.MaxWinInPips/pipsFactor, pr.MaxWinInPips)
	log.Infof("%25s: %.2f%% %.2f (%.2f pips)", "Max loss", pr.MaxLossInPercent, pr.MaxLossInPips/pipsFactor, pr.MaxLossInPips)
	log.Infof("%25s: %.2f (%.2f pips)", "Total performance", pr.TotalPerformanceInPips/pipsFactor, pr.TotalPerformanceInPips)
	log.Infof("%25s: %.2f (%.2f pips)", "AVG Performance", pr.AVGPerformanceInPips/pipsFactor, pr.AVGPerformanceInPips)
}

func (tr *Trader) SavePerformanceRecord(chartHTML string) error {
	performanceRecord, err := tr.GetPerformanceRecord(chartHTML)
	if err != nil {
		return err
	}

	if err := tr.gormDB.Create(&performanceRecord).Error; err != nil {
		log.WithError(err).Fatal("Cannot save PerformanceRecord to DB")
	}

	for _, pos := range performanceRecord.ClosedPositions {
		pos.PerformanceRecordID = performanceRecord.ID
		if err := tr.gormDB.Create(&pos).Error; err != nil {
			log.WithError(err).Error("Cannot save closed position to DB")
		}
	}
	return nil
}
