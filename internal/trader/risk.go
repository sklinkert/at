package trader

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/tick"
)

const (
//minDayPerformancePercentageLong  = 0.05
//minDayPerformancePercentageShort = -0.05
)

//var locEST *time.Location
//var noTradingDays []date.Date

func init() {
	//var err error
	//locEST, err = time.LoadLocation("EST")
	//if err != nil {
	//	log.WithError(err).Fatal("risk: unable to load EST timezone")
	//}
	//
	//noTradingDays = []date.Date{
	//	// US election days
	//	date.New(2016, 11, 9),
	//	date.New(2020, 11, 3),
	//	date.New(2020, 11, 4),
	//
	//	// ECB meetings
	//	date.New(2016, 1, 21),
	//	date.New(2016, 3, 10),
	//	date.New(2016, 4, 21),
	//	date.New(2016, 6, 2),
	//	date.New(2016, 7, 21),
	//	date.New(2016, 9, 8),
	//	date.New(2016, 10, 20),
	//	date.New(2016, 12, 8),
	//	date.New(2017, 1, 19),
	//	date.New(2017, 3, 9),
	//	date.New(2017, 4, 27),
	//	date.New(2017, 6, 8),
	//	date.New(2017, 7, 20),
	//	date.New(2017, 9, 7),
	//	date.New(2017, 10, 26),
	//	date.New(2017, 12, 14),
	//	date.New(2018, 1, 25),
	//	date.New(2018, 3, 8),
	//	date.New(2018, 4, 26),
	//	date.New(2018, 6, 14),
	//	date.New(2018, 7, 26),
	//	date.New(2018, 9, 13),
	//	date.New(2018, 10, 25),
	//	date.New(2018, 12, 13),
	//	date.New(2019, 1, 24),
	//	date.New(2019, 3, 7),
	//	date.New(2019, 4, 10),
	//	date.New(2019, 6, 6),
	//	date.New(2019, 7, 25),
	//	date.New(2019, 9, 12),
	//	date.New(2019, 10, 24),
	//	date.New(2019, 12, 12),
	//	date.New(2020, 1, 23),
	//	date.New(2020, 3, 12),
	//	date.New(2020, 4, 30),
	//	date.New(2020, 5, 27), // ECB press release
	//	date.New(2020, 6, 4),
	//	date.New(2020, 7, 16),
	//	date.New(2020, 9, 10),
	//	date.New(2020, 10, 29),
	//	date.New(2020, 12, 10),
	//}
}

