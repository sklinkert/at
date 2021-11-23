package igmarkets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// session - IG auth response (OAuth only)
// Version 3
type session struct {
	ClientID              string     `json:"clientId"`
	AccountId             string     `json:"accountId"`
	LightstreamerEndpoint string     `json:"lightstreamerEndpoint"`
	OAuthToken            OAuthToken `json:"oauthToken"`
	TimezoneOffset        int        `json:"timezoneOffset"` // In hours
}

// SessionVersion2 - IG auth response
// required for LightStreamer API
type SessionVersion2 struct {
	AccountType           string `json:"accountType"`      // "CFD"
	CurrencyIsoCode       string `json:"currencyIsoCode"`  // "EUR"
	CurrencySymbol        string `json:"currencySymbol"`   // "E"
	CurrentAccountId      string `json:"currentAccountId"` // "ABDGS"
	LightstreamerEndpoint string `json:"lightstreamerEndpoint"`
	ClientID              string `json:"clientId"`
	TimezoneOffset        int    `json:"timezoneOffset"` // In hours
	HasActiveDemoAccounts bool   `json:"hasActiveDemoAccounts"`
	HasActiveLiveAccounts bool   `json:"hasActiveLiveAccounts"`
	TrailingStopsEnabled  bool   `json:"trailingStopsEnabled"`
	DealingEnabled        bool   `json:"dealingEnabled"`

	// Extracted from HTTP Header
	CSTToken string
	XSTToken string
}

// OAuthToken - part of the session
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// RefreshToken - Get new OAuthToken from API and set it to IGMarkets object
func (ig *IGMarkets) RefreshToken(ctx context.Context) error {
	bodyReq := new(bytes.Buffer)

	var authReq = refreshTokenRequest{
		RefreshToken: ig.OAuthToken.RefreshToken,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session/refresh-token"), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 1, OAuthToken{})
	if err != nil {
		return err
	}
	oauthToken, _ := igResponseInterface.(*OAuthToken)

	if oauthToken.AccessToken == "" {
		return fmt.Errorf("igmarkets: got response but access token is empty")
	}

	expiry, err := strconv.ParseInt(oauthToken.ExpiresIn, 10, 32)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to parse OAuthToken expiry field: %v", err)
	}

	// Refresh token before it will expire
	if expiry <= 10 {
		return fmt.Errorf("igmarkets: token expiry is too short for periodically renewals")
	}

	ig.Lock()
	ig.OAuthToken = *oauthToken
	ig.Unlock()

	return nil
}

// Login - Get new OAuthToken from API and set it to IGMarkets object
func (ig *IGMarkets) Login(ctx context.Context) error {
	bodyReq := new(bytes.Buffer)

	var authReq = authRequest{
		Identifier: ig.Identifier,
		Password:   ig.Password,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session"), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequestWithoutOAuth(ctx, req, 3, session{})
	if err != nil {
		return err
	}
	session, _ := igResponseInterface.(*session)

	if session.OAuthToken.AccessToken == "" {
		return fmt.Errorf("igmarkets: got response but access token is empty")
	}

	expiry, err := strconv.ParseInt(session.OAuthToken.ExpiresIn, 10, 32)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to parse OAuthToken expiry field: %v", err)
	}

	// Refresh token before it will expire
	if expiry <= 10 {
		return fmt.Errorf("igmarkets: token expiry is too short for periodically renewals")
	}

	ig.Lock()
	ig.OAuthToken = session.OAuthToken
	ig.TimeZone = timeZoneOffset2Location(session.TimezoneOffset)
	ig.Unlock()

	return nil
}

// LoginVersion2 - use old login version. contains required data for LightStreamer API
func (ig *IGMarkets) LoginVersion2(ctx context.Context) (*SessionVersion2, error) {
	bodyReq := new(bytes.Buffer)

	var authReq = authRequest{
		Identifier: ig.Identifier,
		Password:   ig.Password,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return nil, fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session"), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}

	igResponseInterface, headers, err := ig.doRequestWithResponseHeaders(ctx, req, 2, SessionVersion2{}, false)
	if err != nil {
		return nil, err
	}
	session, _ := igResponseInterface.(*SessionVersion2)
	if headers != nil {
		session.CSTToken = headers.Get("CST")
		session.XSTToken = headers.Get("X-SECURITY-TOKEN")
	}
	return session, nil
}
