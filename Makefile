SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -euo pipefail -c
MAKEFLAGS += --no-builtin-rules
.DEFAULT_GOAL := help

VENV := .venv
PY := $(VENV)/bin/python

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help setup venv install lint go-test py-test test migrate ci cloc format clean clean-all

help: # @help Show available targets
	@awk 'BEGIN {FS=":.*?# @help "} \
	/^[a-zA-Z0-9_.-]+:.*?# @help / {printf "  %-18s %s\n", $$1, $$2}' \
	$(firstword $(MAKEFILE_LIST))

venv: # @help Ensure virtual environment
	@test -d $(VENV) || python3 -m venv $(VENV)

install: venv # @help Install dev tooling
	pip install --upgrade pip
	./scripts/install.sh

setup: install # @help Configure dev environment

lint: # @help Run format and lint checks
	test -z "$$(gofmt -l .)"
	go vet ./...
	ruff format --check .
	ruff check .

go-test: # @help Run go tests
	go test ./... -count=1 -race -cover

py-test: venv # @help Run Python e2e tests
	pytest test/e2e -vn auto

test: go-test py-test # @help Run all tests

migrate: # @help Run DB migrations against TEST_DATABASE_URL
	@test -n "$${TEST_DATABASE_URL:-}" || (echo "TEST_DATABASE_URL is not set"; exit 1)
	goose -dir migrations postgres "$${TEST_DATABASE_URL}" up

ci: install migrate lint test # @help Run full CI pipeline

cloc: # @help Count lines of code
	cloc web/ internal/ test/ migrations/ scripts/ cmd/                                                                                                                                      ─╯

format: # @help Autoformat source code
	gofmt -w .
	ruff format .

clean: # @help Remove caches and build artifacts
	find . -type d -name "__pycache__" -prune -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	rm -rf .pytest_cache .ruff_cache
	go clean
	rm -f run-app coverage.out
	find . -name ".DS_Store" -delete

clean-all: clean # @help Also remove venv + Go module cache
	rm -rf $(VENV)
	go clean -modcache
