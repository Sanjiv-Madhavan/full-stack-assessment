FROM golang:1.24.2-alpine AS builder

RUN apk update
RUN apk add git ca-certificates curl

ADD . /go/src/github.com/Sanjiv-Madhavan/full-stack-assessment
WORKDIR /go/src/github.com/Sanjiv-Madhavan/full-stack-assessment

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o full-stack-backend cmd/main.go


RUN mkdir -p /target_root/etc/
RUN cp /etc/passwd /target_root/etc/
RUN cp full-stack-backend /target_root/

FROM alpine:latest

RUN apk add --no-cache bash

COPY --from=builder /target_root /
USER 1000
WORKDIR /
ENTRYPOINT ["/full-stack-backend"]