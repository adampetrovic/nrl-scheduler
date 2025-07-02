.PHONY: build test clean migrate-up migrate-down migrate-create

# Build variables
BINARY_NAME=nrl-scheduler
BUILD_DIR=./bin
GO=go
GOFLAGS=-v

# Build the application
build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

# Build CLI
build-cli:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-cli ./cmd/cli

# Run tests
test:
	$(GO) test -v ./...

# Run tests with race detector
test-race:
	$(GO) test -race -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f *.db *.sqlite *.sqlite3

# Database migrations
migrate-up:
	migrate -path ./migrations -database "sqlite3://nrl-scheduler.db" up

migrate-down:
	migrate -path ./migrations -database "sqlite3://nrl-scheduler.db" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Install go-migrate
install-migrate:
	go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Development run
run:
	$(GO) run ./cmd/api

# Format code
fmt:
	$(GO) fmt ./...

# Lint code
lint:
	golangci-lint run