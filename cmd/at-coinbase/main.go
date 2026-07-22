package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/sklinkert/at/internal/broker/coinbase"
	"github.com/sklinkert/at/internal/paperwallet"
	"github.com/sklinkert/at/internal/strategy/rsiadx"
	"github.com/sklinkert/at/internal/trader"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Example setup for Coinbase broker

func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

func mustConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("at-demo.db"), &gorm.Config{})
	if err != nil {
		fatal("failed to connect database", err)
	}
	if err := db.AutoMigrate(&tick.Tick{}, &ohlc.OHLC{}, &trader.PerformanceRecord{}); err != nil {
		fatal("db.AutoMigrate() failed", err)
	}
	return db
}

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	instrument := "BTC-USD"
	candleDuration := time.Minute * 1
	strategyBackend := rsiadx.New(instrument, candleDuration)
	wallet := paperwallet.New()
	brokerBackend := coinbase.New(instrument, wallet)

	db := mustConnectDB()

	tr := trader.New(ctx, instrument, "", db,
		trader.WithBroker(brokerBackend),
		trader.WithPersistCandleData(true),
		trader.WithStrategy(strategyBackend),
		trader.WithFeedStoredCandles(strategyBackend),
	)
	if err := tr.Start(); err != nil {
		fatal("failed to start trader", err)
	}
	tr.Summary()
}
