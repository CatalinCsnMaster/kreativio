FROM golang:latest AS build

ADD . /frontend

WORKDIR /frontend

RUN go build -v

###########################################################################
FROM debian:stretch-slim

COPY --from=build /frontend .

COPY kreativio kreativio

ENV FRONTEND_PORT=80

EXPOSE 80

CMD  ./frontend

