
.PHONEY: up down test lint build cli

up:
	docker-compose up -d --build
down:
	docker-compose down
test:
	go test -v -race ./...
lint:
	golangci-lint run
cli:
	go run ./cmd/cli/main.go
