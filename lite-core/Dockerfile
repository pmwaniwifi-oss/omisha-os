FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/omisha-lite .
FROM alpine:3.21
RUN addgroup -S omisha && adduser -S -G omisha -u 10001 omisha && mkdir /data && chown omisha:omisha /data
COPY --from=builder /out/omisha-lite /usr/local/bin/omisha-lite
USER omisha
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/omisha-lite"]
