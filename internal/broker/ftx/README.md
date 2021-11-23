# FTX Broker

Implement the FTX.com trading API as broker backend. It fetches ticks from FTX.com but `paperwallet` for trading. So no trades are excuted via the FTX.com API right now as they don't offer sandboxes accounts for testing/demo purpose.

If you want to enable trading with real money feel free to replace the paperwallet method calls by actual API calls.