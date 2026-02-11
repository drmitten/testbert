FROM golang:1.25 AS builder

RUN mkdir -p /go/src/github.com/drmitten/testbert
WORKDIR /go/src/github.com/drmitten/testbert

COPY . /go/src/github.com/drmitten/testbert

RUN go build -o bin/testbert-server server/main.go

FROM alpine:latest

RUN set -eux; \
  apk --update upgrade && \
  apk --no-cache add tzdata gcompat && \
  rm -rf /var/cache/apk/*

ENV TZ=Etc/UTC

RUN mkdir -p /opt/app
WORKDIR /opt/app

COPY --from=builder --chmod=+x /go/src/github.com/drmitten/testbert/bin/testbert-server /opt/app

RUN chmod +x -R /opt/app/

EXPOSE 50013
