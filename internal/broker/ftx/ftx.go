package ftx

import (
	"context"
	"fmt"
	"github.com/go-numb/go-ftx/realtime"
	"github.com/go-numb/go-ftx/rest"
	"github.com/shopspring/decimal"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/paperwallet"
	"github.com/sklinkert/at/pkg/tick"
)

type FTX struct {
	paperwallet *paperwallet.Paperwallet
	client      *rest.Client
	instrument  string
}

func New(instrument string, paperwallet *paperwallet.Paperwallet) *FTX {
	return &FTX{
		paperwallet: paperwallet,
		instrument:  instrument,
	}
}

func (f *FTX) Buy(order broker.Order) (broker.Position, error) {
	return f.paperwallet.Buy(order)
}

func (f *FTX) Sell(position broker.Position) error {
	return f.paperwallet.Sell(position)
}

func (f *FTX) GetOpenPosition(positionRef string) (position broker.Position, err error) {
	return f.paperwallet.GetOpenPosition(positionRef)
}

func (f *FTX) GetOpenPositions() ([]broker.Position, error) {
	return f.paperwallet.GetOpenPositions()
}

func (f *FTX) GetOpenPositionsByInstrument(instrument string) ([]broker.Position, error) {
	return f.paperwallet.GetOpenPositionsByInstrument(instrument)
}

func (f *FTX) GetClosedPositions() ([]broker.Position, error) {
	return f.paperwallet.GetClosedPositions()
}

func (f *FTX) ListenToPriceFeed(tickChan chan tick.Tick) {
	defer close(tickChan)
	ctx := context.Background()

	ch := make(chan realtime.Response)
	symbols := []string{f.instrument}
	realtime.Connect(ctx, ch, []string{"ticker"}, symbols, nil)

	for {
		select {
		case v := <-ch:
			switch v.Type {
			case realtime.TICKER:
				fmt.Printf("%s	%+v\n", v.Symbol, v.Ticker)
				bid := decimal.NewFromFloat(v.Ticker.Bid)
				ask := decimal.NewFromFloat(v.Ticker.Ask)
				ticker := tick.New(v.Symbol, v.Ticker.Time.Time, bid, ask)
				tickChan <- ticker
			}
		}
	}
}
