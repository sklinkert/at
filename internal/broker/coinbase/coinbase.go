package coinbase

import (
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/paperwallet"
)

type Coinbase struct {
	paperwallet *paperwallet.Paperwallet
	cbClient    *coinbasepro.Client
	instruments []string
}

func New(instrument string, paperwallet *paperwallet.Paperwallet) *Coinbase {
	return &Coinbase{
		paperwallet: paperwallet,
		instruments: []string{instrument}, // we support only 1 for now
		cbClient:    coinbasepro.NewClient(),
	}
}

func (cb *Coinbase) Buy(order broker.Order) (broker.Position, error) {
	return cb.paperwallet.Buy(order)
}

func (cb *Coinbase) Sell(position broker.Position) error {
	return cb.paperwallet.Sell(position)
}

func (cb *Coinbase) GetOpenPosition(positionRef string) (position broker.Position, err error) {
	return cb.paperwallet.GetOpenPosition(positionRef)
}

func (cb *Coinbase) GetOpenPositions() ([]broker.Position, error) {
	return cb.paperwallet.GetOpenPositions()
}

func (cb *Coinbase) GetOpenPositionsByInstrument(instrument string) ([]broker.Position, error) {
	return cb.paperwallet.GetOpenPositionsByInstrument(instrument)
}

func (cb *Coinbase) GetClosedPositions() ([]broker.Position, error) {
	return cb.paperwallet.GetClosedPositions()
}
