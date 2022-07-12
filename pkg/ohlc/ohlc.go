package ohlc

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/tick"
	"gorm.io/gorm"
	"time"
)

const maxGapBetweenTicksInSeconds = 60

// OHLC represents a full candle
type OHLC struct {
	Instrument        string
	Open              decimal.Decimal `gorm:"type:decimal(13,6);"`
	High              decimal.Decimal `gorm:"type:decimal(13,6);"`
	HighTime          time.Time
	Low               decimal.Decimal `gorm:"type:decimal(13,6);"`
	LowTime           time.Time
	Close             decimal.Decimal `gorm:"type:decimal(13,6);"`
	Start             time.Time       `gorm:"index"`
	End               time.Time
	Duration          time.Duration `gorm:"index"`
	Gaps              bool
	priceDataSeen     bool
	closed            bool
	lastReceivedPrice time.Time
}

func New(instrument string, now time.Time, duration time.Duration, round bool) *OHLC {
	var start = now
	if round {
		start = smoothCandleStart(start, duration)
	}
	return &OHLC{
		Instrument: instrument,
		Start:      start,
		End:        start.Add(duration),
		Duration:   duration,
	}
}

func (o *OHLC) ForceClose() {
	o.closed = true
	o.End = o.lastReceivedPrice
}

func (o *OHLC) String() string {
	return fmt.Sprintf("OHLC(%s, Open=%s High=%s Low=%s Close=%s, Start=%s End=%s)",
		o.Instrument, o.Open, o.High, o.Low, o.Close, o.Start, o.End)
}

// Closed return true if candle is already closed (by time) or false if not.
func (o *OHLC) Closed() bool {
	return o.closed
}

// NewPrice handles new price data
// Returns true if data was considered
// Returns false if the candle is already closed.
func (o *OHLC) NewPrice(price decimal.Decimal, now time.Time) bool {
	if o.Closed() {
		return false
	}

	if now.After(o.End) || now.Equal(o.End) {
		o.closed = true
		return false
	}

	if price.GreaterThan(o.High) {
		o.High = price
		o.HighTime = now
	} else if price.LessThan(o.Low) {
		o.Low = price
		o.LowTime = now
	}

	if !o.priceDataSeen {
		o.Open = price
		o.Low = price
		o.LowTime = now
	}

	if o.priceDataSeen {
		diffLastReceivedValidPrice := now.Sub(o.lastReceivedPrice)
		if diffLastReceivedValidPrice.Seconds() > maxGapBetweenTicksInSeconds {
			o.Gaps = true
			//fmt.Printf("Gap to last received valid tick is %.2f seconds at %s\n",
			//	diffLastReceivedValidPrice.Seconds(), o.lastReceivedPrice)
		}
	}

	o.lastReceivedPrice = now
	o.Close = price
	o.priceDataSeen = true

	return true
}

func (o *OHLC) HasGaps() bool {
	return o.Gaps
}

func (o *OHLC) HasPriceData() bool {
	return o.priceDataSeen
}

func (o *OHLC) Validate() error {
	if !o.priceDataSeen {
		return errors.New("no data received")
	}
	if o.Low.GreaterThan(o.High) {
		return errors.New("low is higher than High")
	}
	if o.Open.GreaterThan(o.High) {
		return errors.New("open is higher than High")
	}
	if o.Open.LessThan(o.Low) {
		return errors.New("open is lower than Low")
	}
	if o.Close.GreaterThan(o.High) {
		return errors.New("close is higher than High")
	}
	if o.Close.LessThan(o.Low) {
		return errors.New("close is lower than Low")
	}
	if o.End.Before(o.Start) {
		return errors.New("end is before start")
	}
	if o.Instrument == "" {
		return errors.New("instrument name is missing")
	}
	return nil
}

func (o *OHLC) PerformanceFromOpenToHighAbsolute() decimal.Decimal {
	if o.Open.IsZero() {
		log.Panicf("ohlc.Open is zero: %+v", o)
	}
	return o.High.Sub(o.Open).Div(o.Open).Mul(decimal.NewFromFloat(100)).Round(4)
}

