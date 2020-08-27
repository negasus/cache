SHELL       =   /bin/sh
PKG_PREFIX  :=  github.com/negasus/cache
TAG         ?=  latest

.SUFFIXES:
.PHONY: help test

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	GO111MODULE=on go test -coverprofile=coverage.txt -covermode=atomic .
