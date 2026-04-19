.DEFAULT_GOAL := help

BINARY    := babeliocli
LDFLAGS   := -s -w

.PHONY: build test lint check clean help

build: ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test: ## Run tests with race detector
	go test ./... -race

lint: ## Run golangci-lint
	golangci-lint run

check: lint test ## Run lint + test (same as CI)

clean: ## Remove build artifacts
	rm -f $(BINARY)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
