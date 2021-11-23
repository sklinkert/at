package amcharts

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/ohlc"
	"html/template"
	"io"
	"net/http"
	"sort"
	"time"
)

type dataPoint struct {
	Start     string            `json:"date"`
	End       string            `json:"-"`
	Open      string            `json:"open"`
	High      string            `json:"high"`
	Low       string            `json:"low"`
	Close     string            `json:"close"`
	Volume    int               `json:"volume"`
	Positions []broker.Position `json:"-"`
}

//go:embed assets
var staticFiles embed.FS

type Chart struct {
	port       int
	candles    []ohlc.OHLC
	positions  []broker.Position
	instrument string
}

func (c *Chart) OnPosition(position broker.Position) {
	c.positions = append(c.positions, position)
}

func (c *Chart) OnCandle(candle ohlc.OHLC) {
	if candle.Closed() {
		c.candles = append(c.candles, candle)
	}
}

func dec2Float(d decimal.Decimal) float64 {
	fl, _ := d.Float64()
	return fl
}

type positionList []broker.Position

func (p positionList) Len() int {
	return len(p)
}

func (p positionList) Less(i, j int) bool {
	if p[i].BuyTime.Equal(p[j].BuyTime) {
		return p[i].SellTime.Before(p[j].SellTime)
	}
	return p[i].BuyTime.Before(p[j].BuyTime)
}

func (p positionList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (c *Chart) RenderChartToHTML() (string, error) {
	var buf = new(bytes.Buffer)
	if err := c.RenderChart(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *Chart) RenderChart(w io.Writer) error {
	var dataPoints []dataPoint

	t, err := template.ParseFS(staticFiles, "assets/amcharts.html")
	if err != nil {
		return err
	}

	locBerlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return err
	}

	for _, candle := range c.candles {
		var openPositions positionList
		var volume = 0

		for _, position := range c.positions {
			position.BuyTime = position.BuyTime.In(locBerlin)
			position.SellTime = position.SellTime.In(locBerlin)

			// Sum up all open closedPositions
			if position.BuyTime.Equal(candle.Start) || (position.BuyTime.After(candle.Start) && position.BuyTime.Before(candle.End)) {
				volume++
				openPositions = append(openPositions, position)
			}
		}
		sort.Sort(openPositions)

		dataPoints = append(dataPoints, dataPoint{
			Start:     candle.Start.In(locBerlin).Format("2006-01-02 15:04"),
			End:       candle.End.In(locBerlin).Format("2006-01-02 15:04"),
			Open:      fmt.Sprintf("%.5f", dec2Float(candle.Open)),
			High:      fmt.Sprintf("%.5f", dec2Float(candle.High)),
			Low:       fmt.Sprintf("%.5f", dec2Float(candle.Low)),
			Close:     fmt.Sprintf("%.5f", dec2Float(candle.Close)),
			Volume:    volume,
			Positions: openPositions,
		})
	}

	dataPointsJSON, err := json.Marshal(&dataPoints)
	if err != nil {
		return err
	}

	chartData := struct {
		DataPoints     []dataPoint
		DataPointsJSON string
	}{
		dataPoints,
		string(dataPointsJSON[:]),
	}

	return t.Execute(w, chartData)
}

func (c *Chart) RenderEquityCurve(w io.Writer) error {
	t, err := template.ParseFS(staticFiles, "assets/amcharts-equity.html")
	if err != nil {
		return err
	}

	type EquityCurvePoint struct {
		Date         time.Time
		TotalBalance float64
	}
	var equityCurvePoints []EquityCurvePoint
	var totalBalance float64
	for _, position := range c.positions {
		totalBalance += position.PerformanceAbsolute(decimal.Zero, decimal.Zero)
		equityCurvePoints = append(equityCurvePoints, EquityCurvePoint{
			Date:         position.BuyTime,
			TotalBalance: totalBalance,
		})
	}

	dataPointsJSON, err := json.Marshal(&equityCurvePoints)
	if err != nil {
		return err
	}

	chartData := struct {
		DataPointsJSON string
		Instrument     string
	}{
		string(dataPointsJSON[:]),
		c.instrument,
	}
	return t.Execute(w, chartData)
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
	http.HandleFunc("/equity", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		if err := c.RenderEquityCurve(w); err != nil {
			log.WithError(err).Error("RenderChart() failed")
		}
	})
	fmt.Printf("Chart available at http://localhost:%d/chart\n", c.port)
	fmt.Printf("Equity curve available at http://localhost:%d/equity\n", c.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}

type Option func(*Chart)

func WithPort(port int) Option {
	return func(chart *Chart) {
		chart.port = port
	}
}

func NewChart(instrument string, options ...Option) *Chart {
	chart := &Chart{
		port:       8080,
		instrument: instrument,
	}
	for _, option := range options {
		option(chart)
	}
	return chart
}
