package igmarkets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"
)

const (
	// DemoAPIURL - Demo API URL
	DemoAPIURL = "https://demo-api.ig.com"
	// LiveAPIURL - Live API URL - Real trading!
	LiveAPIURL = "https://api.ig.com"
)

// IGMarkets - Object with all information we need to access IG REST API
type IGMarkets struct {
	APIURL                string
	APIKey                string
	AccountID             string
	Identifier            string
	Password              string
	TimeZone              *time.Location
	TimeZoneLightStreamer *time.Location
	OAuthToken            OAuthToken
	httpClient            *http.Client
	sync.RWMutex
}

// New - Create new instance of igmarkets
func New(apiURL, apiKey, accountID, identifier, password string) *IGMarkets {
	if apiURL != DemoAPIURL && apiURL != LiveAPIURL {
		log.Panic("Invalid endpoint URL", apiURL)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
	}

	return &IGMarkets{
		APIURL:     apiURL,
		APIKey:     apiKey,
		AccountID:  accountID,
		Identifier: identifier,
		Password:   password,
		httpClient: httpClient,
	}
}

func (ig *IGMarkets) doRequestWithoutOAuth(ctx context.Context, req *http.Request, endpointVersion int, igResponse interface{}) (interface{}, error) {
	object, _, err := ig.doRequestWithResponseHeaders(ctx, req, endpointVersion, igResponse, false)
	return object, err
}

func (ig *IGMarkets) doRequest(ctx context.Context, req *http.Request, endpointVersion int, igResponse interface{}) (interface{}, error) {
	object, _, err := ig.doRequestWithResponseHeaders(ctx, req, endpointVersion, igResponse, true)
	return object, err
}

func (ig *IGMarkets) doRequestWithResponseHeaders(ctx context.Context, req *http.Request, endpointVersion int, igResponse interface{}, oAuth bool) (interface{}, http.Header, error) {
	ig.RLock()
	if ig.OAuthToken.AccessToken != "" && oAuth {
		req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	}
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.RUnlock()

	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", fmt.Sprintf("%d", endpointVersion))

	req = req.WithContext(ctx)
	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return igResponse, nil, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("igmarkets.doRequest:  resp.Body.Close() failed: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return igResponse, nil, fmt.Errorf("igmarkets: unable to get body of transactions markets data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return igResponse, nil, fmt.Errorf("igmarkets: unexpected HTTP status code: %d (body=%q)", resp.StatusCode, body)
	}

	if igResponse != nil {
		objType := reflect.TypeOf(igResponse)
		obj := reflect.New(objType).Interface()
		if obj != nil {
			if err := json.Unmarshal(body, &obj); err != nil {
				return obj, nil, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
			}

			return obj, resp.Header, nil
		}
	}

	return igResponse, resp.Header, nil
}
