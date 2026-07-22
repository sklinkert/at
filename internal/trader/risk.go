package trader

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/pkg/tick"
)

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
