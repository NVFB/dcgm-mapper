# Makefile for dcgm-mapper

# Binary name
BINARY_NAME=dcgm-mapper

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Build flags
LDFLAGS=-ldflags "-s -w"

# Docker parameters
DOCKER_IMAGE=ghcr.io/yourusername/dcgm-mapper
DOCKER_TAG?=latest
DOCKER_PLATFORM?=linux/amd64,linux/arm64

.PHONY: all build run run-daemon test clean coverage help install fmt vet docker-build docker-build-multiarch docker-push docker-run docker-run-interactive

# Default target
default: help

# Run all tests and build the binary
all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## run: Run the program once (single execution)
run:
	@echo "Running $(BINARY_NAME)..."
	$(GORUN) .

## run-daemon: Run the program in daemon mode (continuous monitoring)
run-daemon:
	@echo "Running $(BINARY_NAME) in daemon mode..."
	$(GORUN) . -daemon -interval 10s -verbose

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-short: Run tests in short mode
test-short:
	@echo "Running tests (short mode)..."
	$(GOTEST) -v -short ./...

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

## clean: Remove binary and test artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) .
	@echo "Installation complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

## docker-build-multiarch: Build multi-architecture Docker image
docker-build-multiarch:
	@echo "Building multi-arch Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker buildx build --platform $(DOCKER_PLATFORM) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Multi-arch Docker image built successfully"

## docker-push: Push Docker image to registry
docker-push:
	@echo "Pushing Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker image pushed successfully"

## docker-run: Run Docker container locally
docker-run:
	@echo "Running Docker container..."
	docker run --rm --gpus all \
		-v /tmp/dcgm-job-mapping:/var/lib/dcgm-mapper \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

## docker-run-interactive: Run Docker container interactively
docker-run-interactive:
	@echo "Running Docker container interactively..."
	docker run --rm -it --gpus all \
		-v /tmp/dcgm-job-mapping:/var/lib/dcgm-mapper \
		$(DOCKER_IMAGE):$(DOCKER_TAG) /bin/sh

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

