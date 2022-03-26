FROM golang as builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN go get ./... && \
    CGO_ENABLED=0 GOOS=linux go build -o fake-elasticache-memcached-proxy main.go

FROM alpine
COPY --from=builder /go/src/app/fake-elasticache-memcached-proxy /fake-elasticache-memcached-proxy
WORKDIR /
EXPOSE 11211
ENTRYPOINT ["./fake-elasticache-memcached-proxy"]