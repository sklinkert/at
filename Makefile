GO ?= go
GOLANGCI_LINT ?= golangci-lint
PACKAGES := ./...
GOFMT_DIRS := cmd internal pkg

.PHONY: build test vet fmt fmt-check lint tidy docker clean help

## build: compile all packages
build:
	$(GO) build $(PACKAGES)

## test: run tests with the race detector and coverage
test:
	$(GO) test -race -cover $(PACKAGES)

## vet: run go vet
vet:
	$(GO) vet $(PACKAGES)

## fmt: format all Go sources in place
fmt:
	gofmt -w $(GOFMT_DIRS)

## fmt-check: fail if any Go source is not gofmt-formatted
fmt-check:
	@unformatted=$$(gofmt -l $(GOFMT_DIRS)); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt-formatted:"; echo "$$unformatted"; exit 1; \
	fi

## lint: run golangci-lint
lint:
	$(GOLANGCI_LINT) run $(PACKAGES)

## tidy: tidy and re-vendor dependencies
tidy:
	$(GO) mod tidy
	$(GO) mod vendor

## docker: build the container image
docker:
	docker build -f docker/Dockerfile -t at .

## clean: remove build/backtest artifacts
clean:
	rm -rf results data

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
