FROM golang:1.23-alpine AS build
ARG GITREV="unknown"

# gcc/musl-dev are required because the SQLite driver needs CGO.
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=1 GOOS=linux go build -mod=vendor \
    -ldflags="-X main.GitRev=${GITREV}" \
    -o /at-backtesting ./cmd/backtesting

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /at-backtesting /usr/local/bin/at-backtesting

ENTRYPOINT ["at-backtesting"]
