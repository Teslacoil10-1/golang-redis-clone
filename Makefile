
.PHONEY: up down test lint build

up:
	docker-compose up -d --build
down:
	docker-compose down
test:
	go test -v -race ./...
lint:
	golangci-lint run
