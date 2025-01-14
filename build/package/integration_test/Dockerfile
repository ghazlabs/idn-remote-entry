FROM golang:1.23.4-bookworm AS golang-build
FROM chromedp/headless-shell:133.0.6905.0

# Install Golang on chromedp image
COPY --from=golang-build /usr/local/go /usr/local/go/
ENV PATH="${PATH}:/usr/local/go/bin"

WORKDIR /go/src/github.com/ghazlabs/idn-remote-entry

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download -x