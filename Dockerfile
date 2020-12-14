FROM golang:alpine as builder

ENV APP_USER dutybot
ENV APP_HOME /dutybot

RUN addgroup -S $APP_USER && adduser -S $APP_USER -G $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

RUN apk add build-base

WORKDIR $APP_HOME
USER $APP_USER

COPY go.mod go.sum ./
RUN go mod download

COPY ./ $APP_HOME
RUN go build -o dutybot.app ./cmd/dutybot/main.go

FROM alpine:latest

ENV APP_USER dutybot
ENV APP_HOME /dutybot

RUN addgroup -S $APP_USER && adduser -S $APP_USER -G $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME
WORKDIR $APP_HOME

COPY --chown=0:0 --from=builder $APP_HOME/dutybot.app $APP_HOME

EXPOSE 8443
USER $APP_USER
CMD ["./dutybot.app"]
