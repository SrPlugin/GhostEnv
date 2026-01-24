BINARY_NAME=ghostenv
MAIN_PACKAGE=./cmd/ghostenv
INSTALL_PATH=/usr/local/bin
BUILD_DIR=./bin

UNAME_S := $(shell uname -s 2>/dev/null || echo "Linux")
ifeq ($(UNAME_S),Linux)
	INSTALL_CMD = sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)
endif
ifeq ($(UNAME_S),Darwin)
	INSTALL_CMD = sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)
endif
ifeq ($(UNAME_S),Windows_NT)
	INSTALL_CMD = echo "Use Makefile.windows on Windows"
endif

.PHONY: build install clean tidy test help lint vet fmt build-windows build-linux build-darwin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"

build-darwin:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-darwin*"

install: build
	@echo "Installing to $(INSTALL_PATH)..."
	@$(INSTALL_CMD)
	@echo "Installed. You can now use '$(BINARY_NAME)' anywhere."

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.gev
	@rm -f vault.gev
	@echo "Clean complete."

tidy:
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Modules tidied."

test:
	@echo "Running tests..."
	@go test -v ./...

lint:
	@echo "Running linter..."
	@go vet ./...
	@echo "Lint complete."

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete."

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete."

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build the binary for current platform"
	@echo "  build-windows  Build for Windows (amd64)"
	@echo "  build-linux    Build for Linux (amd64)"
	@echo "  build-darwin   Build for macOS (amd64 and arm64)"
	@echo "  install        Build and install to $(INSTALL_PATH)"
	@echo "  clean          Remove build artifacts and vault files"
	@echo "  tidy           Run go mod tidy"
	@echo "  test           Run tests"
	@echo "  lint           Run go vet"
	@echo "  vet            Run go vet"
	@echo "  fmt            Format code with go fmt"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Note: On Windows, use 'make -f Makefile.windows' or use the provided Makefile.windows"
