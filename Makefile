ARTIFACTS_FOLDER=.artifacts

# Default target
all: help

run:
	@echo "Running the project..."
	@go run cmd/gommarizer/main.go

test:
	@echo "Running unit tests..."
	@go test -v ./...

test_rustorka:
	@echo "Running rustorka tests..."
	@mkdir -p $(ARTIFACTS_FOLDER)
	@cd $(ARTIFACTS_FOLDER) && go run ../test/rustorka_test/main.go && sed -i '/kot.png/d' rustorka_topic*html

deps:
	@echo "Installing dependencies..."
	@go mod tidy

clean:
	@echo "Cleaning artifacts..."
	@rm -rf $(ARTIFACTS_FOLDER)

help:
	@echo "Makefile commands:"
	@echo "  make               - Show this help message"
	@echo "  make run           - Run the project"
	@echo "  make test          - Run unit tests"
	@echo "  make test_rustorka - Run rustorka scrapper test"
	@echo "  make clean         - Clean $(ARTIFACTS_FOLDER)"
	@echo "  make help          - Show this help message"

.PHONY: all run test help test_rustorka deps clean