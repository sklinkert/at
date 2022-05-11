package trader

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/at/internal/broker"
	"github.com/sklinkert/at/internal/strategy"
	"github.com/sklinkert/at/pkg/ohlc"
	"github.com/sklinkert/at/pkg/tick"
	"gorm.io/gorm"
	"sort"
	"sync"
	"time"
)

type Trader struct {
	ctx                         context.Context
	StartTime                   time.Time
	Instrument                  string
	TickChan                    chan tick.Tick
	running                     bool
	clog                        *log.Entry
	broker                      broker.Broker
	strategy                    strategy.Strategy
	persistTickData             bool
	persistCandleData           bool
	today                       *ohlc.OHLC
	maxConcurrentPositions      int
	MaxAggregatedDrawdownInPips decimal.Decimal
	reversedPerformanceInPips   map[ohlc.OHLC]float64
	gormDB                      *gorm.DB
	positionBuyTime             map[string]time.Time
	openCandles                 []*ohlc.OHLC // strategy's candle + today's candle
	closedCandles               []*ohlc.OHLC
	lastReceivedTick            *tick.Tick
	gitRev                      string
	candleSubscribers           []CandleSubscriber
	positionSubscribers         []PositionSubscriber
	orderSubscribers            []OrderSubscriber
	closedPositionReferences    map[string]bool
	currencyCode                string
	sync.Mutex
}

type PositionSubscriber interface {
	OnPosition(position broker.Position)
}

type CandleSubscriber interface {
	OnCandle(candle ohlc.OHLC)
}

type OrderSubscriber interface {
	OnOrder(order broker.Order)
}

type Option func(*Trader)

func WithBroker(broker broker.Broker) Option {
	return func(trader *Trader) {
		trader.broker = broker
	}
}

func WithCandleSubscription(subscriber CandleSubscriber) Option {
	return func(trader *Trader) {
		trader.candleSubscribers = append(trader.candleSubscribers, subscriber)
	}
}

func WithPositionSubscription(subscriber PositionSubscriber) Option {
	return func(trader *Trader) {
		trader.positionSubscribers = append(trader.positionSubscribers, subscriber)
	}
}

//func WithOrderSubscription(subscriber OrderSubscriber) Option {
//	return func(trader *Trader) {
//		trader.orderSubscribers = append(trader.orderSubscribers, subscriber)
//	}
//}
//
//func WithGatherPerformanceData() Option {
//	return func(trader *Trader) {
//		trader.gatherPerformanceData = true
//	}
//}

func WithPersistTickData(persist bool) Option {
	return func(trader *Trader) {
		trader.persistTickData = persist
	}
}

func WithPersistCandleData(persist bool) Option {
	return func(trader *Trader) {
		trader.persistCandleData = persist
	}
}

func WithStrategy(strategy strategy.Strategy) Option {
	return func(trader *Trader) {
		trader.strategy = strategy
	}
}

func WithCurrencyCode(currencyCode string) Option {
	return func(trader *Trader) {
		trader.currencyCode = currencyCode
	}
}

func WithFeedStoredCandles(strategy strategy.Strategy) Option {
	return func(trader *Trader) {
		var limit = int(strategy.GetWarmUpCandleAmount())
		var candlePeriod = strategy.GetCandleDuration()

		log.Infof("Searching for warmup candles with period %s", candlePeriod)

		var candles ohlc.OHLCList
		if err := trader.gormDB.Limit(limit).Order("\"end\" DESC").Where("instrument = ? AND duration = ?", trader.Instrument, candlePeriod).Find(&candles).Error; err != nil {
			log.WithError(err).Fatal("fetching stored candles failed")
		}
		sort.Sort(candles)

		log.Infof("WithFeedStoredCandles: Sending %d candles to strategy for warming up", len(candles))
		for _, candle := range candles {
			candle.ForceClose()
			strategy.OnWarmUpCandle(&candle)
		}
	}
}

func New(ctx context.Context, instrument, gitRev string, db *gorm.DB, options ...Option) *Trader {
	var clog = log.WithFields(log.Fields{
		"INSTRUMENT": instrument,
		"GIT_REV":    gitRev,
	})
	tr := &Trader{
		ctx:                       ctx,
		Instrument:                instrument,
		StartTime:                 time.Now(),
		clog:                      clog,
		TickChan:                  make(chan tick.Tick),
		reversedPerformanceInPips: make(map[ohlc.OHLC]float64),
		positionBuyTime:           make(map[string]time.Time),
		closedPositionReferences:  make(map[string]bool),
		gitRev:                    gitRev,
		gormDB:                    db,
		currencyCode:              "USD", // default
	}

	for _, option := range options {
		option(tr)
	}

	if tr.gormDB == nil {
		if tr.persistTickData || tr.persistCandleData {
			log.Fatalf("Persistence of ticks or/and candles requested but no DB given!")
		}
	} else {
		if err := tr.gormDB.AutoMigrate(&ohlc.OHLC{}, &PerformanceRecord{}, &tick.Tick{}, &broker.Position{}); err != nil {
			log.WithError(err).Fatal("db.AutoMigrate() failed")
		}
	}

	return tr
}

