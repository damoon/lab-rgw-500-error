# build-env ###################################################
FROM golang:1.14.4 AS build-env
WORKDIR /app

COPY cmd cmd
RUN go build -o /proxy ./cmd

###############################################################
FROM debian:buster

RUN apt-get update && apt-get install -y \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build-env /proxy /usr/local/bin/proxy
ENTRYPOINT [ "proxy" ]