// RiskLevelOK check if current risk level is low enough.
// Approves trading with return value 'true', 'false' otherwise.
//func (tr *Trader) riskLevelOK(currentTick tick.Tick, direction broker.BuyDirection, openPositions []broker.Position) (bool, string) {
//	const maxOpenPositions = 20
//
//	if len(openPositions)+1 > maxOpenPositions {
//		tr.clog.Debug("Too many open positions")
//		return false, "too many open positions"
//	}
//
//	var estTime = currentTick.Datetime.In(locEST)
//	if estTime.Hour() == 9 || estTime.Hour() == 16 { // 09:30 - 16:00 EST
//		// Avoid trading during US stocks markets (NYSE + NASDAQ) opening and closing
//		tr.clog.Debug("No trading during US stocks market opening and closing")
//		return false, "US stocks market is opening/closing"
//	}
//
//	// Avoid trading in Monday night
//	if currentTick.Datetime.Weekday() == time.Monday && currentTick.Datetime.Hour() < 8 {
//		return false, "no trading during Monday night"
//	}
//
//	if currentTick.Datetime.Hour() > 8 && currentTick.Datetime.Hour() < 10 {
//		return false, "Europe stocks market opening"
//	}
//
//	// Check if try to buyAfterRiskCheck against the price trend of the current day candle
//	perfPercentage := tr.today.PerformanceInPercentage()
//	tr.clog.WithFields(log.Fields{"TODAY_PERF_PERCENT": perfPercentage}).Debug("Today's performance")
//	switch direction {
//	case broker.BuyDirectionLong:
//		// Concur with daily trend
//		if perfPercentage.LessThan(decimal.NewFromFloat(minDayPerformancePercentageLong)) {
//			return false, fmt.Sprintf("long: daily perf < %.2f%%", minDayPerformancePercentageLong)
//		}
//
//	// Don't trade when daily performance is too high, likelihood of turnaround is too big
//	//dailyHighPerf, err := tr.dayTracker.PerformanceInPercentageQuantile(1.00 - maxDayPerformanceQuantile)
//	//if err != nil {
//	//	tr.clog.WithError(err).Debugf("tr.dayTracker.PerformanceInPercentageQuantile() failed")
//	//	return false, "long: cannot determine recent biggest high"
//	//}
//	//tr.clog.Debugf("dailyHighPerf is %.2f%%", dailyHighPerf)
//	//if perfPercentage.GreaterThanOrEqual(decimal.NewFromFloat(dailyHighPerf)) {
//	//	return false, "long: daily perf > recent biggest high"
//	//}
//	case broker.BuyDirectionShort:
//		// Concur with daily trend
//		if perfPercentage.GreaterThan(decimal.NewFromFloat(minDayPerformancePercentageShort)) {
//			return false, fmt.Sprintf("short: daily perf > %.2f%%", minDayPerformancePercentageShort)
//		}
//
//		// Don't trade when daily performance is too low, likelihood of turnaround is too big
//		//dailyLowPerf, err := tr.dayTracker.PerformanceInPercentageQuantile(maxDayPerformanceQuantile)
//		//if err != nil {
//		//	tr.clog.WithError(err).Debugf("tr.dayTracker.PerformanceInPercentageQuantile() failed")
//		//	return false, "short: cannot determine recent biggest low"
//		//}
//		//tr.clog.Debugf("dailyLowPerf is %.2f%%", dailyLowPerf)
//		//if dailyLowPerf < 0 && perfPercentage.LessThanOrEqual(decimal.NewFromFloat(dailyLowPerf)) {
//		//	return false, "short: daily perf < recent biggest low"
//		//}
//	}
//
//	for _, noTradingDay := range noTradingDays {
//		if date.FromTime(currentTick.Datetime).Equals(noTradingDay) {
//			return false, "no trading day"
//		}
//	}
//
//	//if tr.distanceSMAOK(currentTick, direction) != nil {
//	//	return false, "distance to SMA is too big"
//	//}
//
//	return true, ""
//}

var maxAllowedDistanceInPercent = decimal.NewFromFloat(0.5)

// flashCrashCheck - throw error if distance between to ticks is too big
func flashCrashCheck(previousTick, currentTick tick.Tick) error {
	distanceAsk := distanceInPercentage(previousTick.Ask, currentTick.Ask).Abs()
	distanceBid := distanceInPercentage(previousTick.Bid, currentTick.Bid).Abs()
	if distanceAsk.GreaterThan(maxAllowedDistanceInPercent) ||
		distanceBid.GreaterThan(maxAllowedDistanceInPercent) {
		return fmt.Errorf("distance between %v and %v is too bigger than %s%%",
			previousTick, currentTick, maxAllowedDistanceInPercent)
	}
	return nil
}

//func (tr *Trader) distanceSMAOK(currentTick tick.Tick, direction broker.BuyDirection) error {
//	gapToSMAPercent, err := tr.gapToSMAInPercent(currentTick)
//	if err != nil {
//		return err
//	}
//
//	switch direction {
//	case broker.BuyDirectionLong:
//		if gapToSMAPercent.GreaterThan(decimal.NewFromFloat(maxGapToSMAInPercent)) {
//			return fmt.Errorf("gap to SMA is too big: %s%%", gapToSMAPercent)
//		}
//	case broker.BuyDirectionShort:
//		if gapToSMAPercent.LessThan(decimal.NewFromFloat(maxGapToSMAInPercent).Neg()) {
//			return fmt.Errorf("gap to SMA is too big: %s%%", gapToSMAPercent)
//		}
//	}
//	return nil
//}
