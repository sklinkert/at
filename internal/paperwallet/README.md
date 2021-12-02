# Paperwallet

A paperwallet can be used any broker implementation. It's useful for brokers that don't offer testing accounts for trading without real money. It keeps track of all trades (broker.Positions), trading fees, and the balance.

For example the broker `backtest` consists of a paperwallet and reading ticks from various sources.
