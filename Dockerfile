FROM golang:1.12

WORKDIR /go/src/github.com/jarm-proxy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/jarm-proxy github.com/jarm-proxy

###

FROM python:3.8-slim-buster

COPY --from=0 /usr/local/bin/jarm-proxy /usr/local/bin/jarm-proxy
