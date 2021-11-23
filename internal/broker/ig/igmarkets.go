package ig

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/pkg/tick"
	"github.com/sklinkert/igmarkets"
	"strings"
	"sync"
	"time"
)

const openPositionsRefreshInterval = time.Second * 5
const closedPositionsRefreshInterval = time.Second * 5
const maxTokenRefreshFailures = 5

type Broker struct {
	igHandle                   *igmarkets.IGMarkets
	watchlistID                string
	instrument                 string
	loc                        *time.Location
	cachedOpenPositions        []broker.Position
	cachedClosedPositions      []broker.Position
	openPositionsLastChecked   time.Time
	closedPositionsLastChecked time.Time
	tokenRefreshFailures       int
	sync.RWMutex
}

// New creates new broker instance and does the API login
func New(instrument string, apiURL, apiKey, accountID, identifier, password string) (*Broker, error) {
	var igHandle = igmarkets.New(apiURL, apiKey, accountID, identifier, password)
	if err := igHandle.Login(context.Background()); err != nil {
		return nil, err
	}

	locBerlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.WithError(err).Fatal("cannot load time zone")
	}

	b := &Broker{
		igHandle:    igHandle,
		watchlistID: "_at",
		instrument:  instrument,
		loc:         locBerlin,
	}

	const watchlistName = "_at"
	watchlistID, err := b.getWatchListID(watchlistName)
	if err != nil {
		return nil, err
	}
	b.watchlistID = watchlistID

	go b.refreshIGToken()

	return b, nil
}

func (b *Broker) getWatchListID(watchlistName string) (watchlistID string, err error) {
	b.RLock()
	defer b.RUnlock()

	watchlists, err := b.igHandle.GetAllWatchlists(context.Background())
	if err != nil {
		return "", err
	}

	for _, watchlist := range *watchlists {
		if watchlist.Name == watchlistName {
			watchlistID = watchlist.ID
			break
		}
	}
	return
}

// refreshIGHandle - new handle with fresh token
func (b *Broker) refreshIGToken() {
	for {
		time.Sleep(time.Second * 50)

		now := time.Now().In(b.loc)
		if !areMarketsOpen(now) {
			continue
		}
		log.Debug("Refreshing IG token")

		b.Lock()
		if err := b.igHandle.Login(context.Background()); err != nil {
			log.WithError(err).WithFields(log.Fields{"Failures": b.tokenRefreshFailures}).Error("Failed to refresh IG token")
			b.tokenRefreshFailures++
			if b.tokenRefreshFailures > maxTokenRefreshFailures {
				log.Fatal("too many failures for IG token refresh")
			}
		}
		b.tokenRefreshFailures = 0
		b.Unlock()
	}
}

func toDirection(direction broker.BuyDirection) string {
	switch direction {
	case broker.BuyDirectionLong:
		return "BUY"
	case broker.BuyDirectionShort:
		return "SELL"
	default:
		return ""
	}
}

