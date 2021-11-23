package tick

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type Tick struct {
	// ID is for keeping the order from reading CSV file in DB because the timestamp is not precise enough.
	// We have to deal with multiple ticks at the same time.
	ID         uint            `gorm:"primaryKey"`
	Datetime   time.Time       `gorm:"index"`
	Instrument string          `gorm:"index"`
	Bid        decimal.Decimal `gorm:"type:decimal(13,6);"`
	Ask        decimal.Decimal `gorm:"type:decimal(13,6);"`
	price      decimal.Decimal `gorm:"-"`
}

//var maxSpread = decimal.NewFromFloat(0.00025) // 2.5 pips
var dec2 = decimal.NewFromFloat(2)

func New(instrument string, datetime time.Time, bid, ask decimal.Decimal) Tick {
	return Tick{
		0,
		datetime,
		instrument,
		bid,
		ask,
		bid.Add(ask).Div(dec2),
	}
}

func (t *Tick) Spread() decimal.Decimal {
	return t.Ask.Sub(t.Bid).Abs()
}

func (t *Tick) SpreadInPercent() decimal.Decimal {
	var n = t.Ask.Sub(t.Bid).Div(t.Bid)
	return n.Mul(decimal.NewFromFloat(100)).Abs()
}

func (t *Tick) String() string {
	return fmt.Sprintf("{Datetime=%s Bid=%s Ask=%s}",
		t.Datetime.String(), t.Bid.String(), t.Ask.String())
}

func (t *Tick) Price() decimal.Decimal {
	return t.price
}

func (t *Tick) Validate() error {
	if t.Datetime.IsZero() {
		return errors.New("empty datetime")
	}
	if t.Bid.IsZero() {
		return errors.New("empty bid")
	}
	if t.Ask.IsZero() {
		return errors.New("empty ask")
	}
	if t.Ask.LessThan(t.Bid) {
		return errors.New("ask is less than bid")
	}
	//if t.Spread().GreaterThan(maxSpread) {
	//	return errors.New("spread is too big")
	//}
	return nil
}
