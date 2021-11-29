package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker/coinbase"
	"github.com/sklinkert/at/internal/paperwallet"
	"github.com/sklinkert/at/internal/strategy/rsiadx"
	"github.com/sklinkert/at/internal/trader"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

// Example setup for Coinbase broker

func mustConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("at-demo.db"), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("failed to connect database")
	}
	if err := db.AutoMigrate(&tick.Tick{}, &ohlc.OHLC{}, &trader.PerformanceRecord{}); err != nil {
		log.WithError(err).Fatal("db.AutoMigrate() failed")
	}
	return db
}

func main() {
	ctx := context.Background()
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
		log.WithError(err).Fatal("failed to start trader")
	}
	tr.Summary()
}
