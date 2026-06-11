# File: Makefile
# Purpose: Build automation for MeshGuard backend and dashboard

BINARY_NAME=meshguard-api
DASHBOARD_DIR=apps/dashboard
API_DIR=apps/api
DATA_DIR=data

.PHONY: all build setup run-api run-dashboard clean deps

all: build

deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

build: deps
	@echo "Building MeshGuard API..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./$(API_DIR)/main.go

setup:
	@echo "Creating data directory..."
	mkdir -p $(DATA_DIR)
	@echo "Installing dashboard dependencies..."
	cd $(DASHBOARD_DIR) && npm install

run-api: build
	@echo "Starting MeshGuard API server..."
	./bin/$(BINARY_NAME)

run-dashboard:
	@echo "Starting dashboard dev server..."
	cd $(DASHBOARD_DIR) && npm run dev

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf $(DATA_DIR)/*.db
	rm -rf $(DASHBOARD_DIR)/node_modules
	rm -rf $(DASHBOARD_DIR)/dist
