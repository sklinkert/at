package igmarkets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WatchlistsResponse - Response for getting all watchlists
type WatchlistsResponse struct {
	Watchlists []Watchlist
}

// Watchlist - Response from Watchlist endpoint
type Watchlist struct {
	DefaultSystemWatchlist bool   `json:"defaultSystemWatchlist"`
	Deleteable             bool   `json:"deleteable"`
	Editable               bool   `json:"editable"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
}

// WatchlistData - Response from Watchlist endpoint
type WatchlistData struct {
	Markets []MarketData `json:"markets"`
}

// WatchlistRequest - For adding epic to watchlist
type WatchlistRequest struct {
	Epic string `json:"epic"`
}

// CreateWatchlistResponse - Response for creating new watchlist
type CreateWatchlistResponse struct {
	Status      string `json:"status"`
	WatchlistID string `json:"watchlistId"`
}

// CreateWatchlistRequest - Request for creating new watchlist
type CreateWatchlistRequest struct {
	Epics []string `json:"epics"`
	Name  string   `json:"name"`
}

// DeleteFromWatchlist - Delete epic from watchlist
func (ig *IGMarkets) DeleteFromWatchlist(ctx context.Context, watchListID, epic string) error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/gateway/deal/watchlists/%s/%s",
		ig.APIURL, watchListID, epic), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(ctx, req, 1, nil)

	return err
}

// AddToWatchlist - Add epic to existing watchlist
func (ig *IGMarkets) AddToWatchlist(ctx context.Context, watchListID, epic string) error {
	wreq := WatchlistRequest{Epic: epic}
	bodyReq, err := json.Marshal(&wreq)
	if err != nil {
		return fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}

	req, err := http.NewRequest("PUT", ig.APIURL+"/gateway/deal/watchlists/"+watchListID, bytes.NewReader(bodyReq))
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(ctx, req, 1, nil)

	return err
}

// GetWatchlist - Get specific watchlist
func (ig *IGMarkets) GetWatchlist(ctx context.Context, watchListID string) (*WatchlistData, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/watchlists/"+watchListID, bodyReq)
	if err != nil {
		return &WatchlistData{}, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, WatchlistData{})
	igResponse, _ := igResponseInterface.(*WatchlistData)

	return igResponse, err
}

// GetAllWatchlists - Get all watchlist
func (ig *IGMarkets) GetAllWatchlists(ctx context.Context) (*[]Watchlist, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/watchlists", bodyReq)
	if err != nil {
		return &[]Watchlist{}, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, WatchlistsResponse{})
	igResponse, _ := igResponseInterface.(*WatchlistsResponse)

	return &igResponse.Watchlists, err
}

// DeleteWatchlist - Delete whole watchlist
func (ig *IGMarkets) DeleteWatchlist(ctx context.Context, watchListID string) error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/gateway/deal/watchlists/%s",
		ig.APIURL, watchListID), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(ctx, req, 1, nil)

	return err
}

// CreateWatchlist - Create new watchlist
func (ig *IGMarkets) CreateWatchlist(ctx context.Context, name string, epics []string) (watchlistID string, err error) {
	wreq := CreateWatchlistRequest{
		Name:  name,
		Epics: epics,
	}
	bodyReq, err := json.Marshal(&wreq)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/gateway/deal/watchlists",
		ig.APIURL), bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, CreateWatchlistResponse{})
	if err != nil {
		return "", err
	}
	igResponse, _ := igResponseInterface.(*CreateWatchlistResponse)

	return igResponse.WatchlistID, nil
}
