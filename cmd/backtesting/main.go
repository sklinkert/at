package main

import (
	"context"
	"github.com/lfritz/env"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker/backtest"
	"github.com/sklinkert/at/internal/paperwallet"
	"github.com/sklinkert/at/internal/strategy"
	heikinashi "github.com/sklinkert/at/internal/strategy/HeikinAshi"
	"github.com/sklinkert/at/internal/strategy/doji"
	"github.com/sklinkert/at/internal/strategy/engulfing"
	"github.com/sklinkert/at/internal/strategy/harami"
	"github.com/sklinkert/at/internal/strategy/lowcandle"
	"github.com/sklinkert/at/internal/strategy/rsi"
	"github.com/sklinkert/at/internal/strategy/rsiadx"
	"github.com/sklinkert/at/internal/strategy/scalper"
	"github.com/sklinkert/at/internal/strategy/sma10"
	"github.com/sklinkert/at/internal/strategy/stochrsi"
	"github.com/sklinkert/at/internal/trader"
	chart "github.com/sklinkert/at/pkg/chart"
	"github.com/sklinkert/at/pkg/chart/amcharts"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var conf struct {
	importHistDataCSVFiles []string
	broker                 string
	gatherPerformanceData  bool
	instrument             string
	persistData            bool
	debug                  bool
	dbHost                 string
	dbUser                 string
	dbPassword             string
	dbName                 string
	dbPort                 int
	strategyName           string
	priceDBFile            string
	priceSource            string
	yearFrom               int
	yearTo                 int
	candleDuration         string
	igAPIURL               string
	igIdentifier           string
	igAPIKey               string
	igPassword             string
	igAccountID            string
}

func main() {
	const BrokerBacktest = "backtest"
	var graph chart.Chart
	var ctx = context.Background()

	var err error
	var e = env.New()
	e.Flag("DEBUG", &conf.debug, "Enable debug logging")
	e.Flag("PERFORMANCE_DATA", &conf.gatherPerformanceData, "Gather performance data and print as CSV")
	e.OptionalList("IMPORT_HISTDATA_CSV_FILES", &conf.importHistDataCSVFiles, ",", []string{}, "Import CSV files from histdata.com")
	e.OptionalString("PRICE_SOURCE", &conf.priceSource, "LOCAL_DB", "Price source for backtesting. E.g. 'PATTERN_TRADING'")
	e.OptionalString("PRICE_DB_FILE", &conf.priceDBFile, "/data/EURUSD/db", "SQLite DB file for price data (OHLCs)")
	e.OptionalString("INSTRUMENT", &conf.instrument, "CS.D.EURUSD.MINI.IP", "instrument to trade")
	e.OptionalString("BROKER", &conf.broker, BrokerBacktest, "Broker backend")
	e.OptionalString("STRATEGY", &conf.strategyName, "meanreversion", "strategy to be executed")
	e.OptionalString("CANDLE_DURATION", &conf.candleDuration, "60m", "Duration for OHLC candle")
	e.OptionalInt("YEAR_FROM", &conf.yearFrom, 1970, "Backtesting beginning")
	e.OptionalInt("YEAR_TO", &conf.yearTo, 2022, "Backtesting end")

	if err := e.Load(); err != nil {
		log.WithError(err).Fatal("env loading failed")
	}
	if conf.debug {
		log.SetLevel(log.DebugLevel)
	}

	db, err := gorm.Open(sqlite.Open("backtesting.db"), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("failed to open database file")
	}
	candleDuration, err := time.ParseDuration(conf.candleDuration)
	if err != nil {
		log.WithError(err).Fatal("cannot parse candle duration")
	}

	var strategyBackend strategy.Strategy
	switch conf.strategyName {
	case strategy.NameDOJI:
		strategyBackend = doji.New(conf.instrument)
	case strategy.NameHeikinAshi:
		strategyBackend = heikinashi.New(conf.instrument)
	case strategy.NameScalper:
		strategyBackend = scalper.New(conf.instrument)
	case strategy.NameStochRSI:
		strategyBackend = stochrsi.New(conf.instrument)
	case strategy.NameLowCandle:
		strategyBackend = lowcandle.New(conf.instrument, candleDuration)
	case strategy.NameHarami:
		strategyBackend = harami.New(conf.instrument, candleDuration)
	case strategy.NameSMA10:
		strategyBackend = sma10.New(conf.instrument, candleDuration)
	case strategy.NameEngulfing:
		strategyBackend = engulfing.New(conf.instrument, candleDuration)
	case strategy.NameRSI:
		strategyBackend = rsi.New(conf.instrument, candleDuration)
	case strategy.NameRSIADX:
		strategyBackend = rsiadx.New(conf.instrument, candleDuration)
	default:
		log.Fatalf("unsupported strategy %q", conf.strategyName)
	}

	log.Info("Starting broker ", conf.broker)
	var dataFeed backtest.Option
	if len(conf.importHistDataCSVFiles) == 0 {
		dataFeed = backtest.WithPriceDBFile(conf.priceDBFile, time.Minute)
	} else {
		dataFeed = backtest.WithTickDataFiles(conf.importHistDataCSVFiles)
	}

	periodFrom := time.Date(conf.yearFrom, 1, 1, 0, 0, 0, 0, time.UTC)
	periodTo := time.Date(conf.yearTo, 12, 31, 23, 23, 59, 0, time.UTC)

	var priceDBOption backtest.Option
	switch conf.priceSource {
	case "PATTERN_TRADING":
		priceDBOption = backtest.WithQuotesSource(backtest.QuotesSourcePatternTrading)
	default:
		priceDBOption = backtest.WithQuotesSource(backtest.QuotesSourceSqlite)
	}

	initialBalance := decimal.NewFromFloat(1000)
	tradingFeePercent := decimal.NewFromFloat(0.01)
	papperWallet := paperwallet.New(
		paperwallet.WithInitialBalance(initialBalance),
		paperwallet.WithTradingFeePercent(tradingFeePercent),
		//paperwallet.WithSlippage(slippageAbsolute),
	)

	brokerBackend := backtest.New(conf.instrument, periodFrom, periodTo, papperWallet, dataFeed,
		backtest.WithCandlePeriod(candleDuration),
		priceDBOption,
	)

	//graph = plotly.NewChart()
	graph = amcharts.NewChart(conf.instrument)

	tr := trader.New(ctx, conf.instrument, "", db,
		trader.WithBroker(brokerBackend),
		trader.WithStrategy(strategyBackend),
		trader.WithCandleSubscription(graph),
		trader.WithPositionSubscription(graph),
	)
	if err := tr.Start(); err != nil {
		log.WithError(err).Fatal("failed to start trader")
	}

	chartHTML, err := graph.RenderChartToHTML()
	if err != nil {
		log.WithError(err).Fatal("Unable to render chart as HTML")
	}
	if err := tr.SavePerformanceRecord(chartHTML); err != nil {
		log.WithError(err).Error("unable to store performance record")
	}
	tr.Summary()

	if err := graph.Start(); err != nil {
		log.WithError(err).Error("failed to start amcharts server")
	}
}
