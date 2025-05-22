.PHONY: lint format test cover cover-html benchmark check install clean tidy deps docs changelog release version help all

# Version information (can be used in code)
VERSION ?= $(shell grep -r 'const Version = ' version.go | grep -o '"[^"]*"' | sed 's/"//g')

# Default target
.DEFAULT_GOAL := help

# Run all standard checks
all: format lint test
	@echo "All checks passed!"

lint: deps
	go vet ./...
	golangci-lint run

format:
	go fmt ./...

test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out

benchmark:
	go test -bench=. -benchmem ./...

# Run all checks
check: lint test cover

# Install the package
install:
	go install ./...

# Clean build artifacts
clean:
	rm -f coverage.out
	go clean

# Update Go module dependencies
tidy:
	go mod tidy

deps:
	go mod download
	go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

# Generate documentation
docs:
	@echo "Generating documentation..."
	@echo "Visit http://localhost:6060/github.com/bkovacki/gopenrouter"
	pkgsite -http=:6060 &

changelog:
	git-chglog -o CHANGELOG.md --next-tag v${VERSION}

release:
	@echo "Creating release v$(VERSION)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "Release v$(VERSION) created and pushed"

# Print version information
version:
	@go version
	@echo "Module: $(shell grep "^module" go.mod | cut -d ' ' -f 2)"
	@echo "Go version: $(shell grep "^go" go.mod | cut -d ' ' -f 2)"
	@echo "Version: $(VERSION)"

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Run format, lint, and test"
	@echo "  lint         - Run linters (go vet and golangci-lint)"
	@echo "  format       - Format code with go fmt"
	@echo "  test         - Run tests"
	@echo "  cover        - Generate test coverage report"
	@echo "  cover-html   - Generate and display HTML coverage report"
	@echo "  benchmark    - Run benchmarks"
	@echo "  check        - Run all checks (lint, test, cover)"
	@echo "  install      - Install the package"
	@echo "  clean        - Clean build artifacts"
	@echo "  tidy         - Update Go module dependencies"
	@echo "  deps         - Download Go module tools and dependencies"
	@echo "  docs         - Generate and serve documentation"
	@echo "  version      - Display version information"
	@echo "  help         - Show this help message (default)"
