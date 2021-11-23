package ohlc

import (
	"github.com/AMekss/assert"
	"github.com/shopspring/decimal"
	"sort"
	"testing"
	"time"
)

func TestOHLC_PerformanceInPercentage(t *testing.T) {
	var o = OHLC{
		Open:   decimal.NewFromFloat(100),
		Close:  decimal.NewFromFloat(90),
		closed: true,
	}

	perfFloat, _ := o.PerformanceInPercentage().Float64()
	assert.EqualFloat64(t, -10, perfFloat)

	o = OHLC{
		Open:   decimal.NewFromFloat(100),
		Close:  decimal.NewFromFloat(120),
		closed: true,
	}
	perfFloat, _ = o.PerformanceInPercentage().Float64()
	assert.EqualFloat64(t, 20, perfFloat)

	o = OHLC{
		Open:   decimal.NewFromFloat(50),
		Close:  decimal.NewFromFloat(100),
		closed: true,
	}
	perfFloat, _ = o.PerformanceInPercentage().Float64()
	assert.EqualFloat64(t, 100, perfFloat)
}

func TestOHLC_PerformanceFromOpenToHighAbsolute(t *testing.T) {
	var o = OHLC{
		Open:   decimal.NewFromFloat(100),
		High:   decimal.NewFromFloat(150),
		closed: true,
	}

	perfFloat, _ := o.PerformanceFromOpenToHighAbsolute().Float64()
	assert.EqualFloat64(t, 50, perfFloat)
}

func TestOHLC_PerformanceFromOpenToLowAbsolute(t *testing.T) {
	var o = OHLC{
		Open:   decimal.NewFromFloat(100),
		High:   decimal.NewFromFloat(50),
		closed: true,
	}

	perfFloat, _ := o.PerformanceFromOpenToHighAbsolute().Float64()
	assert.EqualFloat64(t, -50, perfFloat)
}

func TestOHLC_ReversionPerformanceFromHighAbsolute(t *testing.T) {
	var o = OHLC{
		High:   decimal.NewFromFloat(100),
		Close:  decimal.NewFromFloat(50),
		closed: true,
	}

	perfFloat, _ := o.ReversionPerformanceFromHighAbsolute().Float64()
	assert.EqualFloat64(t, -50, perfFloat)
}

func TestOHLC_ForceClose(t *testing.T) {
	var o = OHLC{}
	assert.False(t, o.closed)
	o.ForceClose()
	assert.True(t, o.closed)
}

func TestOHLC_HasGaps(t *testing.T) {
	var o = OHLC{}
	assert.False(t, o.HasGaps())
	o.Gaps = true
	assert.True(t, o.HasGaps())
}

func TestOHLC_Closed(t *testing.T) {
	var o = OHLC{}
	assert.False(t, o.HasGaps())
	o.closed = true
	assert.True(t, o.Closed())
}

func TestOHLC_String(t *testing.T) {
	var o = &OHLC{}
	assert.True(t, o.Validate() != nil)

	o = New("abc", time.Now(), time.Second, false)
	o.NewPrice(decimal.NewFromFloat(1), time.Now())
	assert.True(t, o.String() != "")
}

func TestOHLC_Age(t *testing.T) {
	var now = time.Now()
	var o = OHLC{Start: now}
	assert.EqualFloat64(t, 1, o.Age(now.Add(time.Second)).Seconds())
}

func TestOHLC_Validate(t *testing.T) {
	var o = &OHLC{}
	assert.True(t, o.Validate() != nil)

	o = New("abc", time.Now(), time.Second, false)
	price := decimal.NewFromFloat(1)
	o.NewPrice(price, time.Now())
	assert.NoError(t, o.Validate())

	// open = 0
	obroken := o
	obroken.Open = decimal.Decimal{}
	assert.True(t, obroken.Validate() != nil)

	// low > high
	obroken = o
	obroken.Low = obroken.High.Add(price)
	assert.True(t, obroken.Validate() != nil)

	// close < low
	obroken = o
	obroken.Close = obroken.Low.Sub(price)
	assert.True(t, obroken.Validate() != nil)

	// close > high
	obroken = o
	obroken.Close = obroken.High.Add(price)
	assert.True(t, obroken.Validate() != nil)

	// end < start
	obroken = o
	obroken.End = obroken.Start.Add(-time.Minute)
	assert.True(t, obroken.Validate() != nil)

	// instrument == ""
	obroken = o
	obroken.Instrument = ""
	assert.True(t, obroken.Validate() != nil)
}

func TestOHLC_VolatilityInPercentage(t *testing.T) {
	var o = New("abc", time.Now(), time.Second, false)
	o.NewPrice(decimal.NewFromFloat(1), time.Now())
	o.NewPrice(decimal.NewFromFloat(2), time.Now())

	vola := o.VolatilityInPercentage()
	volaFloat, _ := vola.Float64()
	assert.EqualFloat64(t, 100, volaFloat)
}

func TestOHLC_NewPrice(t *testing.T) {
	var o = New("abc", time.Now(), time.Second, false)
	now := time.Now()

	// open
	price := decimal.NewFromFloat(1)
	o.NewPrice(price, now)
	assertDecimal(t, price, o.Open)
	assert.True(t, o.priceDataSeen)

	// high
	price = decimal.NewFromFloat(2)
	o.NewPrice(price, now)
	assertDecimal(t, price, o.High)

	// low
	price = decimal.NewFromFloat(0.5)
	o.NewPrice(price, now)
	assertDecimal(t, price, o.Low)
	assertDecimal(t, price, o.Close)

	// close
	now = o.End
	price = decimal.NewFromFloat(1.2)
	considered := o.NewPrice(price, now)
	assert.True(t, considered)
	assertDecimal(t, price, o.Close)
	assert.True(t, o.closed)
	assert.True(t, o.Closed())
	assert.EqualTime(t, now, o.End)

	// after end
	now = o.End.Add(time.Second)
	price = decimal.NewFromFloat(1.3)
	considered = o.NewPrice(price, now)
	assert.False(t, considered)
}

