FROM golang:1.27-alpine AS builder
WORKDIR /app

COPY go.mod go.sum /
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o redis-server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/redis-server .

EXPOSE 50051
CMD ["./redis-server"]
