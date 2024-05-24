FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /go-chat
RUN mkdir -p /go-chat/proto/

WORKDIR /go-chat

COPY ./proto/service.pb.go /go-chat/proto
COPY ./main.go /go-chat

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o go-chat .

CMD ./go-chat