func (o *OHLC) PerformanceFromOpenToLowAbsolute() decimal.Decimal {
	if o.Open.IsZero() {
		log.Panicf("ohlc.Open is zero: %+v", o)
	}
	return o.Low.Sub(o.Open).Div(o.Open).Mul(decimal.NewFromFloat(100)).Round(4)
}

func (o *OHLC) ReversionPerformanceFromHighAbsolute() decimal.Decimal {
	if o.High.IsZero() {
		log.Panicf("ohlc.High is zero: %+v", o)
	}
	return o.Close.Sub(o.High).Div(o.High).Mul(decimal.NewFromFloat(100)).Round(4)
}

func (o *OHLC) PerformanceInPercentage() decimal.Decimal {
	if o.Open.IsZero() {
		log.Panicf("ohlc.Open is zero: %+v", o)
	}
	return o.Close.Sub(o.Open).Div(o.Open).Mul(decimal.NewFromFloat(100)).Round(4)
}

func (o *OHLC) VolatilityInPercentage() decimal.Decimal {
	if o.Open.IsZero() {
		log.Panicf("ohlc.Open is zero: %+v", o)
	}
	return o.High.Sub(o.Low).Div(o.Open).Mul(decimal.NewFromFloat(100)).Round(4)
}

func (o *OHLC) Age(now time.Time) time.Duration {
	return now.Sub(o.Start)
}

func (o *OHLC) Store(gormDB *gorm.DB) error {
	var oc = *o
	oc.Start = oc.Start.In(time.UTC)
	oc.End = oc.End.In(time.UTC)
	if err := gormDB.Create(oc).Error; err != nil {
		return err
	}
	return nil
}

func (o *OHLC) OpenTick() tick.Tick {
	return tick.New(o.Instrument, o.Start, o.Open, o.Open)
}

func (o *OHLC) CloseTick() tick.Tick {
	return tick.New(o.Instrument, o.End, o.Close, o.Close)
}

func (o *OHLC) HighTick() tick.Tick {
	return tick.New(o.Instrument, o.HighTime, o.High, o.High)
}

func (o *OHLC) LowTick() tick.Tick {
	return tick.New(o.Instrument, o.LowTime, o.Low, o.Low)
}

// ToTicks converts the OHLC candle to 4 ticks. It ensures the correct
// chronological order of high and low.
func (o *OHLC) ToTicks() []tick.Tick {
	var ticks []tick.Tick
	var high = o.HighTick()
	var low = o.LowTick()

	ticks = append(ticks, o.OpenTick())
	if high.Datetime.After(low.Datetime) {
		ticks = append(ticks, low)
		ticks = append(ticks, high)
	} else {
		ticks = append(ticks, high)
		ticks = append(ticks, low)
	}
	ticks = append(ticks, o.CloseTick())

	return ticks
}

// round ts to the closest period
func smoothCandleStart(ts time.Time, period time.Duration) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute()/int(period.Minutes())*int(period.Minutes()), 0, 0, ts.Location())
}

// ToHeikinAshi calculates a Heikin Ashi candle from two OHLC candles
func ToHeikinAshi(previous, now *OHLC) *OHLC {
	var ha OHLC
	if err := copier.Copy(&ha, now); err != nil {
		log.WithError(err).Fatal("copier.Copy failed")
	}
	ha.Open = decimal.Avg(previous.Open, previous.Close)
	ha.Close = decimal.Avg(now.Open, now.Close, now.High, now.Low)
	ha.High = decimal.Max(now.High, ha.Open, ha.Close)
	ha.Low = decimal.Min(now.Low, ha.Open, ha.Close)
	return &ha
}

type OHLCList []OHLC

func (e OHLCList) Len() int {
	return len(e)
}

func (e OHLCList) Less(i, j int) bool {
	return e[i].End.Before(e[j].End)
}

func (e OHLCList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
