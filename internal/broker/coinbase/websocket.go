package coinbase

import (
	ws "github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/pkg/tick"
	"time"
)

func (cb *Coinbase) getInstrumentMessages() []coinbasepro.MessageChannel {
	var instruments []coinbasepro.MessageChannel
	for _, name := range cb.instruments {
		instruments = append(instruments, coinbasepro.MessageChannel{
			Name:       "ticker",
			ProductIds: []string{name},
		})
	}
	return instruments
}

func (cb *Coinbase) setupConnection() (*ws.Conn, error) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		return nil, err
	}

	subscribe := coinbasepro.Message{
		Type: "subscribe",
		Channels: []coinbasepro.MessageChannel{
			{
				Name: "heartbeat",
				ProductIds: []string{
					cb.instruments[0],
				},
			},
		},
	}
	subscribe.Channels = append(subscribe.Channels, cb.getInstrumentMessages()...)

	if err := wsConn.WriteJSON(subscribe); err != nil {
		return nil, err
	}
	return wsConn, nil
}

func (cb *Coinbase) ListenToPriceFeed(tickChan chan tick.Tick) {
	defer close(tickChan)

	const defaultRetryDelay = time.Second * 5
	const maxRetryDelay = time.Hour

	var retryDelay = defaultRetryDelay
	var wsConn *ws.Conn
	var err error

	for {
		time.Sleep(retryDelay)

		wsConn, err = cb.setupConnection()
		if err != nil {
			retryDelay = retryDelay * 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
			log.WithError(err).Errorf("OpenLightStreamerSubscription() failed. Retry in %s", retryDelay)
			continue
		}

		log.Infof("Connected to websocket for instruments %+v", cb.instruments)
		retryDelay = defaultRetryDelay

		const messageType = "ticker"
		for true {
			message := coinbasepro.Message{}
			if err := wsConn.ReadJSON(&message); err != nil {
				println(err.Error())
				break
			}
			if message.Type != messageType {
				continue
			}

			log.Debugf("Tick: %+v", message)

			bid, err := decimal.NewFromString(message.BestBid)
			if err != nil {
				continue
			}
			ask, err := decimal.NewFromString(message.BestAsk)
			if err != nil {
				continue
			}

			tickData := tick.New(message.ProductID, message.Time.Time(), bid, ask)
			if err := tickData.Validate(); err != nil {
				log.WithError(err).Warnf("Invalid tick: %s", tickData.String())
				continue
			}

			tickChan <- tickData
		}
	}
}
