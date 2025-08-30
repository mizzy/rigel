.PHONY: build run test clean install deps lint

BINARY_NAME=rigel
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=cmd/rigel/main.go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X 'github.com/mizzy/rigel/internal/version.Version=$(VERSION)' \
           -X 'github.com/mizzy/rigel/internal/version.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/mizzy/rigel/internal/version.BuildDate=$(BUILD_DATE)'

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) $(MAIN_PATH)

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
	go install -ldflags "$(LDFLAGS)" ./cmd/rigel

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
