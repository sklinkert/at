package backtest

import (
	"encoding/csv"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/helper"
	"os"
	"time"
)

func (b *Backtest) writeCSV() {
	b.RLock()
	defer b.RUnlock()

	file, err := os.Create("./results/backtesting_result.csv")
	if err != nil {
		log.WithError(err).Error("creating CSV file failed")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.WithError(err).Warn("file.Close() failed")
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"#",
		"Weekday",
		"BuyTime",
		"SellTime",
		"Direction",
		"Size",
		"BuyPrice",
		"SellPrice",
		"TargetPrice",
		"StopLossPrice",
		"TargePips",
		"StopLossPips",
		"PerformanceInPips",
		"TotalPerformanceInPips",
		"MaxSurgePips",
		"MaxDrawdownPips",
		"Duration",
		"TodayPerf",
		//"OHLCAgeOnBuy",
		"GapToSMA",
	}
	if err := writer.Write(header); err != nil {
		log.WithError(err).Fatal("cannot write to file")
	}

	closedPositions, _ := b.GetClosedPositions()
	csvPrintPosition(writer, closedPositions)

	openPositions, _ := b.GetOpenPositions()
	csvPrintPosition(writer, openPositions)
}

func csvPrintPosition(writer *csv.Writer, positions []broker.Position) {
	var totalPerfPips decimal.Decimal

	locBerlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.WithError(err).Fatal("cannot load time zone")
	}

	for i, position := range positions {
		targetInPips := helper.Cent2Pips(position.TargetPrice.Sub(position.BuyPrice)).Round(2)
		stopLossInPips := helper.Cent2Pips(position.BuyPrice.Sub(position.StopLossPrice)).Round(2)
		if position.BuyDirection == broker.BuyDirectionShort {
			targetInPips = targetInPips.Neg()
			stopLossInPips = stopLossInPips.Neg()
		}

		perfAbs := position.PerformanceAbsolute(position.SellPrice, position.SellPrice)
		perfPips := helper.Cent2Pips(decimal.NewFromFloat(perfAbs))
		totalPerfPips = totalPerfPips.Add(perfPips)

		record := []string{
			fmt.Sprintf("%d", i+1),
			position.BuyTime.In(locBerlin).Weekday().String(),
			position.BuyTime.In(locBerlin).Format("2006-01-02 15:04:05"),
			position.SellTime.In(locBerlin).Format("2006-01-02 15:04:05"),
			position.BuyDirection.String(),
			fmt.Sprintf("%.1f", position.Size),
			position.BuyPrice.Round(5).String(),
			position.SellPrice.Round(5).String(),
			position.TargetPrice.Round(5).String(),
			position.StopLossPrice.Round(5).String(),
			targetInPips.String(),
			stopLossInPips.String(),
			perfPips.Round(2).String(),
			totalPerfPips.Round(2).String(),
			fmt.Sprintf("%.2f", position.MaxSurge),
			fmt.Sprintf("%.2f", position.MaxDrawdown),
			position.Duration().String(),
			position.TodayPerformanceInPercent.Round(2).String(),
			//position.OHLCAgeOnBuy.String(),
			position.GapToSMA.Round(5).String(),
		}
		if err := writer.Write(record); err != nil {
			log.WithError(err).Fatal("cannot write to file")
		}
	}
}
