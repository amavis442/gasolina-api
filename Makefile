ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY_EXT  := .exe
    RM          := cmd /C del /Q /F
    RMDIR       := cmd /C rmdir /S /Q
    MKDIR       := cmd /C mkdir
    SEP         := \\
else
    DETECTED_OS := $(shell uname -s)
    BINARY_EXT  :=
    RM          := rm -f
    RMDIR       := rm -rf
    MKDIR       := mkdir -p
    SEP         := /
endif

BINARY  := gasolina-api$(BINARY_EXT)
BIN_DIR := bin

.PHONY: help build run test e2e swagger clean

help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@findstr /R "^[a-zA-Z_-]*:.*##" $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
endif

swagger: ## Regenerate Swagger docs (installs swag if missing)
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	swag init --generalInfo cmd/api/main.go --output docs

build: swagger ## Build the API binary — generates Swagger docs first
	@echo Building for $(DETECTED_OS)...
	go build -o $(BIN_DIR)$(SEP)$(BINARY) ./cmd/api
	@echo Build complete: $(BIN_DIR)$(SEP)$(BINARY)

run: ## Run the API server
	go run ./cmd/api

test: ## Run all tests
	go test ./...

e2e: ## Run Deno E2E tests against a running server
	cd e2e && deno task test

clean: ## Remove build artifacts
	$(RMDIR) $(BIN_DIR)