func (tr *Trader) ID() string {
	return fmt.Sprintf("rev_%s_strategy_%s", tr.gitRev, tr.strategy.Name())
}

func (tr *Trader) Start() error {
	if tr.running {
		return errors.New("already running")
	}
	tr.running = true
	tr.clog.Info("Starting trader")

	go tr.receiveTicks()
	tr.broker.ListenToPriceFeed(tr.TickChan)

	return nil
}

func (tr *Trader) Stop() error {
	if !tr.running {
		return errors.New("already stopped")
	}
	tr.clog.Info("Stopping trader")

	tr.Lock()
	defer tr.Unlock()

	close(tr.TickChan)
	tr.running = false
	tr.printPositionPerformanceByNotes()

	return nil
}

func (tr *Trader) GetClosedPositions() ([]broker.Position, error) {
	positions, err := tr.broker.GetClosedPositions()
	if err != nil {
		return []broker.Position{}, err
	}
	for i := range positions {
		positions[i].CandleBuyTime = tr.positionBuyTime[positions[i].Reference]
	}
	return positions, nil
}

func (tr *Trader) processTodayCandle(currentTick tick.Tick) {
	const eodPeriod = time.Hour * 24 * 1 // 1d

	if tr.today == nil || tr.today.Start.Day() != currentTick.Datetime.Day() {
		if tr.today != nil {
			tr.today.ForceClose()
		}
		tr.today = ohlc.New(tr.Instrument, currentTick.Datetime, eodPeriod, false)
	}
	tr.today.NewPrice(currentTick.Bid, currentTick.Datetime)
}

func (tr *Trader) getOpenPositions() ([]broker.Position, error) {
	positions, err := tr.broker.GetOpenPositions()
	if err != nil {
		return []broker.Position{}, err
	}
	for i := range positions {
		positions[i].CandleBuyTime = tr.positionBuyTime[positions[i].Reference]
	}
	return positions, nil
}

func (tr *Trader) persistTick(t tick.Tick) {
	if err := tr.gormDB.Create(&t).Error; err != nil {
		log.WithError(err).Errorf("Cannot persist tick: %+v", t)
	}
}

func (tr *Trader) receiveTicks() {
	var locBerlin, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.WithError(err).Fatal("Unable to load Europe/Berlin timezone")
	}

	for currentTick := range tr.TickChan {
		if tr.persistTickData {
			go tr.persistTick(currentTick)
		}
		currentTick.Datetime = currentTick.Datetime.In(locBerlin)

		if err := currentTick.Validate(); err != nil {
			tr.clog.WithError(err).Debugf("Invalid tick data received: %+v", currentTick)
			continue
		}

		//tr.clog.Debugf("New tick received %s", currentTick.String())
		tr.Lock()
		tr.processTodayCandle(currentTick)
		tr.processTick(currentTick)
		tr.Unlock()
	}
}

func (tr *Trader) processTick(currentTick tick.Tick) {
	var closedCandles = tr.processTickByOpenCandles(currentTick)
	for _, closedCandle := range closedCandles {
		tr.processClosedCandle(closedCandle, currentTick)
	}
}

func (tr *Trader) processClosedCandle(closedCandle *ohlc.OHLC, currentTick tick.Tick) {
	tr.clog.Debugf("Processing closed candle: %s", closedCandle)

	if !closedCandle.HasPriceData() {
		tr.clog.Debugf("processClosedCandle: candle has missing price data, cannot process further: %s", closedCandle)
		return
	}

	openOrders, err := tr.broker.GetOpenOrders()
	if err != nil {
		tr.clog.WithError(err).Error("Cannot get open orders")
		return
	}

	if tr.strategy.GetCandleDuration() != closedCandle.Duration {
		return
	}
	openPositions, err := tr.getOpenPositions()
	if err != nil {
		tr.clog.WithError(err).Error("Cannot get open positions")
		return
	}
	closedPositions, err := tr.GetClosedPositions()
	if err != nil {
		tr.clog.WithError(err).Error("Cannot get closed positions")
		return
	}
	tr.detectClosedPositions(closedPositions)
	tr.processOpenPositions(closedCandle, openPositions)

	toOpen, toCloseOrderIDs, toClosePositons := tr.strategy.OnCandle(closedCandle, tr.closedCandles, currentTick,
		openOrders, openPositions, closedPositions)
	tr.processClosableOrders(toCloseOrderIDs)
	tr.processClosablePositions(toClosePositons)
	tr.processOrders(closedCandle, currentTick, toOpen)

	for _, subscriber := range tr.candleSubscribers {
		subscriber.OnCandle(*closedCandle)
	}
}

