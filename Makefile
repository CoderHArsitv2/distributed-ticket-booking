.PHONY: run build test tidy fmt vet migrate-up migrate-down

# Binary output
BIN := bin/server

run:
	go run ./cmd/server

build:
	go build -o $(BIN) ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

# Requires golang-migrate (https://github.com/golang-migrate/migrate)
# and DATABASE_URL exported in your environment.
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