func (b *Broker) Buy(order broker.Order) (broker.Position, error) {
	b.Lock()
	defer b.Unlock()

	igDirection := toDirection(order.Direction)
	clog := log.WithFields(log.Fields{
		"Instrument":    order.Instrument,
		"Size":          order.Size,
		"TargetPrice":   order.TargetPrice.String(),
		"StopLossPrice": order.StopLossPrice.String(),
		"CurrencyCode":  order.CurrencyCode,
		"Direction":     igDirection,
	})

	targetStr := ""
	if order.HasTargetPrice() {
		targetStr = order.TargetPrice.String()
	}

	igOrder := igmarkets.OTCOrderRequest{
		Epic:         order.Instrument,
		OrderType:    "MARKET",
		CurrencyCode: order.CurrencyCode,
		Direction:    igDirection,
		Size:         order.Size,
		Expiry:       "-",
		LimitLevel:   targetStr,
		StopLevel:    order.StopLossPrice.String(),
		//GuaranteedStop: true,
		ForceOpen: true,
	}

	if !order.TrailingStopDistanceInPips.IsZero() {
		igOrder.StopDistance = order.TrailingStopDistanceInPips.String()
		igOrder.TrailingStopIncrement = order.TrailingStopIncrementSizeInPips.String()
		igOrder.StopLevel = ""
		igOrder.TrailingStop = true
	}

	clog.Debugf("New order: %v", order)

	now := time.Now()
	dealRef, err := b.igHandle.PlaceOTCOrder(context.Background(), igOrder)
	if err != nil {
		clog.WithError(err).Error("Unable to place order")
		return broker.Position{}, err
	}

	clog = clog.WithFields(log.Fields{"Reference": dealRef})
	clog.Infof("New order placed successfully. Took %s", time.Since(now))

	time.Sleep(1 * time.Second)

	var confirmation *igmarkets.OTCDealConfirmation
	attempts := 0
	for {
		attempts++
		confirmation, err = b.igHandle.GetDealConfirmation(context.Background(), dealRef.DealReference)
		if err != nil {
			clog.WithError(err).Error("Error while getting deal confirmation for")
			if attempts >= 100 {
				clog.Error("too many failures for b.igHandle.GetDealConfirmation(dealRef)", dealRef)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	if confirmation.Status != "OPEN" {
		clog.WithFields(log.Fields{"Status": confirmation.Status}).Errorf("Unexpected order status: %+v", confirmation)
		return broker.Position{}, fmt.Errorf("unexpected order status %q", confirmation.Status)
	}

	b.openPositionsLastChecked = time.Time{} // invalidate cache

	return broker.Position{
		Reference:     toInternalReference(confirmation.AffectedDeals[0].DealID, confirmation.DealReference),
		Instrument:    confirmation.Epic,
		BuyPrice:      decimal.NewFromFloat(confirmation.Level),
		BuyTime:       time.Now(),
		BuyDirection:  order.Direction,
		TargetPrice:   decimal.NewFromFloat(confirmation.LimitLevel),
		StopLossPrice: decimal.NewFromFloat(confirmation.StopLevel),
		Size:          order.Size,
	}, nil
}

func toInternalReference(dealID, dealReference string) string {
	return fmt.Sprintf("%s:%s", dealID, dealReference)
}

func fromInternalReference(position broker.Position) (dealID, dealReference string) {
	s := strings.Split(position.Reference, ":")
	if len(s) != 2 {
		log.WithFields(log.Fields{
			"position.Reference": position.Reference,
		}).Fatalf("Malformed internal reference")
	}
	return s[0], s[1]
}

func (b *Broker) Sell(position broker.Position) error {
	b.Lock()
	defer b.Unlock()

	clog := log.WithFields(log.Fields{
		"Reference": position.Reference,
	})
	clog.Info("Sell called")

	var closeDirection broker.BuyDirection
	switch position.BuyDirection {
	case broker.BuyDirectionLong:
		closeDirection = broker.BuyDirectionShort
	case broker.BuyDirectionShort:
		closeDirection = broker.BuyDirectionLong
	default:
		return fmt.Errorf("unexpected direction of position +%v", position)
	}

	dealID, _ := fromInternalReference(position)
	closeReq := igmarkets.OTCPositionCloseRequest{
		DealID:    dealID,
		OrderType: "MARKET",
		Direction: toDirection(closeDirection),
		Size:      position.Size,
		Expiry:    "-",
	}

	dealRef, err := b.igHandle.CloseOTCPosition(context.Background(), closeReq)
	if err != nil {
		clog.WithError(err).Error("Unable to close position")
		return err
	}

	confirmation, err := b.igHandle.GetDealConfirmation(context.Background(), dealRef.DealReference)
	if err != nil {
		clog.WithError(err).Error("Cannot get deal confirmation")
		return err
	}

	clog.WithFields(log.Fields{
		"DealRef":    dealRef.DealReference,
		"DealStatus": confirmation.DealStatus,
		"Profit":     confirmation.Profit,
		"Currency":   confirmation.ProfitCurrency,
		"Reason":     confirmation.Reason,
		"Level":      confirmation.Level,
	}).Info("Deal confirmed")

	// Invalidate cache
	b.openPositionsLastChecked = time.Time{}
	b.closedPositionsLastChecked = time.Time{}

	return nil
}

func (b *Broker) GetOpenPosition(positionRef string) (position broker.Position, err error) {
	positions, err := b.GetOpenPositions()
	if err != nil {
		return broker.Position{}, err
	}
	for _, position := range positions {
		_, dealReference := fromInternalReference(position)
		if dealReference == positionRef {
			return position, nil
		}
	}
	return broker.Position{}, broker.ErrPositionNotFound
}

// GetOpenPositions returns all open positions for given instrument name
func (b *Broker) GetOpenPositionsByInstrument(instrument string) ([]broker.Position, error) {
	positions, err := b.GetOpenPositions()
	if err != nil {
		return []broker.Position{}, err
	}

	var foundPositions []broker.Position
	for _, position := range positions {
		if position.Instrument == instrument {
			foundPositions = append(foundPositions, position)
		}
	}
	return foundPositions, nil
}

func (b *Broker) GetOpenPositions() ([]broker.Position, error) {
	b.RLock()
	defer b.RUnlock()

	if time.Since(b.openPositionsLastChecked).Seconds() < openPositionsRefreshInterval.Seconds() {
		return b.cachedOpenPositions, nil
	}

	var positions []broker.Position
	posResponse, err := b.igHandle.GetPositions(context.Background())
	if err != nil {
		return positions, err
	}

	for _, positionData := range posResponse.Positions {
		position := positionData.Position

		var direction broker.BuyDirection
		switch position.Direction {
		case "BUY":
			direction = broker.BuyDirectionLong
		case "SELL":
			direction = broker.BuyDirectionShort
		default:
			return positions, fmt.Errorf("unexpected buy direction: %+v", position)
		}

		buyTime, err := time.Parse("2006-01-02T15:04:05", position.CreatedDateUTC)
		if err != nil {
			return positions, fmt.Errorf("cannot parse CreatedDateUTC: %+v", position)
		}

		positions = append(positions, broker.Position{
			Reference:     toInternalReference(position.DealID, position.DealReference),
			Instrument:    positionData.MarketData.Epic,
			BuyPrice:      decimal.NewFromFloat(position.Level),
			BuyTime:       buyTime,
			BuyDirection:  direction,
			TargetPrice:   decimal.NewFromFloat(position.LimitLevel),
			StopLossPrice: decimal.NewFromFloat(position.StopLevel),
		})
	}

	b.RUnlock()
	b.Lock()
	b.cachedOpenPositions = positions
	b.openPositionsLastChecked = time.Now()
	b.Unlock()
	b.RLock()
	return positions, nil
}

func (b *Broker) GetClosedPositions() ([]broker.Position, error) {
	b.RLock()
	defer b.RUnlock()

	if time.Since(b.closedPositionsLastChecked).Seconds() < closedPositionsRefreshInterval.Seconds() {
		return b.cachedClosedPositions, nil
	}

	var positions []broker.Position
	const transactionType = "ALL_DEAL"
	var last24h = time.Now().Add(-(time.Hour * 24))
	transResponse, err := b.igHandle.GetTransactions(context.Background(), transactionType, last24h)
	if err != nil {
		return positions, err
	}

	for _, transaction := range transResponse.Transactions {
		buyTime, err := time.Parse("2006-01-02T15:04:05", transaction.OpenDateUtc)
		if err != nil {
			return positions, fmt.Errorf("cannot parse OpenDateUtc: %+v", transaction)
		}
		sellTime, err := time.Parse("2006-01-02T15:04:05", transaction.DateUTC)
		if err != nil {
			return positions, fmt.Errorf("cannot parse OpenDateUtc: %+v", transaction)
		}

		buyPrice, _ := decimal.NewFromString(transaction.OpenLevel)
		sellPrice, _ := decimal.NewFromString(transaction.CloseLevel)

		positions = append(positions, broker.Position{
			Reference:  transaction.Reference,
			Instrument: transaction.InstrumentName,
			BuyPrice:   buyPrice,
			BuyTime:    buyTime,
			SellPrice:  sellPrice,
			SellTime:   sellTime,
		})
	}

	b.RUnlock()
	b.Lock()
	b.cachedClosedPositions = positions
	b.closedPositionsLastChecked = time.Now()
	b.Unlock()
	b.RLock()

	return positions, nil
}

func (b *Broker) ListenToPriceFeed(tickChan chan tick.Tick) {
	const maxTimeDelta = time.Minute * 5

	for {
		time.Sleep(time.Second * 5)

		if !areMarketsOpen(time.Now()) {
			continue
		}

		var lightStreamReceiver = make(chan igmarkets.LightStreamerTick)
		b.Lock()
		err := b.igHandle.OpenLightStreamerSubscription(context.Background(), []string{b.instrument}, lightStreamReceiver)
		b.Unlock()
		if err != nil {
			log.WithError(err).Error("OpenLightStreamerSubscription() failed")
			continue
		}

		//var timeOfLastPriceUpdate time.Time
		for market := range lightStreamReceiver {
			if market.Epic != b.instrument {
				continue
			}
			log.Debugf("Tick: %+v", market)

			// Price is too old
			if time.Since(market.Time) > maxTimeDelta {
				log.Errorf("Time delta too big: nowUTC %s received %s epic %s",
					time.Now().UTC().String(), market.Time.String(), market.Epic)
				continue
			}

			// Price is from future
			nowUTC := time.Now().UTC()
			marketTimeUTC := market.Time.UTC()
			if marketTimeUTC.After(nowUTC.Add(maxTimeDelta)) {
				log.Errorf("price date %s is too far in future, skipping (nowUTC: %s)",
					marketTimeUTC, nowUTC)
				continue
			}

			bid := decimal.NewFromFloat(market.Bid)
			ask := decimal.NewFromFloat(market.Ask)
			tickData := tick.New(market.Epic, market.Time, bid, ask)

			//if tickData.Datetime == timeOfLastPriceUpdate {
			//	log.WithFields(log.Fields{
			//		"price.Datetime":        tickData.Datetime,
			//		"timeOfLastPriceUpdate": timeOfLastPriceUpdate,
			//		"instrument":            b.instrument,
			//	}).Warn("Received price is stale")
			//	continue
			//}
			//timeOfLastPriceUpdate = tickData.Datetime

			tickChan <- tickData
		}
	}
}

func areMarketsOpen(now time.Time) bool {
	now = now.UTC()
	if now.Weekday() == time.Friday && now.Hour() >= 22 {
		return false
	}
	if now.Weekday() == time.Saturday {
		return false
	}
	if now.Weekday() == time.Sunday && now.Hour() < 22 {
		return false
	}
	return true
}