func (tr *Trader) processClosableOrders(orderIDs []string) {
	for _, orderID := range orderIDs {
		if err := tr.broker.CancelOrder(orderID); err != nil {
			tr.clog.WithError(err).WithFields(log.Fields{"OrderID": orderID}).Error("Unable to cancel order")
		}
	}
}

func (tr *Trader) processOpenPositions(candle *ohlc.OHLC, openPositions []broker.Position) {
	for _, openPosition := range openPositions {
		_, exists := tr.positionBuyTime[openPosition.Reference]
		if !exists {
			tr.positionBuyTime[openPosition.Reference] = candle.Start
		}
	}
}

func (tr *Trader) processClosablePositions(toClose []broker.Position) {
	for _, position := range toClose {
		if err := tr.broker.Sell(position); err != nil {
			tr.clog.WithError(err).WithFields(log.Fields{"Reference": position.Reference}).Error("Unable to sell position")
		}
	}
}

// processOrders - Execute order and open new positions
func (tr *Trader) processOrders(candle *ohlc.OHLC, currentTick tick.Tick, toOpen []broker.Order) {
	for _, order := range toOpen {
		order.CurrencyCode = tr.currencyCode

		_, err := tr.broker.Buy(order)
		if err != nil {
			tr.clog.WithError(err).Errorf("Unable to open position: %+v", order)
			continue
		}

		tr.clog.Infof("Got new order: %s", order.String())

		for _, subscriber := range tr.orderSubscribers {
			subscriber.OnOrder(order)
		}
	}
}

func (tr *Trader) processTickByOpenCandles(currentTick tick.Tick) (closedCandles []*ohlc.OHLC) {
	var stillOpenCandles []*ohlc.OHLC

	defer func() {
		lastReceivedTick := currentTick
		tr.lastReceivedTick = &lastReceivedTick
		tr.openCandles = stillOpenCandles
	}()

	if len(tr.openCandles) == 0 {
		candle := ohlc.New(tr.Instrument, currentTick.Datetime, tr.strategy.GetCandleDuration(), true)
		tr.openCandles = append(tr.openCandles, candle)

		if tr.persistCandleData && tr.strategy.GetCandleDuration() != time.Hour*24 {
			candle := ohlc.New(tr.Instrument, currentTick.Datetime, time.Hour*24, true)
			tr.openCandles = append(tr.openCandles, candle)
		}
	}

	for _, candle := range tr.openCandles {
		switch candle.Duration {
		case time.Hour * 24:
			if tr.lastReceivedTick != nil && tr.lastReceivedTick.Datetime.Day() != currentTick.Datetime.Day() {
				candle.ForceClose()
			}
		case time.Hour:
			if tr.lastReceivedTick != nil && tr.lastReceivedTick.Datetime.Hour() != currentTick.Datetime.Hour() {
				candle.ForceClose()
			}
		}

		isOpen := candle.NewPrice(currentTick.Price(), currentTick.Datetime)
		if isOpen {
			stillOpenCandles = append(stillOpenCandles, candle)
			continue
		}

		newCandle := tr.closeCandle(currentTick, candle)
		stillOpenCandles = append(stillOpenCandles, newCandle)
		closedCandles = append(closedCandles, candle)
	}
	return
}

func (tr *Trader) closeCandle(tick tick.Tick, candle *ohlc.OHLC) (newCandle *ohlc.OHLC) {
	tr.closedCandles = append(tr.closedCandles, candle)

	var candlesToKeep = 100
	if len(tr.closedCandles) > candlesToKeep {
		tr.closedCandles = tr.closedCandles[len(tr.closedCandles)-candlesToKeep:]
	}

	if tr.gormDB != nil && tr.persistCandleData {
		go func() {
			if err := candle.Store(tr.gormDB); err != nil {
				tr.clog.WithError(err).Errorf("Failed to store OHLC: %+v", candle)
			}
		}()
	}

	// Replace closed OHLC from openOHLCs list
	openCandle := ohlc.New(candle.Instrument, tick.Datetime, candle.Duration, true)
	openCandle.NewPrice(tick.Price(), tick.Datetime)
	return openCandle
}

func (tr *Trader) detectClosedPositions(brokerClosedPositions []broker.Position) {
	for _, closedByBroker := range brokerClosedPositions {
		_, exists := tr.closedPositionReferences[closedByBroker.Reference]
		if !exists {
			tr.closePosition(closedByBroker)
			tr.closedPositionReferences[closedByBroker.Reference] = true
		}
	}
}

func (tr *Trader) closePosition(position broker.Position) {
	for _, subscriber := range tr.positionSubscribers {
		subscriber.OnPosition(position)
	}
}
