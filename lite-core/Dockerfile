FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/omisha-lite .
FROM scratch
COPY --from=builder /out/omisha-lite /omisha-lite
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/omisha-lite"]