func TestOHLC_NewPrice_with_Gaps(t *testing.T) {
	var o = New("abc", time.Now(), time.Hour, false)
	now := time.Now()

	price := decimal.NewFromFloat(1)
	o.NewPrice(price, now)
	assert.False(t, o.HasGaps())

	o.NewPrice(price, now.Add(maxGapBetweenTicksInSeconds).Add(time.Minute))
	assert.True(t, o.HasGaps())
}

func assertDecimal(t *testing.T, want, got decimal.Decimal) {
	wantFloat, _ := want.Float64()
	gotFloat, _ := got.Float64()
	assert.EqualFloat64(t, wantFloat, gotFloat)
}

func Test__smoothCandleStart(t *testing.T) {
	period := time.Minute * 15
	want := time.Date(2020, 12, 17, 21, 15, 0, 0, time.UTC)
	is := time.Date(2020, 12, 17, 21, 24, 7, 8, time.UTC)
	assert.EqualTime(t, want, smoothCandleStart(is, period))

	period = time.Minute * 15
	want = time.Date(2020, 12, 17, 21, 30, 0, 0, time.UTC)
	is = time.Date(2020, 12, 17, 21, 33, 7, 8, time.UTC)
	assert.EqualTime(t, want, smoothCandleStart(is, period))

	period = time.Minute * 45
	want = time.Date(2020, 12, 17, 21, 0, 0, 0, time.UTC)
	is = time.Date(2020, 12, 17, 21, 15, 7, 8, time.UTC)
	assert.EqualTime(t, want, smoothCandleStart(is, period))

	period = time.Minute * 60
	want = time.Date(2020, 12, 17, 21, 0, 0, 0, time.UTC)
	is = time.Date(2020, 12, 17, 21, 15, 7, 8, time.UTC)
	assert.EqualTime(t, want, smoothCandleStart(is, period))
}

func Test__hightime(t *testing.T) {
	var one = decimal.NewFromFloat(1)
	var o = New("abc", time.Now(), time.Hour, false)
	now := time.Now()

	// Price: 1
	price := one
	o.NewPrice(price, now)

	// Price: 2 -> our high
	price = price.Add(one)
	now = now.Add(time.Minute)
	highTime := now
	o.NewPrice(price, now)

	// Price: 1
	price = price.Sub(one)
	now = now.Add(time.Minute)
	o.NewPrice(price, now)

	assert.EqualTime(t, highTime, o.HighTime)
}

func Test__lowtime(t *testing.T) {
	var one = decimal.NewFromFloat(1)
	var o = New("abc", time.Now(), time.Hour, false)
	now := time.Now()

	// Price: 1
	price := one
	o.NewPrice(price, now)

	// Price: 0 -> our low
	price = decimal.Zero
	now = now.Add(time.Minute)
	lowTime := now
	o.NewPrice(price, now)

	// Price: 1
	price = one
	now = now.Add(time.Minute)
	o.NewPrice(price, now)

	assert.EqualTime(t, lowTime, o.LowTime)
}

func Test__ToTicks(t *testing.T) {
	var one = decimal.NewFromFloat(1)
	now := time.Now()
	var o = New("abc", now, time.Hour, false)

	// open
	price := one
	openTime := now
	o.NewPrice(price, now)

	// low
	price = decimal.Zero
	now = now.Add(time.Minute)
	lowTime := now
	o.NewPrice(price, now)

	// high
	price = decimal.NewFromFloat(5)
	now = now.Add(time.Minute)
	highTime := now
	o.NewPrice(price, now)

	price = decimal.NewFromFloat(3)
	now = now.Add(time.Minute)
	closeTime := now
	o.NewPrice(price, now)
	o.ForceClose()

	assert.EqualTime(t, lowTime, o.LowTime)
	assert.EqualTime(t, highTime, o.HighTime)
	assert.EqualTime(t, openTime, o.Start)
	assert.EqualTime(t, closeTime, o.End)

	ticks := o.ToTicks()
	assert.EqualInt(t.Fatalf, 4, len(ticks))
	assert.EqualTime(t, ticks[0].Datetime, openTime)
	assert.EqualTime(t, ticks[1].Datetime, lowTime)
	assert.EqualTime(t, ticks[2].Datetime, highTime)
	assert.EqualTime(t, ticks[3].Datetime, closeTime)
}

func TestOHLC__Sort(t *testing.T) {
	var now = time.Now()
	var o1 = generateOHLC(now, 1)
	var o2 = generateOHLC(now.Add(time.Minute), 2)
	var ohlcList = []OHLC{*o2, *o1}
	sort.Sort(OHLCList(ohlcList))

	assert.EqualTime(t, ohlcList[0].End, o1.End)
	assert.EqualTime(t, ohlcList[1].End, o2.End)
}

func generateOHLC(when time.Time, price float64) *OHLC {
	var o = New("abc", when, time.Hour, false)
	priceDec := decimal.NewFromFloat(price)
	o.NewPrice(priceDec, when)
	o.ForceClose()
	return o
}
