.PHONY: lint run db down fmt

lint:
	@golangci-lint run ./...

run:
	@go run cmd/api/main.go

db:
	@docker-compose up -d db

down:
	@docker-compose down

fmt:
	@go fmt ./...
	@golangci-lint fmt ./...
