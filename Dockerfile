FROM golang:alpine as builder

RUN apk update && apk add alpine-sdk && mkdir /build
COPY go.mod /build
RUN cd /build && go mod download

COPY . /build
RUN cd /build && go build -ldflags="-s -w" -o feed2tg .

#--------------------------
FROM alpine
COPY --from=builder /build/feed2tg  /usr/bin/feed2tg

WORKDIR /opt
CMD ["/usr/bin/feed2tg"]