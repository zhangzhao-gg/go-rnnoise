# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=rnnoise-cli
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the example CLI tool
.PHONY: build
build:
	cd cmd/rnnoise-cli && $(GOBUILD) -o $(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all:
	cd cmd/rnnoise-cli && \
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 . && \
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64 . && \
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)-darwin-arm64 . && \
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe .

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f cmd/rnnoise-cli/$(BINARY_NAME)*
	rm -f cmd/rnnoise-cli/*.exe

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linting
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Run the example CLI with help
.PHONY: help
help:
	cd cmd/rnnoise-cli && $(GOBUILD) -o $(BINARY_NAME) . && ./$(BINARY_NAME)

# Run a simple test
.PHONY: demo
demo:
	cd cmd/rnnoise-cli && $(GOBUILD) -o $(BINARY_NAME) . && ./$(BINARY_NAME) stream

# Build and run examples
.PHONY: examples
examples:
	$(GOBUILD) -o examples/basic/simple-demo examples/basic/simple.go
	$(GOBUILD) -o examples/advanced/batch-demo examples/advanced/batch_processing.go
	$(GOBUILD) -o examples/streaming/streaming-demo examples/streaming/real_time.go

# Install development tools
.PHONY: install-tools
install-tools:
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest

# Security check
.PHONY: security
security:
	govulncheck ./...

# Generate documentation
.PHONY: doc
doc:
	$(GOCMD) doc -all ./rnnoise

# Run all checks
.PHONY: check
check: fmt lint test security

# Default target
.DEFAULT_GOAL := build
