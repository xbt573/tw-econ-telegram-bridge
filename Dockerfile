FROM golang:alpine as build

WORKDIR /app

RUN apk add alpine-sdk

COPY . .
RUN go build

FROM alpine:latest

COPY --from=build /app/tw-econ-telegram-bridge /
ENTRYPOINT ["/tw-econ-telegram-bridge"]