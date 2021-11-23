package histdatacom

import (
	"encoding/csv"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/tick"
	"io"
	"os"
	"time"
)

func ImportFromCSV(instrument string, files []string, tickChan chan tick.Tick) {
	for _, file := range files {
		importFromCSV(instrument, file, tickChan)
	}
	close(tickChan)
}

func importFromCSV(instrument, file string, tickChan chan tick.Tick) {
	clog := log.WithFields(log.Fields{
		"FILE": file,
	})

	clog.Info("Going to open tick data file")
	csvFile, err := os.Open(file)
	if err != nil {
		clog.WithError(err).Fatal("unable to open CSV file")
	}
	defer func() {
		if err := csvFile.Close(); err != nil {
			clog.WithError(err).Warn("cannot close file handle")
		}
	}()

	// Eastern Standard Time (EST) time-zone WITHOUT Day Light Savings adjustments
	// http://www.histdata.com/f-a-q/data-files-detailed-specification/
	loc, err := time.LoadLocation("EST")
	if err != nil {
		clog.WithError(err).Fatal("cannot load timezone")
	}

	// DateTime Stamp;Bar OPEN Bid Quote;Bar HIGH Bid Quote;Bar LOW Bid Quote;Bar CLOSE Bid Quote;Volume
	r := csv.NewReader(csvFile)
	var tmp decimal.Decimal
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			clog.WithError(err).Fatal("reading CSV file failed: ", file)
		}
		if len(record) != 4 {
			clog.Warn("Ignoring malformed line: ", record)
			continue
		}

		timeStr := record[0]
		datetime, err := time.ParseInLocation("20060102 150405", timeStr[0:len(timeStr)-3], loc)
		if err != nil {
			clog.WithError(err).Warn("cannot parse time", timeStr)
			continue
		}

		bid, err := decimal.NewFromString(record[1])
		if err != nil {
			clog.WithError(err).Warn("cannot parse bid", record[1])
			continue
		}
		ask, err := decimal.NewFromString(record[2])
		if err != nil {
			clog.WithError(err).Warn("cannot parse ask", record[2])
			continue
		}

		if bid.GreaterThan(ask) {
			// Some histdata files contain a wrong bid/ask order
			tmp = bid
			bid = ask
			ask = tmp
		}

		tickData := tick.New(instrument, datetime, bid, ask)
		tickChan <- tickData

		//log.Infof("IMPORT: %s %s %s", datetime, bid, ask)
	}
}
