.PHONY: build test clean run help

BINARY_NAME=crypto-cli

help:
	@echo "Available targets:"
	@echo "  build    - Build the CLI application"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  run      - Run the application with default args"
	@echo "  dev      - Run in development mode (with auto-rebuild)"
	@echo "  deps     - Install dependencies"

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/cli
	@echo "Build complete: ./$(BINARY_NAME)"

test:
	@echo "Running tests..."
	@go test ./tests/... -v

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f *.txt *.enc *.dec
	@echo "Clean complete"

run: build
	@./$(BINARY_NAME) help

dev:
	@echo "Starting in dev mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Installing air for live reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

deps:
	@echo "Installing tools..."
	@go install github.com/cosmtrek/air@latest
	@echo "Dependencies installed"