FROM golang:1.18.1-alpine
MAINTAINER Simon Klinkert
ARG GITREV="unknown"

RUN mkdir -p /go/src/github.com/sklinkert/at
WORKDIR /go/src/github.com/sklinkert/at

COPY pkg ./pkg
COPY cmd ./cmd
COPY vendor ./vendor
COPY internal ./internal
COPY go.mod ./go.mod
COPY go.sum ./go.sum

## cmd/at-demo
WORKDIR /go/src/github.com/sklinkert/at/cmd/at-demo
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s -X main.GitRev=${GITREV}" -o /at-demo .

## cmd/at-web
WORKDIR /go/src/github.com/sklinkert/at/cmd/at-web
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s -X main.GitRev=${GITREV}" -o /at-web .

FROM scratch

# SSL certs
COPY docker/ca-bundle.pem /etc/ssl/certs/ca-certificates.crt

# timezones
COPY docker/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

COPY --from=0 /at-demo /at-demo
COPY --from=0 /at-web /at-web
COPY --from=0 /go/src/github.com/sklinkert/at/internal/web/templates /internal/web/templates

ENTRYPOINT ["/at-web"]
