.PHONY: all
all: build test

.PHONY: run
run: build
	@bin/dctna

.PHONY: download
download:
	@echo Downloading dependencies
	@go mod download

.PHONY: test
test:
	@echo Testing
	@go test -v ./...

.PHONY: coverage
coverage:
	@echo Testing with code coverage
	@go test -v -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: coverage-out
coverage-out: coverage
	@echo Coverage details
	@go tool cover -func=coverage.out

.PHONY: coverage-htlm
coverage-html: coverage
	@go tool cover -html=coverage.out

.PHONY: build
build: download
	@echo Building binary
	@go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/dctna .
