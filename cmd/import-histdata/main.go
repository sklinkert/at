package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lfritz/env"
	"github.com/sklinkert/at/pkg/histdatacom"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	//"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

var conf struct {
	dbHost     string
	dbUser     string
	dbPassword string
	dbName     string
	dbPort     int
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	var c = make(chan tick.Tick)
	var e = env.New()
	var csvFiles []string
	var instrument string
	e.List("IMPORT_HISTDATA_CSV_FILES", &csvFiles, ",", "Import CSV files from histdata.com")
	e.String("INSTRUMENT", &instrument, "Instrument name e.g. EURUSD")
	e.OptionalString("DB_HOST", &conf.dbHost, "", "DB host")
	e.OptionalString("DB_USER", &conf.dbUser, "guest", "DB user")
	e.OptionalString("DB_PASSWORD", &conf.dbPassword, "guest", "DB password")
	e.OptionalString("DB_NAME", &conf.dbName, "guest", "DB name")
	e.OptionalInt("DB_PORT", &conf.dbPort, 25060, "DB port")
	if err := e.Load(); err != nil {
		fatal("env loading failed", err)
	}

	//dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=require",
	//	conf.dbHost, conf.dbUser, conf.dbPassword, conf.dbName, conf.dbPort)
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	const dataDir = "./data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		fatal("cannot create data directory", err)
	}
	dsn := fmt.Sprintf("%s/%s.db", dataDir, instrument)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		fatal("failed to connect database", err)
	}

	if err := db.AutoMigrate(&tick.Tick{}, &ohlc.OHLC{}); err != nil {
		fatal("db.AutoMigrate() failed", err)
	}

	go histdatacom.ImportFromCSV(instrument, csvFiles, c)

	// sqlite pragmas; source: https://avi.im/blag/2021/fast-sqlite-inserts/
	db.Exec("PRAGMA journal_mode = OFF;")
	db.Exec("PRAGMA synchronous = 0;")
	db.Exec("PRAGMA locking_mode = EXCLUSIVE;")

	tx := db.Begin()

	var currentTime time.Time
	var imported uint
	var candle *ohlc.OHLC
	const candleDuration = time.Minute * 1
	for currentTick := range c {
		if candle != nil {
			isOpen := candle.NewPrice(currentTick.Price(), currentTick.Datetime)
			if !isOpen {
				if err := tx.Create(candle).Error; err != nil {
					fatal("Cannot store candle", err)
				}
				candle = nil // force new candle opening
			}
		}
		if candle == nil {
			candle = ohlc.New(instrument, currentTick.Datetime, candleDuration, true)
			candle.NewPrice(currentTick.Price(), currentTick.Datetime)
		}
		//if err := tx.Create(&currentTick).Error; err != nil {
		//	log.WithError(err).Warn("db.Create() failed: %v", currentTick)
		//	continue
		//}
		if imported%1000 == 0 {
			tx.Commit()
			tx = db.Begin()
		}

		if currentTime.Day() != currentTick.Datetime.Day() {
			slog.Info(fmt.Sprintf("Importing day %s", currentTick.Datetime))
		}
		currentTime = currentTick.Datetime
		imported++
	}
	tx.Commit()
	slog.Info(fmt.Sprintf("%d ticks imported", imported))
}
