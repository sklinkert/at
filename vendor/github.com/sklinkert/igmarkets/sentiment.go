package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// ClientSentimentResponse - Response for client sentiment
type ClientSentimentResponse struct {
	LongPositionPercentage  float64 `json:"longPositionPercentage"`
	ShortPositionPercentage float64 `json:"shortPositionPercentage"`
}

// GetClientSentiment - Get the client sentiment for the given instrument's market
func (ig *IGMarkets) GetClientSentiment(ctx context.Context, MarketID string) (*ClientSentimentResponse, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/clientsentiment/"+MarketID, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request for GetClientSentiment: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, ClientSentimentResponse{})
	igResponse, _ := igResponseInterface.(*ClientSentimentResponse)
	return igResponse, err
}
