.DEFAULT_GOAL := help

# Load .env. see https://lithic.tech/blog/2020-05/makefile-dot-env
ifneq (,$(wildcard ./.env))
  include .env
  export
endif

# Environment
GO111MODULE := on
PATH := $(CURDIR)/.go-tools/bin:$(PATH)
SHELL := bash

# This is a magic code to output help message at default
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[/0-9a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## Format go code
	@go fmt $$(go list ./... | grep -v vendor)

.PHONY: deps
deps: ## Install go modules
	@go mod tidy

.PHONY: tools/install
tools/install: ## Install dev tools
	@export GO111MODULE=off && export GOPATH=$(CURDIR)/.go-tools && \
		go get -u github.com/axw/gocov/gocov && \
		go get -u gopkg.in/matm/v1/gocov-html
	@rm -rf $(CURDIR)/.go-tools/pkg && rm -rf $(CURDIR)/.go-tools/src

.PHONY: tools/clean
tools/clean: ## Clean installed tools
	@export GO111MODULE=off && export GOPATH=$(CURDIR)/.go-tools && rm -rf $(CURDIR)/.go-tools

.PHONY: test
test: ## Test go code
	@go test -cover ./...

.PHONY: test/verbose
test/verbose: ## Run all tests with verbose outputting.
	@go test -v -cover ./...

.PHONY: test/coverage
test/coverage: ## Run all tests with coverage report outputting.
	@gocov test ./... | gocov-html > coverage-report.html

.PHONY: clean
clean:  ## Clean the generated contents
	@rm -rf coverage-report.html
