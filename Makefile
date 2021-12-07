PROJECT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SHELL = /bin/bash

test:	init
	@echo "### Running unit tests..."
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic ./internal/... ./cmd/...

run:	init
	go run ./cmd/codd/main.go

init:
	@cd $(PROJECT_DIR)
	go mod tidy
