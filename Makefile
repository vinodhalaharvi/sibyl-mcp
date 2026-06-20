BINARY      := sibyl-mcp
PKG         := ./...
COVERAGE    := coverage.out

GO          ?= go
GOFLAGS     ?=

.DEFAULT_GOAL := help

## help: list available targets
.PHONY: help
help:
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

## build: compile the binary into ./bin
.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o bin/$(BINARY) .

## run: run the MCP server (speaks MCP over stdio; needs a Temporal worker running)
.PHONY: run
run:
	$(GO) run .

## test: run tests with the race detector
.PHONY: test
test:
	$(GO) test $(GOFLAGS) -race $(PKG)

## cover: run tests and open an HTML coverage report
.PHONY: cover
cover:
	$(GO) test $(GOFLAGS) -race -coverprofile=$(COVERAGE) $(PKG)
	$(GO) tool cover -html=$(COVERAGE)

## vet: run go vet
.PHONY: vet
vet:
	$(GO) vet $(PKG)

## fmt: format all Go files in place
.PHONY: fmt
fmt:
	gofmt -w .

## fmt-check: fail if any Go file is not gofmt-clean
.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt-clean:"; echo "$$unformatted"; exit 1; \
	fi

## tidy: sync go.mod / go.sum
.PHONY: tidy
tidy:
	$(GO) mod tidy

## lint: run golangci-lint (install: https://golangci-lint.run/welcome/install/)
.PHONY: lint
lint:
	golangci-lint run

## check: the local equivalent of CI (fmt-check + vet + test)
.PHONY: check
check: fmt-check vet test

## clean: remove build artifacts
.PHONY: clean
clean:
	rm -rf bin $(COVERAGE)
