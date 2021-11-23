package backtest

import (
	"github.com/sklinkert/at/internal/broker"
)

// GetClosedPositions returns all closed positions
func (b *Backtest) GetClosedPositions() ([]broker.Position, error) {
	b.RLock()
	defer b.RUnlock()

	return b.paperwallet.GetClosedPositions()
}

// GetOpenPositionsByInstrument returns all open positions for given instrument name
func (b *Backtest) GetOpenPositionsByInstrument(instrument string) ([]broker.Position, error) {
	return b.paperwallet.GetOpenPositionsByInstrument(instrument)
}

// GetOpenPositions returns all open positions
func (b *Backtest) GetOpenPositions() ([]broker.Position, error) {
	b.RLock()
	defer b.RUnlock()

	return b.paperwallet.GetOpenPositions()
}

// GetOpenPosition returns the position for the given reference
func (b *Backtest) GetOpenPosition(positionRef string) (broker.Position, error) {
	b.RLock()
	defer b.RUnlock()

	return b.paperwallet.GetOpenPosition(positionRef)
}
