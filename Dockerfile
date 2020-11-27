FROM golang:alpine as builder

ENV APP_USER dutybot
ENV APP_HOME /dutybot

RUN addgroup -S $APP_USER && adduser -S $APP_USER -G $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

RUN apk add build-base

WORKDIR $APP_HOME
USER $APP_USER
COPY ./ $APP_HOME

RUN go mod download
RUN go build -o dutybot.app ./cmd/dutybot/main.go

FROM alpine:latest

ENV APP_USER dutybot
ENV APP_HOME /dutybot

RUN addgroup -S $APP_USER && adduser -S $APP_USER -G $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME
RUN apk add sqlite
WORKDIR $APP_HOME

COPY --chown=0:0 --from=builder $APP_HOME/dutybot.app $APP_HOME

USER $APP_USER
CMD ["./dutybot.app"]
