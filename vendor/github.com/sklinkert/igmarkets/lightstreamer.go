package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LightStreamerTick struct {
	Epic string
	Time time.Time
	Bid  float64
	Ask  float64
}

const lightStreamerContentType = "application/x-www-form-urlencoded"

// OpenLightStreamerSubscription GetOTCWorkingOrders - Get all working orders
// epic: e.g. CS.D.BITCOIN.CFD.IP
// tickReceiver: receives all ticks from lightstreamer API
func (ig *IGMarkets) OpenLightStreamerSubscription(ctx context.Context, epics []string, tickReceiver chan LightStreamerTick) error {
	// Obtain CST and XST tokens first
	sessionVersion2, err := ig.LoginVersion2(ctx)
	if err != nil {
		return fmt.Errorf("ig.LoginVersion2() failed: %v", err)
	}

	timeZone, err := time.LoadLocation("Europe/London")
	if err != nil {
		return err
	}

	ig.Lock()
	ig.TimeZoneLightStreamer = timeZone
	ig.Unlock()

	tr := &http.Transport{
		MaxIdleConns:       1,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	c := &http.Client{Transport: tr}

	sessionID, sessionMsg, err := ig.lightStreamerConnect(c, sessionVersion2)
	if err != nil {
		return err
	}

	if err := ig.lightStreamerSubscribe(c, sessionVersion2.LightstreamerEndpoint, sessionID, sessionMsg, epics); err != nil {
		return err
	}

	httpStream, err := ig.lightStreamerBindToConnection(c, sessionVersion2.LightstreamerEndpoint, sessionID)
	if err != nil {
		return err
	}

	go ig.lightstreamerReadSubscription(epics, tickReceiver, httpStream)

	return nil
}

// connectToLightStream - Create new lightstreamer session
func (ig *IGMarkets) lightStreamerConnect(client *http.Client, sessionVersion2 *SessionVersion2) (sessionID, sessionMsg string, err error) {
	body := []byte("LS_polling=true&LS_polling_millis=0&LS_idle_millis=0&LS_op2=create&LS_password=CST-" +
		sessionVersion2.CSTToken + "|" + "XST-" + sessionVersion2.XSTToken + "&LS_user=" +
		sessionVersion2.CurrentAccountId + "&LS_cid=mgQkwtwdysogQz2BJ4Ji kOj2Bg")
	bodyBuf := bytes.NewBuffer(body)
	url := fmt.Sprintf("%s/lightstreamer/create_session.txt", sessionVersion2.LightstreamerEndpoint)
	resp, err := client.Post(url, lightStreamerContentType, bodyBuf)
	if err != nil {
		if resp != nil {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return "", "", fmt.Errorf("calling lightstreamer endpoint %s failed: %v; reading HTTP body also failed: %v",
					url, err, err2)
			}
			return "", "", fmt.Errorf("calling lightstreamer endpoint %s failed: %v http.StatusCode:%d Body: %q",
				url, err, resp.StatusCode, string(body))
		}
		return "", "", fmt.Errorf("calling lightstreamer endpoint %q failed: %v", url, err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	sessionMsg = string(respBody[:])
	if !strings.HasPrefix(sessionMsg, "OK") {
		return "", "", fmt.Errorf("unexpected response from lightstreamer session endpoint %q: %q", url, sessionMsg)
	}
	sessionParts := strings.Split(sessionMsg, "\r\n")
	sessionID = sessionParts[1]
	sessionID = strings.ReplaceAll(sessionID, "SessionId:", "")
	return sessionID, sessionMsg, nil
}

// lightStreamerSubscribe - Adding subscription for epics
func (ig *IGMarkets) lightStreamerSubscribe(client *http.Client, lightStreamerEndpoint, sessionID, sessionMsg string, epics []string) error {
	var epicList string
	for _, epic := range epics {
		epicList = epicList + "MARKET:" + epic + "+"
	}
	body := []byte("LS_session=" + sessionID +
		"&LS_polling=true&LS_polling_millis=0&LS_idle_millis=0&LS_op=add&LS_Table=1&LS_id=" +
		epicList + "&LS_schema=UPDATE_TIME+BID+OFFER+MARKET_STATE&LS_mode=MERGE")
	bodyBuf := bytes.NewBuffer(body)
	url := fmt.Sprintf("%s/lightstreamer/control.txt", lightStreamerEndpoint)
	resp, err := client.Post(url, lightStreamerContentType, bodyBuf)
	if err != nil {
		if resp != nil {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return fmt.Errorf("calling lightstreamer endpoint %s failed: %v; reading HTTP body also failed: %v",
					url, err, err2)
			}
			return fmt.Errorf("calling lightstreamer endpoint %q failed: %v http.StatusCode:%d Body: %q",
				url, err, resp.StatusCode, string(body))
		}
		return fmt.Errorf("calling lightstreamer endpoint %q failed: %v", url, err)
	}
	body, _ = ioutil.ReadAll(resp.Body)
	if !strings.HasPrefix(sessionMsg, "OK") {
		return fmt.Errorf("unexpected control.txt response: %q", body)
	}
	return nil
}

func (ig *IGMarkets) lightStreamerBindToConnection(client *http.Client, lightStreamerEndpoint, sessionID string) (httpStream *http.Response, err error) {
	body := []byte("LS_session=" + sessionID + "&LS_polling=false&LS_polling_millis=0&LS_idle_millis=0")
	bodyBuf := bytes.NewBuffer(body)
	url := fmt.Sprintf("%s/lightstreamer/bind_session.txt", lightStreamerEndpoint)
	resp, err := client.Post(url, lightStreamerContentType, bodyBuf)
	if err != nil {
		if resp != nil {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return nil, fmt.Errorf("calling lightstreamer endpoint %s failed: %v; reading HTTP body also failed: %v",
					url, err, err2)
			}
			return nil, fmt.Errorf("calling lightstreamer endpoint %q failed: %v http.StatusCode:%d Body: %q",
				url, err, resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("calling lightstreamer endpoint %q failed: %v", url, err)
	}
	return resp, nil
}

func (ig *IGMarkets) lightstreamerReadSubscription(epics []string, tickReceiver chan LightStreamerTick, resp *http.Response) {
	const epicNameUnknown = "unknown"
	var respBuf = make([]byte, 64)
	var lastTicks = make(map[string]LightStreamerTick, len(epics)) // epic -> tick

	defer close(tickReceiver)

	// map table index -> epic name
	var epicIndex = make(map[string]string, len(epics))
	for i, epic := range epics {
		epicIndex[fmt.Sprintf("1,%d", i+1)] = epic
	}

	for {
		read, err := resp.Body.Read(respBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("reading lightstreamer subscription failed: %v", err)
			break
		}

		priceMsg := string(respBuf[0:read])
		priceParts := strings.Split(priceMsg, "|")

		// Sever ends streaming
		if priceMsg == "LOOP\r\n\r\n" {
			fmt.Printf("ending\n")
			break
		}

		if len(priceParts) != 5 {
			//fmt.Printf("Malformed price message: %q\n", priceMsg)
			continue
		}

		var parsedTime time.Time
		if priceParts[1] != "" {
			priceTime := priceParts[1]
			now := time.Now().In(ig.TimeZoneLightStreamer)
			parsedTime, err = time.ParseInLocation("2006-1-2 15:04:05", fmt.Sprintf("%d-%d-%d %s",
				now.Year(), now.Month(), now.Day(), priceTime), ig.TimeZoneLightStreamer)
			if err != nil {
				fmt.Printf("parsing time failed: %v time=%q\n", err, priceTime)
				continue
			}
		}
		tableIndex := priceParts[0]
		priceBid, _ := strconv.ParseFloat(priceParts[2], 64)
		priceAsk, _ := strconv.ParseFloat(priceParts[3], 64)

		epic, found := epicIndex[tableIndex]
		if !found {
			fmt.Printf("unknown epic %q\n", tableIndex)
			epic = epicNameUnknown
		}

		if epic != epicNameUnknown {
			var lastTick, found = lastTicks[epic]
			if found {
				if priceAsk == 0 {
					priceAsk = lastTick.Ask
				}
				if priceBid == 0 {
					priceBid = lastTick.Bid
				}
				if parsedTime.IsZero() {
					parsedTime = lastTick.Time
				}
			}
		}

		tick := LightStreamerTick{
			Epic: epic,
			Time: parsedTime,
			Bid:  priceBid,
			Ask:  priceAsk,
		}
		tickReceiver <- tick
		lastTicks[epic] = tick
	}
}
