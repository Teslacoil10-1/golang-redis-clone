.PHONY: up down test lint build cli proto

up:
	docker-compose up -d --build

down:
	docker-compose down

test:
	go test -v -race ./...

lint:
	golangci-lint run

build:
	go build -o bin/server .

cli:
	go run ./cmd/cli/main.go

proto:
	cd proto && protoc --go_out=. --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false store.proto
