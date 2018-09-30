FROM golang:alpine AS builder
COPY . /go/src/github.com/dafanasev/dbmigrate
WORKDIR /go/src/github.com/dafanasev/dbmigrate
RUN apk add --no-cache git gcc libc-dev \
    && go get ./... \
    && go build -o=cmd/dbmigrate ./cmd \
    && apk del git gcc libc-dev

FROM alpine
MAINTAINER Dmitrii Afanasev <dimarzio1986@gmail.com>
COPY --from=builder /go/src/github.com/dafanasev/dbmigrate/cmd/dbmigrate /
CMD ["/dbmigrate"]