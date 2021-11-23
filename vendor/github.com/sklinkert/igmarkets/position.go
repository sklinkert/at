package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// PositionsResponse - Response from positions endpoint
type PositionsResponse struct {
	Positions []Position `json:"positions"`
}

// Position - part of PositionsResponse
type Position struct {
	MarketData MarketData `json:"market"`
	Position   struct {
		ContractSize         float64 `json:"contractSize"`
		ControlledRisk       bool    `json:"controlledRisk"`
		CreatedDate          string  `json:"createdDate"`
		CreatedDateUTC       string  `json:"createdDateUTC"`
		Currencry            string  `json:"currency"`
		DealID               string  `json:"dealId"`
		DealReference        string  `json:"dealReference"`
		Direction            string  `json:"direction"`
		Level                float64 `json:"level"`
		LimitLevel           float64 `json:"limitLevel"`
		Size                 float64 `json:"size"`
		StopLevel            float64 `json:"stopLevel"`
		TrailingStep         float64 `json:"trailingStep"`
		TrailingStopDistance float64 `json:"trailingStopDistance"`
	} `json:"position"`
}

// GetPositions - Get all open positions
func (ig *IGMarkets) GetPositions(ctx context.Context) (*PositionsResponse, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/positions/", bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, PositionsResponse{})
	if err != nil {
		return nil, err
	}

	igResponse, _ := igResponseInterface.(*PositionsResponse)
	return igResponse, nil
}
