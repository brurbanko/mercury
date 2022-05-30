BINARY     := bga-crawler
BUILDFLAGS := "-s -w"
WATCHFILES := (.+\.go|.+\.toml?)$
UNAME_S := $(shell uname -s)
PLATFORM := darwin
ifeq ($(UNAME_S),Linux)
	PLATFORM = linux
endif

.SUFFIXES:
.PHONY: help \
		build dev dependencies lint test

.DEFAULT_GOAL := help

build: dependencies ## Build development binary file for development server
	@mkdir -p build
	GOOS=darwin             go build -ldflags $(BUILDFLAGS) -o build/$(BINARY)-darwin cmd/crawler/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o build/$(BINARY)-linux cmd/crawler/main.go

dev: dependencies compiledaemon ## Run development server with CompileDaemon
	@printf "\033[36m%s\033[0m\n" "Starting development server"
	CompileDaemon -color=true -pattern='$(WATCHFILES)' \
		-build="make build" -command="./build/$(BINARY)-$(PLATFORM)" \
		-exclude-dir=".git" -exclude-dir=".idea" -exclude-dir="vendor" \
		-exclude-dir="data" -exclude-dir="_data" -exclude-dir="build"

dependencies: ## Install dependencies needed for project
	@printf "\033[36m%s\033[0m\n" "Installing dependencies for project:"
	go mod download
	@printf "\033[36m%s\033[0m\n" "Getting the used modules to be tidy again..."
	go mod tidy

compiledaemon:
ifeq (, $(shell which CompileDaemon))
	@printf "\033[36m%s\033[0m\n" "Installing compile daemon..."
	go get -v -u github.com/githubnemo/CompileDaemon
endif

lint: ## Run linter
	golangci-lint run

test: ## Run tests
	go test -v ./...

# Auto documented Makefile https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'