FROM golang:1.13-alpine AS builder

RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/app
ADD . .

RUN dep ensure
RUN go build -o /go/src/app/slack-webm-sentinel



FROM jrottenberg/ffmpeg:4.1-alpine

RUN apk add --no-cache ca-certificates ffmpeg

WORKDIR /app
COPY --from=builder /go/src/app/slack-webm-sentinel /app/slack-webm-sentinel

ENTRYPOINT ["/app/slack-webm-sentinel"]
