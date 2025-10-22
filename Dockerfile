FROM golang:1.23.4-alpine3.19 as builder

RUN apk update
RUN apk add git ca-certificates curl

ADD . /go/src/github.com/Sanjiv-Madhavan/service-broker
WORKDIR /go/src/github.com/Sanjiv-Madhavan/service-broker

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o service-broker main.go

RUN mkdir -p /target_root/etc/ssl/certs/ \
    && cp /etc/ssl/certs/ca-certificates.crt /target_root/etc/ssl/certs/ \
    && cp /etc/passwd /target_root/etc/ \
    && cp service-broker /target_root/ \
    && find /target_root/.

FROM alpine:latest

# Install bash as root before switching user
RUN apk add --no-cache bash

COPY --from=builder /target_root /
USER 1000
WORKDIR /
ENTRYPOINT ["/service-broker"]