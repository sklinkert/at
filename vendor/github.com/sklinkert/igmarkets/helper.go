package igmarkets

import (
	"fmt"
	"time"
)

func timeZoneOffset2Location(timeZoneOffset int) *time.Location {
	var name = fmt.Sprintf("UTC+%d", timeZoneOffset)
	if timeZoneOffset < 0 {
		name = fmt.Sprintf("UTC-%d", timeZoneOffset)
	}
	return time.FixedZone(name, timeZoneOffset)
}
