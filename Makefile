# MAH Build Configuration
BINARY_NAME=mah
VERSION?=$(shell cat VERSION 2>/dev/null || echo "0.1.0")
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directories
BUILD_DIR=build
DIST_DIR=dist

.PHONY: all build clean test deps tidy fmt lint install dev cross-compile

# Default target
all: clean fmt test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mah
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for development (no optimization)
dev:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mah
	@echo "✓ Development build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Tidy up go.mod
tidy:
	@echo "Tidying up dependencies..."
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w . 2>/dev/null || echo "goimports not installed, skipping..."

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed, skipping..."

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out
	@echo "✓ Clean complete"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Cross-compile for multiple platforms
cross-compile: clean
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mah
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mah
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mah
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mah
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mah
	
	@echo "✓ Cross-compilation complete:"
	@ls -la $(DIST_DIR)/

# Create release packages
release: cross-compile
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip $$binary.zip $$binary; \
		else \
			tar -czf $$binary.tar.gz $$binary; \
		fi; \
	done
	@echo "✓ Release packages created:"
	@ls -la $(DIST_DIR)/*.{tar.gz,zip} 2>/dev/null || echo "No packages created"

# Development helpers
run: dev
	@echo "Running $(BINARY_NAME) with sample config..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --help

# Quick development cycle
quick: fmt build
	@echo "✓ Quick build complete"

# Show help
help:
	@echo "MAH Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  dev             - Build for development (faster)"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  deps            - Install dependencies"
	@echo "  tidy            - Tidy go.mod"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  install         - Install binary to GOPATH/bin"
	@echo "  cross-compile   - Build for multiple platforms"
	@echo "  release         - Create release packages"
	@echo "  run             - Build and run with help"
	@echo "  quick           - Format and build quickly"
	@echo "  help            - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION         - Binary version (default: 0.1.0)"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make cross-compile VERSION=1.0.0"
	@echo "  make test"