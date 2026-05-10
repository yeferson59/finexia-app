.PHONY: lint run db cache down fmt

lint:
	@golangci-lint run ./...

run:
	@go run cmd/api/main.go

db:
	@docker-compose up -d db

cache:
	@docker-compose up -d cache

down:
	@docker-compose down

fmt:
	@go fmt ./...
	@golangci-lint fmt ./...
