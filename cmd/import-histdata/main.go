package main

import (
	"fmt"
	"github.com/lfritz/env"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/atari/pkg/histdatacom"
	"github.com/sklinkert/atari/pkg/ohlc"
	"github.com/sklinkert/atari/pkg/tick"
	//"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var conf struct {
	dbHost     string
	dbUser     string
	dbPassword string
	dbName     string
	dbPort     int
}

func main() {
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
		log.WithError(err).Fatal("env loading failed")
	}

	//dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=require",
	//	conf.dbHost, conf.dbUser, conf.dbPassword, conf.dbName, conf.dbPort)
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	dsn := fmt.Sprintf("./data/%s.db", instrument)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("failed to connect database")
	}

	if err := db.AutoMigrate(&tick.Tick{}, &ohlc.OHLC{}); err != nil {
		log.WithError(err).Fatal("db.AutoMigrate() failed")
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
					log.WithError(err).Fatal("Cannot store candle")
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
			log.Infof("Importing day %s", currentTick.Datetime)
		}
		currentTime = currentTick.Datetime
		imported++
	}
	tx.Commit()
	log.Infof("%d ticks imported", imported)
}
