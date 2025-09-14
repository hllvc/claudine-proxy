.PHONY: all run build test test-coverage fmt lint audit clean

GO := go

BINARY_NAME := proxy
MAIN := ./cmd/proxy

all: test build

run:
	$(GO) run $(MAIN)

build:
	$(GO) build -o $(BINARY_NAME) $(MAIN)

test:
	$(GO) test -race ./...

# Use -coverprofile for unit tests; -test.gocoverdir is for integration tests with built binaries
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

audit:
	$(GO) vet ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
