.PHONY: run build test tidy fmt vet

# Binary output
BIN := bin/server

run:
	go run .

build:
	go build -o $(BIN) .

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...
