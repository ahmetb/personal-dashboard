FROM alpine:latest
MAINTAINER Ahmet Alp Balkan

RUN apk --update upgrade && \
    apk add ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY ./bin /bin
