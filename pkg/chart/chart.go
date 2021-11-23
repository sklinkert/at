package chart

import (
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/ohlc"
	"io"
)

type Chart interface {
	OnPosition(position broker.Position)
	OnCandle(candle ohlc.OHLC)
	RenderChart(w io.Writer) error
	RenderChartToHTML() (string, error)
	Start() error
}
