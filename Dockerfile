FROM golang:1.27-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev rust cargo


COPY aof_engine/ ./aof_engine/
WORKDIR /app/aof_engine
RUN cargo build --release

WORKDIR /app
COPY go.mod go.sum /
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o redis-server ./cmd/server

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache libgcc

COPY --from=builder /app/redis-server .

EXPOSE 50051
CMD ["./redis-server"]
