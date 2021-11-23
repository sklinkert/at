package igmarkets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OTCPositionCloseRequest - request struct for closing positions
type OTCPositionCloseRequest struct {
	DealID      string  `json:"dealId,omitempty"`
	Direction   string  `json:"direction"` // "BUY" or "SELL"
	Epic        string  `json:"epic,omitempty"`
	Expiry      string  `json:"expiry,omitempty"`
	Level       string  `json:"level,omitempty"`
	OrderType   string  `json:"orderType"`
	QuoteID     string  `json:"quoteId,omitempty"`
	Size        float64 `json:"size"`                  // Deal size
	TimeInForce string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
}

// AffectedDeal - part of order confirmation
type AffectedDeal struct {
	DealID   string `json:"dealId"`
	Constant string `json:"constant"` // "FULLY_CLOSED"
}

// DealReference - deal reference struct for responses
type DealReference struct {
	DealReference string `json:"dealReference"`
}

// OTCOrderRequest - request struct for placing orders
type OTCOrderRequest struct {
	Epic                  string  `json:"epic"`
	Level                 string  `json:"level,omitempty"`
	ForceOpen             bool    `json:"forceOpen"`
	OrderType             string  `json:"orderType"`
	CurrencyCode          string  `json:"currencyCode"`
	Direction             string  `json:"direction"` // "BUY" or "SELL"
	Expiry                string  `json:"expiry"`
	Size                  float64 `json:"size"` // Deal size
	StopDistance          string  `json:"stopDistance,omitempty"`
	StopLevel             string  `json:"stopLevel,omitempty"`
	LimitDistance         string  `json:"limitDistance,omitempty"`
	LimitLevel            string  `json:"limitLevel,omitempty"`
	QuoteID               string  `json:"quoteId,omitempty"`
	TimeInForce           string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
	TrailingStop          bool    `json:"trailingStop"`
	TrailingStopIncrement string  `json:"trailingStopIncrement,omitempty"`
	GuaranteedStop        bool    `json:"guaranteedStop"`
	DealReference         string  `json:"dealReference,omitempty"`
}

// WorkingOrderData - Subset of OTCWorkingOrder
type WorkingOrderData struct {
	CreatedDate     string  `json:"createdDate"`
	CreatedDateUTC  string  `json:"createdDateUTC"`
	CurrencyCode    string  `json:"currencyCode"`
	DealID          string  `json:"dealId"`
	Direction       string  `json:"direction"` // "BUY" or "SELL"
	DMA             bool    `json:"dma"`
	Epic            string  `json:"epic"`
	GoodTillDate    string  `json:"goodTillDate"`
	GoodTillDateISO string  `json:"goodTillDateISO"`
	GuaranteedStop  bool    `json:"guaranteedStop"`
	LimitDistance   float64 `json:"limitDistance"`
	OrderLevel      float64 `json:"orderLevel"`
	OrderSize       float64 `json:"orderSize"` // Deal size
	OrderType       string  `json:"orderType"`
	StopDistance    float64 `json:"stopDistance"`
	TimeInForce     string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
}

