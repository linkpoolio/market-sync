FROM golang:1.13-alpine as builder

WORKDIR /go/src/github.com/linkpoolio/market-sync

RUN apk add --no-cache make curl git g++ gcc musl-dev linux-headers

ENV GO111MODULE=on
ADD . .
RUN go install
RUN go build

# Copy into a second stage container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/linkpoolio/market-sync/market-sync /usr/local/bin/

ENTRYPOINT ["market-sync"]