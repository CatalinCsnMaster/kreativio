FROM golang:1.14 AS build

COPY . /go/src/github.com/moapis/shop
WORKDIR /go/src/github.com/moapis/shop/cmd/server
RUN go build

FROM debian:buster-slim

RUN apt-get update && \
    apt-get install -y ca-certificates

COPY --from=build /go/src/github.com/moapis/shop/cmd/server/server /server
COPY cmd/server/templates /templates

ENTRYPOINT [ "/server" ]