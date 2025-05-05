FROM golang:1.23.5-bullseye AS base

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry

COPY go.mod go.sum ./
RUN go mod download -x

COPY ./internal ./internal
COPY ./cmd/crawler-worker ./cmd/crawler-worker

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry/cmd/crawler-worker
RUN go build -o crawler-worker

FROM debian:bullseye
RUN apt update

# need for go-rod
RUN apt install -y chromium

WORKDIR /app
COPY --from=base /go/src/github.com/ghazlabs/idn-remote-entry/cmd/crawler-worker/crawler-worker /app/crawler-worker

ENTRYPOINT [ "./crawler-worker" ]