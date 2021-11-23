package trader

//func Test_distanceSMAOK(t *testing.T) {
//	tr := New("", false, false, &backtest.Backtest{}, "", getDBHandle(t))
//	bid := decimal.NewFromFloat(10)
//	ask := decimal.NewFromFloat(10.01)
//	currentTick := tick.New("", time.Now(), bid, ask)
//	assert.ErrorIncludesMessage(t, "not enough data", tr.distanceSMAOK(currentTick, broker.BuyDirectionLong))
//
//	for i := 0; i < 200; i++ {
//		tr.sma.Insert(10)
//	}
//	assert.NoError(t, tr.distanceSMAOK(currentTick, broker.BuyDirectionLong))
//
//	bid = decimal.NewFromFloat(15.1)
//	ask = decimal.NewFromFloat(15.1)
//	currentTick = tick.New("", time.Now(), bid, ask)
//	assert.ErrorIncludesMessage(t, "gap to SMA is too big", tr.distanceSMAOK(currentTick, broker.BuyDirectionLong))
//
//	bid = decimal.NewFromFloat(8)
//	ask = decimal.NewFromFloat(8.01)
//	currentTick = tick.New("", time.Now(), bid, ask)
//	assert.ErrorIncludesMessage(t, "gap to SMA is too big", tr.distanceSMAOK(currentTick, broker.BuyDirectionShort))
//}
