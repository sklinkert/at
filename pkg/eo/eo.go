package eo

import (
	ring "github.com/falzm/golang-ring"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/helper"
	"github.com/sklinkert/at/pkg/ohlc"
	"sort"
)

// Environment Overlays
// Adapt strategy setup towards market current momentum

type EnvironmentOverlay struct {
	candles             ring.Ring
	priceChangesPercent ring.Ring
}

type RiskLevel int

const (
	minCandles = 50

	RLow RiskLevel = iota
	RModerate
	RHigh
	RExtreme

	DefaultRisk = RModerate
)

func New() *EnvironmentOverlay {
	candles := ring.Ring{}
	candles.SetCapacity(minCandles)

	priceChangesPercent := ring.Ring{}
	priceChangesPercent.SetCapacity(minCandles)

	return &EnvironmentOverlay{
		candles:             candles,
		priceChangesPercent: priceChangesPercent,
	}
}

func (eo *EnvironmentOverlay) AddCandle(candle *ohlc.OHLC) {
	closePrice, _ := candle.Close.Float64()
	eo.candles.Enqueue(closePrice)

	perfPercent, _ := candle.PerformanceInPercentage().Float64()
	eo.priceChangesPercent.Enqueue(perfPercent)
}

func (eo *EnvironmentOverlay) riskLevel() RiskLevel {
	var prices = eo.candles.Values()
	var priceChangesPercent = eo.priceChangesPercent.Values()
	if len(prices) < minCandles || len(priceChangesPercent) < minCandles {
		return DefaultRisk
	}

	var lastPriceChange = priceChangesPercent[len(prices)-1]
	var sortedPriceChanges = priceChangesPercent
	sort.Float64s(sortedPriceChanges)

	var (
		priceChangeQ1 = helper.GetPercentile(sortedPriceChanges, 25)
		priceChangeQ2 = helper.GetPercentile(sortedPriceChanges, 50)
		priceChangeQ3 = helper.GetPercentile(sortedPriceChanges, 75)
	)

	if lastPriceChange < priceChangeQ1 {
		return RLow
	} else if lastPriceChange < priceChangeQ2 {
		return RModerate
	} else if lastPriceChange < priceChangeQ3 {
		return RHigh
	} else {
		return RExtreme
	}
}

func (eo *EnvironmentOverlay) RSI() (upperThreshold, lowerThreshold float64) {
	var riskLevel = eo.riskLevel()

	switch riskLevel {
	case RLow:
		return 80, 20
	case RModerate:
		return 85, 15
	case RHigh:
		return 90, 10
	case RExtreme:
		return 95, 5
	default:
		log.Panicf("Unsupported risk level %d", int(riskLevel))
		return 100, 0
	}
}
