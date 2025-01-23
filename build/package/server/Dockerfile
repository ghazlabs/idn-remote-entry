FROM golang:1.23.5-alpine3.21 AS base

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry

COPY go.mod go.sum ./
RUN go mod download -x

COPY ./internal ./internal
COPY ./cmd/server ./cmd/server

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry/cmd/server
RUN go build -o server

FROM alpine:3.21
RUN apk add ca-certificates tzdata

WORKDIR /app
COPY --from=base /go/src/github.com/ghazlabs/idn-remote-entry/cmd/server/server /app/server

ENTRYPOINT [ "./server" ]