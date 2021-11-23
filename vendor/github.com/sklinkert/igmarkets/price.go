package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// ResolutionSecond - 1 second price snapshot
	ResolutionSecond = "SECOND"
	// ResolutionMinute - 1 minute price snapshot
	ResolutionMinute = "MINUTE"
	// ResolutionHour - 1 hour price snapshot
	ResolutionHour = "HOUR"
	// ResolutionTwoHour - 2 hour price snapshot
	ResolutionTwoHour = "HOUR_2"
	// ResolutionThreeHour - 3 hour price snapshot
	ResolutionThreeHour = "HOUR_3"
	// ResolutionFourHour - 4 hour price snapshot
	ResolutionFourHour = "HOUR_4"
	// ResolutionDay - 1 day price snapshot
	ResolutionDay = "DAY"
	// ResolutionWeek - 1 week price snapshot
	ResolutionWeek = "WEEK"
	// ResolutionMonth - 1 month price snapshot
	ResolutionMonth = "MONTH"
)

const timeFormat = "2006-01-02T15:04:05"

// PriceResponse - Response for price query
type PriceResponse struct {
	Prices []struct {
		SnapshotTime          string `json:"snapshotTime"`    // "2021/09/24 13:00:00"
		SnapshotTimeUTC       string `json:"snapshotTimeUTC"` // "2021-09-24T11:00:00"
		SnapshotTimeUTCParsed time.Time
		OpenPrice             Price `json:"openPrice"`
		LowPrice              Price `json:"lowPrice"`
		HighPrice             Price `json:"highPrice"`
		ClosePrice            Price `json:"closePrice"`
		LastTradedVolume      int   `json:"lastTradedVolume"`
	}
	InstrumentType string   `json:"instrumentType"`
	MetaData       struct{} `json:"-"`
}

// Price - Subset of PriceResponse
type Price struct {
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	LastTraded float64 `json:"lastTraded"` // Last traded price
}

// GetPriceHistory - Returns a list of historical prices for the given epic, resolution and number of data points
func (ig *IGMarkets) GetPriceHistory(ctx context.Context, epic, resolution string, max int, from, to time.Time) (*PriceResponse, error) {
	var parameters = []string{"pageSize=100"}

	if max > 0 {
		parameters = append(parameters, fmt.Sprintf("max=%d", max))
	}
	if !from.IsZero() {
		parameters = append(parameters, fmt.Sprintf("from=%s", from.Format(timeFormat)))
	}
	if !to.IsZero() {
		parameters = append(parameters, fmt.Sprintf("to=%s", to.Format(timeFormat)))
	}

	bodyReq := new(bytes.Buffer)
	url := fmt.Sprintf("%s/gateway/deal/prices/%s?resolution=%s&%s",
		ig.APIURL, epic, resolution, strings.Join(parameters, "&"))
	req, err := http.NewRequest("GET", url, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get price: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 3, PriceResponse{})
	if err != nil {
		return nil, err
	}
	priceResponse, _ := igResponseInterface.(*PriceResponse)

	for i := range priceResponse.Prices {
		priceResponse.Prices[i].SnapshotTimeUTCParsed, _ =
			time.Parse(timeFormat, priceResponse.Prices[i].SnapshotTimeUTC)
	}

	return priceResponse, err
}

// GetPrice - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPrice(ctx context.Context, epic string) (*PriceResponse, error) {
	return ig.GetPriceHistory(ctx, epic, ResolutionSecond, 1, time.Time{}, time.Time{})
}
