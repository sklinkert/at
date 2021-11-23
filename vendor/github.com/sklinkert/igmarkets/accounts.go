package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

const (
	AccountTypeCFD         = "CFD"
	AccountTypePhysical    = "PHYSICAL"
	AccountTypeSpreadbet   = "SPREADBET"
	AccountStatusEnabed    = "ENABLED"
	AccountStatusDisabled  = "DISABLED"
	AccountStatusSuspended = "SUSPENDED_FROM_DEALING"
)

type Accounts struct {
	Accounts []struct {
		AccountId    string `json:"accountId"`
		AccountName  string `json:"accountName"`
		AccountAlias string `json:"accountAlias"`
		Status       string `json:"status"`
		AccountType  string `json:"accountType"`
		Preferred    bool   `json:"preferred"`
		Balance      struct {
			Balance    float64 `json:"balance"`
			Deposit    float64 `json:"deposit"`
			ProfitLoss float64 `json:"profitLoss"`
			Available  float64 `json:"available"`
		} `json:"balance"`
		Currency        string `json:"currency"`
		CanTransferFrom bool   `json:"canTransferFrom"`
		CanTransferTo   bool   `json:"canTransferTo"`
	} `json:"accounts"`
}

type AccountsPreferences struct {
	TrailingStopsEnabled bool `json:"trailingStopsEnabled"`
}

// GetAccounts - Returns a list of accounts belonging to the logged-in client.
func (ig *IGMarkets) GetAccounts(ctx context.Context) (*Accounts, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/accounts", ig.APIURL), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get accounts: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, Accounts{})
	if err != nil {
		return nil, err
	}
	accounts, _ := igResponseInterface.(*Accounts)

	return accounts, err
}

// GetAccountPreferences - Returns all account related preferences
func (ig *IGMarkets) GetAccountPreferences(ctx context.Context) (*AccountsPreferences, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/accounts/preferences", ig.APIURL), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get account preferences: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, AccountsPreferences{})
	if err != nil {
		return nil, err
	}
	accounts, _ := igResponseInterface.(*AccountsPreferences)

	return accounts, err
}
