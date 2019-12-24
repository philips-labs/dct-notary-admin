all: build test

run: build
	@bin/dctna

download:
	@echo Downloading dependencies
	@go mod download

test:
	@echo Testing
	@go test -v ./...

coverage:
	@echo Testing with code coverage
	@go test -v -covermode=atomic -coverprofile=coverage.out ./...

coverage-out: coverage
	@echo Coverage details
	@go tool cover -func=coverage.out

coverage-html: coverage
	@go tool cover -html=coverage.out

build: download
	@echo Building binary
	@go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/dctna .
