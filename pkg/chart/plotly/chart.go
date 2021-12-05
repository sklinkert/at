package plotly

import (
	"bytes"
	"embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/ohlc"
	"html/template"
	"io"
	"net/http"
)

//go:embed assets
var staticFiles embed.FS

type Chart struct {
	port      int
	candles   []ohlc.OHLC
	positions []broker.Position
}

func (c *Chart) OnPosition(position broker.Position) {
	c.positions = append(c.positions, position)
}

func (c *Chart) OnCandle(candle ohlc.OHLC) {
	if candle.Closed() {
		c.candles = append(c.candles, candle)
	}
}

func (c *Chart) RenderChartToHTML() (string, error) {
	var buf = new(bytes.Buffer)
	if err := c.RenderChart(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *Chart) RenderChart(w io.Writer) error {
	t, err := template.ParseFS(staticFiles, "assets/chart.html")
	if err != nil {
		return err
	}

	var orders []broker.Order
	for _, position := range c.positions {
		orders = append(orders, broker.Order{
			Direction:   position.BuyDirection,
			Size:        position.Size,
			Instrument:  position.Instrument,
			CandleStart: position.CandleBuyTime,
		})
		orders = append(orders, broker.Order{
			Direction:   position.BuyDirection,
			Size:        -position.Size,
			Instrument:  position.Instrument,
			CandleStart: position.CandleSellTime,
		})
	}
	return t.Execute(w, struct {
		Candles []ohlc.OHLC
		Orders  []broker.Order
	}{
		Candles: c.candles,
		Orders:  orders,
	})
}

func (c *Chart) Start() error {
	http.Handle(
		"/assets/",
		http.FileServer(http.FS(staticFiles)),
	)

	http.HandleFunc("/chart", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		if err := c.RenderChart(w); err != nil {
			log.WithError(err).Error("RenderChart() failed")
		}
	})
	fmt.Printf("Chart available at http://localhost:%d/chart\n", c.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}

type Option func(*Chart)

func WithPort(port int) Option {
	return func(chart *Chart) {
		chart.port = port
	}
}

func NewChart(options ...Option) *Chart {
	chart := &Chart{
		port: 8080,
	}
	for _, option := range options {
		option(chart)
	}
	return chart
}
