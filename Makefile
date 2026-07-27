.PHONY: all build run test test-verbose test-race cover vet fmt fmt-check lint tidy clean check \
        frontend-install frontend-build frontend-test frontend-lint extension-build help

BINARY := bookminderapi
GO     ?= go

all: build

## Build the server binary
build:
	$(GO) build -o $(BINARY) .

## Run the server directly (go run)
run:
	$(GO) run .

## Run the test suite
test:
	$(GO) test .

## Run the test suite with verbose output
test-verbose:
	$(GO) test -v .

## Run the test suite with the race detector
test-race:
	$(GO) test -race .

## Run tests and generate a coverage report (coverage.out)
cover:
	$(GO) test -coverprofile=coverage.out .
	$(GO) tool cover -func=coverage.out

## Run go vet
vet:
	$(GO) vet .

## Format Go source files in place
fmt:
	$(GO) fmt ./...

## Fail if any Go file is not gofmt-formatted
fmt-check:
	@unformatted="$$(gofmt -l . | grep -v node_modules || true)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## Run golangci-lint (see .golangci.yml)
lint:
	golangci-lint run .

## Tidy go.mod/go.sum
tidy:
	$(GO) mod tidy

## Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out

## Run the full pre-commit checklist: formatting, vet, and tests
check: fmt-check vet test

## Install frontend (Vue) dependencies
frontend-install:
	cd frontend && npm install

## Build the frontend for production
frontend-build:
	cd frontend && npm run build

## Run frontend unit tests
frontend-test:
	cd frontend && npm run test:run

## Lint the frontend
frontend-lint:
	cd frontend && npm run lint

## Package the browser extension (Chrome/Firefox/Safari)
extension-build:
	cd extension && ./build.sh

## List available targets
help:
	@awk 'BEGIN {FS = ":.*##"} /^## / { desc=substr($$0,4) } /^[a-zA-Z0-9_-]+:/ && desc { target=$$1; sub(/:.*/, "", target); printf "  %-18s %s\n", target, desc; desc="" }' $(MAKEFILE_LIST)
