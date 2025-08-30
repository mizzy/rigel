.PHONY: build run test clean install deps lint

BINARY_NAME=rigel
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=cmd/rigel/main.go

build:
	go build -o $(BINARY_PATH) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v ./...

test-coverage:
	go test -cover ./...

clean:
	go clean
	rm -f $(BINARY_PATH)

install:
	go install ./cmd/rigel

deps:
	go mod download
	go mod tidy

lint:
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping linting"; \
	fi

dev: deps build
	$(BINARY_PATH)

help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install the binary"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linter"
	@echo "  dev           - Build and run for development"
	@echo "  help          - Show this help message"
