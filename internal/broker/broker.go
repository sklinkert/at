package broker

import (
	"errors"
	"github.com/sklinkert/at/pkg/tick"
)

type BuyDirection int

const (
	BuyDirectionLong BuyDirection = iota
	BuyDirectionShort
)

var (
	ErrPositionNotFound    = errors.New("position not found")
	ErrUnknownBuyDirection = errors.New("unknown buy direction")
)

func (bd BuyDirection) String() string {
	return [...]string{"Long", "Short"}[bd]
}

type Broker interface {
	Buy(order Order) (Position, error)
	Sell(position Position) error
	GetOpenPosition(positionRef string) (position Position, err error)
	GetOpenPositions() ([]Position, error)
	GetOpenPositionsByInstrument(instrument string) ([]Position, error)
	GetClosedPositions() ([]Position, error)
	ListenToPriceFeed(chan tick.Tick)
}
