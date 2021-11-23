package main

import (
	"context"
	"fmt"
	"github.com/lfritz/env"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker/ig"
	"github.com/sklinkert/at/internal/strategy"
	heikinashi "github.com/sklinkert/at/internal/strategy/HeikinAshi"
	"github.com/sklinkert/at/internal/strategy/doji"
	"github.com/sklinkert/at/internal/strategy/lowcandle"
	"github.com/sklinkert/at/internal/strategy/rsi"
	"github.com/sklinkert/at/internal/strategy/rsiadx"
	"github.com/sklinkert/at/internal/strategy/scalper"
	"github.com/sklinkert/at/internal/strategy/stochrsi"
	"github.com/sklinkert/at/internal/trader"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/igmarkets"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

// Example setup for IG.com broker

// GitRev is injected by compiler by -X main.Version=$(VERSION)
var GitRev string

var conf struct {
	importHistDataCSVFiles []string
	broker                 string
	gatherPerformanceData  bool
	instrument             string
	debug                  bool
	dbHost                 string
	dbUser                 string
	dbPassword             string
	dbName                 string
	dbPort                 int
	strategyName           string
	priceDBFile            string
	yearFrom               int
	yearTo                 int
	candleDuration         string
	igAPIURL               string
	igIdentifier           string
	igAPIKey               string
	igPassword             string
	igAccountID            string
	currencyCode           string
}

func mustConnectDB() *gorm.DB {
	var err error
	var db *gorm.DB
	if conf.dbHost == "sqlite" {
		db, err = gorm.Open(sqlite.Open("at-demo.db"), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=require",
			conf.dbHost, conf.dbUser, conf.dbPassword, conf.dbName, conf.dbPort)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		log.WithError(err).Fatal("failed to connect database")
	}
	if err := db.AutoMigrate(&tick.Tick{}, &ohlc.OHLC{}, &trader.PerformanceRecord{}); err != nil {
		log.WithError(err).Fatal("db.AutoMigrate() failed")
	}
	return db
}

func main() {
	var ctx = context.Background()

	var err error
	var e = env.New()
	e.Flag("DEBUG", &conf.debug, "Enable debug logging")
	e.Flag("PERFORMANCE_DATA", &conf.gatherPerformanceData, "Gather performance data and print as CSV")
	e.OptionalList("IMPORT_HISTDATA_CSV_FILES", &conf.importHistDataCSVFiles, ",", []string{}, "Import CSV files from histdata.com")
	e.OptionalString("PRICE_DB_FILE", &conf.priceDBFile, "/data/EURUSD/db", "SQLite DB file for price data (OHLCs)")
	e.OptionalString("INSTRUMENT", &conf.instrument, "CS.D.EURUSD.MINI.IP", "instrument to trade")
	e.OptionalString("CURRENCY_CODE", &conf.currencyCode, "EUR", "Currency code")
	e.OptionalString("BROKER", &conf.broker, "none", "Broker backend")
	e.OptionalString("STRATEGY", &conf.strategyName, "meanreversion", "strategy to be executed")
	e.OptionalString("CANDLE_DURATION", &conf.candleDuration, "60m", "Duration for OHLC candle")
	e.OptionalString("IG_API_URL", &conf.igAPIURL, igmarkets.DemoAPIURL, "IG API URL")
	e.OptionalString("IG_IDENTIFIER", &conf.igIdentifier, "", "IG Identifier")
	e.OptionalString("IG_API_KEY", &conf.igAPIKey, "", "IG API key")
	e.OptionalString("IG_PASSWORD", &conf.igPassword, "", "IG password")
	e.OptionalString("IG_ACCOUNT", &conf.igAccountID, "", "IG account ID")
	e.OptionalString("DB_HOST", &conf.dbHost, "guest", "DB host")
	e.OptionalString("DB_USER", &conf.dbUser, "guest", "DB user")
	e.OptionalString("DB_PASSWORD", &conf.dbPassword, "guest", "DB password")
	e.OptionalString("DB_NAME", &conf.dbName, "guest", "DB name")
	e.OptionalInt("DB_PORT", &conf.dbPort, 25060, "DB port")
	e.OptionalInt("YEAR_FROM", &conf.yearFrom, 1970, "Backtesting beginning")
	e.OptionalInt("YEAR_TO", &conf.yearTo, 2022, "Backtesting end")
	if err := e.Load(); err != nil {
		log.WithError(err).Fatal("env loading failed")
	}
	if conf.debug {
		log.SetLevel(log.DebugLevel)
	}
	if GitRev == "" {
		log.Fatal("GitRev not set")
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
	case strategy.NameRSI:
		strategyBackend = rsi.New(conf.instrument, candleDuration)
	case strategy.NameRSIADX:
		strategyBackend = rsiadx.New(conf.instrument, candleDuration)
	default:
		log.Fatalf("unsupported strategy %q", conf.strategyName)
	}

	log.Info("Starting broker ", conf.broker)

	brokerBackend, err := ig.New(conf.instrument, conf.igAPIURL, conf.igAPIKey, conf.igAccountID,
		conf.igIdentifier, conf.igPassword)
	if err != nil {
		log.WithError(err).Fatal("ig.New() failed")
	}

	db := mustConnectDB()

	tr := trader.New(ctx, conf.instrument, GitRev, db,
		trader.WithBroker(brokerBackend),
		trader.WithPersistCandleData(true),
		trader.WithStrategy(strategyBackend),
		trader.WithFeedStoredCandles(strategyBackend),
		trader.WithCurrencyCode(conf.currencyCode),
	)
	if err := tr.Start(); err != nil {
		log.WithError(err).Fatal("failed to start trader")
	}
	tr.Summary()
}
