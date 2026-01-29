.PHONY: build install test clean

# Build variables
BINARY_NAME=pkt
BUILD_DIR=bin
INSTALL_DIR=/usr/local/bin

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

# Run in development
dev: build
	@$(BUILD_DIR)/$(BINARY_NAME)
