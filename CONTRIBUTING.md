# Contributing

Thanks for your interest in improving `at`. This guide covers the local workflow.

## Prerequisites

- Go 1.23 or newer
- (Optional) [golangci-lint](https://golangci-lint.run/) v2 for linting
- (Optional) Docker, to build the container image

## Common tasks

A [`Makefile`](Makefile) wraps the everyday commands:

```sh
make build   # compile all packages
make test    # go test -race -cover ./...
make vet     # go vet ./...
make fmt     # gofmt -w on all Go sources
make lint    # golangci-lint run
make tidy    # go mod tidy && go mod vendor
make docker  # build the Docker image
```

Before opening a pull request, make sure `make build`, `make test`, and `make lint` all pass. CI runs the same checks.

## Project layout

- `cmd/` — entry points (backtesting, histdata import, IG and Coinbase runners)
- `internal/broker` — the `Broker` interface and implementations; `internal/paperwallet` simulates fills
- `internal/strategy` — the `Strategy` interface and example strategies
- `internal/trader` — the orchestrator connecting brokers and strategies
- `pkg/indicator` — the `Indicator` interface and implementations
- `pkg/{ohlc,tick,eo,chart,...}` — supporting libraries

## Adding a strategy

1. Create a package under `internal/strategy/<name>` with a type that implements the [`Strategy`](internal/strategy/strategy.go) interface. [`internal/strategy/rsi`](internal/strategy/rsi) is a good template.
2. Add a name constant to [`internal/strategy/strategy.go`](internal/strategy/strategy.go).
3. Wire the new name into the `switch` in [`cmd/backtesting/main.go`](cmd/backtesting/main.go) so it can be selected via `STRATEGY`.
4. Add a `_test.go` file covering the decision logic.

## Adding an indicator

Implement the [`Indicator`](pkg/indicator/indicator.go) interface in a new package under `pkg/indicator/<name>`, and add tests. Existing indicators wrap [go-talib](https://github.com/markcheno/go-talib) and are good references.

## Style

- Log with the standard library `log/slog`, not third-party loggers.
- Keep library code free of `log.Fatal`/`os.Exit`; return errors instead, and reserve `panic` for unreachable invariants.
- Run `make fmt` before committing.
