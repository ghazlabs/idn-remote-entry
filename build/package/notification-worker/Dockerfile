FROM golang:1.23.5-alpine3.21 AS base

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry

COPY go.mod go.sum ./
RUN go mod download -x

COPY ./internal ./internal
COPY ./cmd/notification-worker ./cmd/notification-worker

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry/cmd/notification-worker
RUN go build -o notification-worker

FROM alpine:3.21
RUN apk add ca-certificates tzdata

WORKDIR /app
COPY --from=base /go/src/github.com/ghazlabs/idn-remote-entry/cmd/notification-worker/notification-worker /app/notification-worker

ENTRYPOINT [ "./notification-worker" ]