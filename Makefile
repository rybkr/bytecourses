SHELL := /usr/bin/env bash

GO   := go
PKGS := ./...
CMD  := ./cmd/server

BIN_DIR := bin
SERVER_BIN := $(BIN_DIR)/server

.PHONY: help build run dev test test-go test-py fmt vet tidy clean

help:
	@echo "targets:"
	@echo "  build     build server binary"
	@echo "  run       run server"
	@echo "  run-dev   run server in development mode"
	@echo "  test      run all tests"
	@echo "  test-go   run go tests"
	@echo "  test-py   run pytest"
	@echo "  format    format go code"
	@echo "  vet       go vet"
	@echo "  tidy      go mod tidy"
	@echo "  clean     remove build artifacts"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) $(CMD)

run:
	$(GO) run $(CMD)

run-dev:
	$(GO) run $(CMD) --seed-users=true

test: test-go test-py

test-go:
	$(GO) test $(PKGS) -cover

test-py:
	pytest -vn auto

format:
	gofmt -w .
	$(GO) fmt $(PKGS)
	ruff format .

vet:
	$(GO) vet $(PKGS)

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)
