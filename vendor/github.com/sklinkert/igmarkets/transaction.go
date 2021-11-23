package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

// HistoryTransactionResponse - Response for  transactions endpoint
type HistoryTransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
	MetaData     struct {
		PageData struct {
			PageNumber int `json:"pageNumber"`
			PageSize   int `json:"pageSize"`
			TotalPages int `json:"totalPages"`
		} `json:"pageData"`
		Size int `json:"size"`
	} `json:"metaData"`
}

// Transaction - Part of HistoryTransactionResponse
type Transaction struct {
	CashTransaction bool   `json:"cashTransaction"`
	CloseLevel      string `json:"closeLevel"`
	Currency        string `json:"currency"`
	Date            string `json:"date"`
	DateUTC         string `json:"dateUtc"`
	InstrumentName  string `json:"instrumentName"`
	OpenDateUtc     string `json:"openDateUtc"`
	OpenLevel       string `json:"openLevel"`
	Period          string `json:"period"`
	ProfitAndLoss   string `json:"profitAndLoss"`
	Reference       string `json:"reference"`
	Size            string `json:"size"`
	TransactionType string `json:"transactionType"`
}

// GetTransactions - Return all transaction
func (ig *IGMarkets) GetTransactions(ctx context.Context, transactionType string, from time.Time) (*HistoryTransactionResponse, error) {
	bodyReq := new(bytes.Buffer)
	fromStr := from.Format("2006-01-02T15:04:05")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/history/transactions?from=%s&type=%s&pageSize=0",
		ig.APIURL, fromStr, transactionType), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get transactions: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 2, HistoryTransactionResponse{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*HistoryTransactionResponse)

	return igResponse, err
}
