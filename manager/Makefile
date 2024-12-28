GOCMD=go

.PHONY: all
all: generate build format lint test ## Format, lint and build

.PHONY: init
init: ## Init
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	npm install -g widdershins

.PHONY: build
run: ## Build
	${GOCMD} run . serve --config-file ../config/manager/config.toml

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
integration: ## Integration-tests
	${GOCMD} test -tags=integration ./...

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
