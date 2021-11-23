package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// MarketData - Subset of OTCWorkingOrder
type MarketData struct {
	Bid                      float64 `json:"bid"`
	DelayTime                int     `json:"delayTime"`
	Epic                     string  `json:"epic"`
	ExchangeID               string  `json:"exchangeId"`
	Expiry                   string  `json:"expiry"`
	High                     float64 `json:"high"`
	InstrumentName           string  `json:"instrumentName"`
	InstrumentType           string  `json:"instrumentType"`
	LotSize                  float64 `json:"lotSize"`
	Low                      float64 `json:"low"`
	MarektStatus             string  `json:"marketStatus"`
	NetChange                float64 `json:"netChange"`
	Offer                    float64 `json:"offer"`
	PercentageChange         float64 `json:"percentageChange"`
	ScalingFactor            int     `json:"scalingFactor"`
	StreamingPricesAvailable bool    `json:"streamingPricesAvailable"`
	UpdateTime               string  `json:"updateTime"`
	UpdateTimeUTC            string  `json:"updateTimeUTC"`
}

// MarketSearchResponse - Contains the response data for MarketSearch()
type MarketSearchResponse struct {
	Markets []MarketData `json:"markets"`
}

// MarketsResponse - Market response for /markets/{epic}
type MarketsResponse struct {
	DealingRules DealingRules `json:"dealingRules"`
	Instrument   Instrument   `json:"instrument"`
	Snapshot     Snapshot     `json:"snapshot"`
}

// DealingRules - Part of MarketsResponse
type DealingRules struct {
	MarketOrderPreference         string         `json:"marketOrderPreference"`
	TrailingStopsPreference       string         `json:"trailingStopsPreference"`
	MaxStopOrLimitDistance        UnitValueFloat `json:"maxStopOrLimitDistance"`
	MinControlledRiskStopDistance UnitValueFloat `json:"minControlledRiskStopDistance"`
	MinDealSize                   UnitValueFloat `json:"minDealSize"`
	MinNormalStopOrLimitDistance  UnitValueFloat `json:"minNormalStopOrLimitDistance"`
	MinStepDistance               UnitValueFloat `json:"minStepDistance"`
}

// Currency - Part of MarketsResponse
type Currency struct {
	BaseExchangeRate float64 `json:"baseExchangeRate"`
	Code             string  `json:"code"`
	ExchangeRate     float64 `json:"exchangeRate"`
	IsDefault        bool    `json:"isDefault"`
	Symbol           string  `json:"symbol"`
}

// UnitValueFloat - Part of MarketsResponse
type UnitValueFloat struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// Instrument - Part of MarketsResponse
type Instrument struct {
	ChartCode                string         `json:"chartCode"`
	ControlledRiskAllowed    bool           `json:"controlledRiskAllowed"`
	Country                  string         `json:"country"`
	Currencies               []Currency     `json:"currencies"`
	Epic                     string         `json:"epic"`
	Expiry                   string         `json:"expiry"`
	StreamingPricesAvailable bool           `json:"streamingPricesAvailable"`
	ForceOpenAllowed         bool           `json:"forceOpenAllowed"`
	Unit                     string         `json:"unit"`
	Type                     string         `json:"type"`
	MarketID                 string         `json:"marketID"`
	LotSize                  float64        `json:"lotSize"`
	MarginFactor             float64        `json:"marginFactor"`
	MarginFactorUnit         string         `json:"marginFactorUnit"`
	SlippageFactor           UnitValueFloat `json:"slippageFactor"`
	LimitedRiskPremium       UnitValueFloat `json:"limitedRiskPremium"`
	NewsCode                 string         `json:"newsCode"`
	ValueOfOnePip            string         `json:"valueOfOnePip"`
	OnePipMeans              string         `json:"onePipMeans"`
	ContractSize             string         `json:"contractSize"`
	SpecialInfo              []string       `json:"specialInfo"`
}

// Snapshot - Part of MarketsResponse
type Snapshot struct {
	MarketStatus              string  `json:"marketStatus"`
	NetChange                 float64 `json:"netChange"`
	PercentageChange          float64 `json:"percentageChange"`
	UpdateTime                string  `json:"updateTime"`
	DelayTime                 float64 `json:"delayTime"`
	Bid                       float64 `json:"bid"`
	Offer                     float64 `json:"offer"`
	High                      float64 `json:"high"`
	Low                       float64 `json:"low"`
	DecimalPlacesFactor       float64 `json:"decimalPlacesFactor"`
	ScalingFactor             float64 `json:"scalingFactor"`
	ControlledRiskExtraSpread float64 `json:"controlledRiskExtraSpread"`
}

// MarketSearch - Search for ISIN or share names to get the epic.
func (ig *IGMarkets) MarketSearch(ctx context.Context, term string) (*MarketSearchResponse, error) {
	bodyReq := new(bytes.Buffer)

	// E.g. https://demo-api.ig.com/gateway/deal/markets?searchTerm=DE0005008007
	url := fmt.Sprintf("%s/gateway/deal/markets?searchTerm=%s", ig.APIURL, term)
	req, err := http.NewRequest("GET", url, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, MarketSearchResponse{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*MarketSearchResponse)

	return igResponse, err
}

// GetMarkets - Return markets information for given epic
func (ig *IGMarkets) GetMarkets(ctx context.Context, epic string) (*MarketsResponse, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/markets/%s",
		ig.APIURL, epic), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 3, MarketsResponse{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*MarketsResponse)

	return igResponse, err
}
