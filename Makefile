BINARY_NAME=ghostenv
MAIN_PATH=./cmd/ghostenv/main.go
INSTALL_PATH=/usr/local/bin
BUILD_DIR=./bin

.PHONY: build install clean tidy test help lint vet fmt

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing to $(INSTALL_PATH)..."
	@sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)
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
	@echo "  build     Build the binary"
	@echo "  install   Build and install to $(INSTALL_PATH)"
	@echo "  clean     Remove build artifacts and vault files"
	@echo "  tidy      Run go mod tidy"
	@echo "  test      Run tests"
	@echo "  lint      Run go vet"
	@echo "  vet       Run go vet"
	@echo "  fmt       Format code with go fmt"
	@echo "  help      Show this help message"
