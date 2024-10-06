# Adapted from https://www.thapaliya.com/en/writings/well-documented-makefiles/
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL:=help

.PHONY: test
test: ## Run all the tests
	go test -v ./...
	
.PHONY: build
build: ## Build the binary
	go build -o bin/ ./...
	
.PHONY: run
run: ## Run the binary
	go run cmd/aar/main.go
	
.PHONY: install
install: ## Install the binary in $GOPATH/bin
	go install ./cmd/aar