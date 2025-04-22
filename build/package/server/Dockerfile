FROM golang:1.23.4-bookworm AS base

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry

COPY go.mod go.sum ./
RUN go mod download -x

COPY ./internal ./internal
COPY ./cmd/server ./cmd/server

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry/cmd/server
RUN go build -o server

FROM chromedp/headless-shell:133.0.6905.0
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

WORKDIR /app
COPY --from=base /go/src/github.com/ghazlabs/idn-remote-entry/cmd/server/server /app/server

ENTRYPOINT [ "./server" ]