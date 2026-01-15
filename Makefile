.PHONY: help install lint go-test py-test test migrate ci

help: # @help Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?# @help ' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?# @help "}; {printf "  %-12s %s\n", $$1, $$2}'

install: # @help Install dev tooling
	go install github.com/pressly/goose/v3/cmd/goose@latest
	python -m pip install --upgrade pip
	./scripts/install.sh

lint: # @help Run format and lint checks
	test -z "$$(gofmt -l .)"
	go vet ./...
	ruff format --check .
	ruff check .

go-test: # @help Run go tests
	go test ./... -count=1 -race -cover

py-test: # @help Run Python tests
	pytest -vn auto

test: go-test py-test # @help Run all tests

migrate: # @help Run DB migrations
	goose -dir migrations postgres "$$TEST_DATABASE_URL" up

ci: install migrate lint test # @help Run full CI pipeline

cloc: # @help Count lines of code
	cloc web/ internal/ test/ migrations/ scripts/ cmd/                                                                                                                                      ─╯