// OTCDealConfirmation - Deal confirmation
type OTCDealConfirmation struct {
	Epic                  string         `json:"epic"`
	AffectedDeals         []AffectedDeal `json:"affectedDeals"`
	Level                 float64        `json:"level"`
	ForceOpen             bool           `json:"forceOpen"`
	DealStatus            string         `json:"dealStatus"`
	Reason                string         `json:"reason"`
	Status                string         `json:"status"`
	OrderType             string         `json:"orderType"`
	Profit                float64        `json:"profit"`
	ProfitCurrency        string         `json:"profitCurrency"`
	CurrencyCode          string         `json:"currencyCode"`
	Direction             string         `json:"direction"` // "BUY" or "SELL"
	Expiry                string         `json:"expiry,omitempty"`
	Size                  float64        `json:"size"` // Deal size
	StopDistance          float64        `json:"stopDistance"`
	StopLevel             float64        `json:"stopLevel"`
	LimitDistance         float64        `json:"limitDistance,omitempty"`
	LimitLevel            float64        `json:"limitLevel"`
	QuoteID               string         `json:"quoteId,omitempty"`
	TimeInForce           string         `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
	TrailingStop          bool           `json:"trailingStop"`
	TrailingStopIncrement float64        `json:"trailingIncrement"`
	GuaranteedStop        bool           `json:"guaranteedStop"`
	DealReference         string         `json:"dealReference,omitempty"`
}

// OTCUpdateOrderRequest - request struct for updating orders
type OTCUpdateOrderRequest struct {
	StopLevel             float64 `json:"stopLevel"`
	LimitLevel            float64 `json:"limitLevel"`
	TrailingStop          bool    `json:"trailingStop"`
	TrailingStopIncrement string  `json:"trailingStopIncrement,omitempty"`
}

// OTCWorkingOrderRequest - request struct for placing workingorders
type OTCWorkingOrderRequest struct {
	CurrencyCode   string  `json:"currencyCode"`
	DealReference  string  `json:"dealReference,omitempty"`
	Direction      string  `json:"direction"` // "BUY" or "SELL"
	Epic           string  `json:"epic"`
	Expiry         string  `json:"expiry"`
	ForceOpen      bool    `json:"forceOpen"`
	GoodTillDate   string  `json:"goodTillDate,omitempty"`
	GuaranteedStop bool    `json:"guaranteedStop"`
	Level          float64 `json:"level"`
	LimitDistance  string  `json:"limitDistance,omitempty"`
	LimitLevel     string  `json:"limitLevel,omitempty"`
	Size           float64 `json:"size"` // Deal size
	StopDistance   string  `json:"stopDistance,omitempty"`
	StopLevel      string  `json:"stopLevel,omitempty"`
	TimeInForce    string  `json:"timeInForce,omitempty"` // "GOOD_TILL_CANCELLED", "GOOD_TILL_DATE"
	Type           string  `json:"type"`
}

// WorkingOrders - Working orders
type WorkingOrders struct {
	WorkingOrders []OTCWorkingOrder `json:"workingOrders"`
}

// OTCWorkingOrder - Part of WorkingOrders
type OTCWorkingOrder struct {
	MarketData       MarketData       `json:"marketData"`
	WorkingOrderData WorkingOrderData `json:"workingOrderData"`
}

// DeletePositionsOTC - Closes one or more OTC positions
func (ig *IGMarkets) DeletePositionsOTC(ctx context.Context) error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", ig.APIURL+"/gateway/deal/positions/otc", bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(ctx, req, 1, nil)
	return err
}

// PlaceOTCWorkingOrder - Place an OTC workingorder
func (ig *IGMarkets) PlaceOTCWorkingOrder(ctx context.Context, order OTCWorkingOrderRequest) (*DealReference, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to marshal JSON: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/workingorders/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, DealReference{})
	if err != nil {
		return nil, err
	}
	return igResponseInterface.(*DealReference), err
}

// GetOTCWorkingOrders - Get all working orders
func (ig *IGMarkets) GetOTCWorkingOrders(ctx context.Context) (*WorkingOrders, error) {
	bodyReq := new(bytes.Buffer)
	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/workingorders/", bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, WorkingOrders{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*WorkingOrders)

	return igResponse, err
}

// DeleteOTCWorkingOrder - Delete OTC working order
func (ig *IGMarkets) DeleteOTCWorkingOrder(ctx context.Context, dealRef string) (*DealReference, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", ig.APIURL+"/gateway/deal/workingorders/otc/"+dealRef, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, DealReference{})
	if err != nil {
		return nil, err
	}

	return igResponseInterface.(*DealReference), nil
}

// PlaceOTCOrder - Place an OTC order
func (ig *IGMarkets) PlaceOTCOrder(ctx context.Context, order OTCOrderRequest) (*DealReference, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, DealReference{})
	if err != nil {
		return nil, err
	}
	return igResponseInterface.(*DealReference), nil
}

// UpdateOTCOrder - Update an exisiting OTC order
func (ig *IGMarkets) UpdateOTCOrder(ctx context.Context, dealID string, order OTCUpdateOrderRequest) (*DealReference, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("PUT", ig.APIURL+"/gateway/deal/positions/otc/"+dealID, bytes.NewReader(bodyReq))
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, DealReference{})
	if err != nil {
		return nil, err
	}
	return igResponseInterface.(*DealReference), err
}

// CloseOTCPosition - Close an OTC position
func (ig *IGMarkets) CloseOTCPosition(ctx context.Context, close OTCPositionCloseRequest) (*DealReference, error) {
	bodyReq, err := json.Marshal(&close)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return nil, fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	req.Header.Set("_method", "DELETE")

	igResponseInterface, err := ig.doRequest(ctx, req, 1, DealReference{})
	if err != nil {
		return nil, err
	}
	return igResponseInterface.(*DealReference), nil
}

// GetDealConfirmation - Check if the given order was closed/filled
func (ig *IGMarkets) GetDealConfirmation(ctx context.Context, dealRef string) (*OTCDealConfirmation, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/confirms/"+dealRef, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, OTCDealConfirmation{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*OTCDealConfirmation)

	return igResponse, nil
}
