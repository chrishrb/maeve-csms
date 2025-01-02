GOCMD=go

.PHONY: all
all: generate build format lint gosec integration ## Format, lint, build and test

.PHONY: init
init: ## Init
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: run
run: ## Run
	docker-compose build manager && docker-compose up -d

.PHONY: build
build: ## Build
	${GOCMD} build -o bin/manager main.go

.PHONY: generate
generate: ## Generate
	${GOCMD} generate ./...
	${GOCMD} mod tidy

.PHONY: test
test: ## Test
	${GOCMD} test ./...

.PHONY: integration
integration: ## Run unit and integration tests
	${GOCMD} test -tags=integration ./...

.PHONY: gosec
gosec: ## Run unit and integration tests
	gosec ./...

.PHONY: format
format: ## Format code
	${GOCMD} fmt ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: clean
clean: ## Cleanup build dir
	rm -r bin/

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
