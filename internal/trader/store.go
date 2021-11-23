package trader

//func (tr *Trader) fillTrackerWithStoredOHLCs(count int) {
//	var ohlcs []ohlc.OHLC
//	var err = tr.gormDB.
//		Limit(count).
//		Order("start ASC").
//		Where("instrument = ? AND duration = ?", tr.Instrument, ohlcPeriod).
//		Find(&ohlcs).Error
//	if err != nil {
//		log.WithError(err).Fatal("failed to fetch OHLCs from DB")
//	}
//	for _, o := range ohlcs {
//		o.ForceClose()
//		tr.volaTracker.AddOHLC(&o)
//		tr.perfTracker.AddOHLC(&o)
//
//		// SMA
//		close, _ := o.Close.Float64()
//		tr.sma.Insert(close)
//	}
//	tr.clog.Infof("fillTrackerWithStoredOHLCs: %d OHLCs imported", len(ohlcs))
//}
//
//func (tr *Trader) fillTrackerWithStoredDailyOHLCs(count int) {
//	var ohlcs []ohlc.OHLC
//	var err = tr.gormDB.
//		Limit(count).
//		Order("start ASC").
//		Where("instrument = ? AND duration = ?", tr.Instrument, eodPeriod).
//		Find(&ohlcs).Error
//	if err != nil {
//		log.WithError(err).Fatal("failed to fetch OHLCs from DB")
//	}
//	for _, o := range ohlcs {
//		o.ForceClose()
//		tr.dayTracker.AddOHLC(&o)
//		tr.dayVolaTracker.AddOHLC(&o)
//	}
//	tr.clog.Infof("fillTrackerWithStoredDailyOHLCs: %d OHLCs imported", len(ohlcs))
//}
//
//func (tr *Trader) setTodayOHLC() {
//	var ticks []tick.Tick
//	var err = tr.gormDB.
//		Order("datetime ASC").
//		Where("instrument = ? AND datetime >= ?", tr.Instrument, time.Now().UTC()).
//		Find(&ticks).Error
//	if err != nil {
//		log.WithError(err).Fatal("failed to fetch ticks from DB")
//	}
//	if tr.today != nil {
//		log.Fatal("setTodayOHLC: tr.today must be null")
//	}
//	if len(ticks) == 0 {
//		log.Warn("setTodayOHLC no ticks found")
//		return
//	}
//	for _, t := range ticks {
//		if tr.today == nil {
//			tr.today = ohlc.New(tr.Instrument, t.Datetime, eodPeriod)
//		}
//		tr.today.NewPrice(t.Bid, t.Datetime)
//	}
//	tr.clog.WithFields(log.Fields{
//		"FirstTick": tr.today.Start,
//		"LastTick":  tr.today.End,
//	}).Infof("setTodayOHLC: %d ticks imported", len(ticks))
//}
