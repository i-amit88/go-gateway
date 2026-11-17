APP_NAME := gateway
UPSTREAM_NAME := upstream
BIN_DIR := bin
GATEWAY_BINARY := $(BIN_DIR)/$(APP_NAME)
UPSTREAM_BINARY := $(BIN_DIR)/$(UPSTREAM_NAME)

.PHONY: build build-gateway build-upstream run run-gateway run-upstream run-both

build: build-gateway build-upstream

build-gateway:
	mkdir -p $(BIN_DIR)
	go build -o $(GATEWAY_BINARY) ./cmd/gateway

build-upstream:
	mkdir -p $(BIN_DIR)
	go build -o $(UPSTREAM_BINARY) ./cmd/upstream

run: run-gateway

run-gateway: build-gateway
	go run ./cmd/gateway

run-upstream: build-upstream
	go run ./cmd/upstream

run-both: build
	@trap 'kill 0' INT TERM EXIT; ./$(UPSTREAM_BINARY) & ./$(GATEWAY_BINARY) & wait
