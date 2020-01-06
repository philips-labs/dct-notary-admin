.PHONY: all
all: build test

.PHONY: run
run: build
	@bin/dctna -notary-config-file ./notary-config.json

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

.PHONY: certs
certs:
	@echo Create SSL certificates
	@mkdir -p certs
	@openssl req \
       -newkey rsa:2048 -nodes -keyout certs/server.key \
	   -subj "/C=NL/O=Philips Labs/CN=localhost:8086" \
       -new -out certs/server.csr
	@openssl x509 \
       -signkey certs/server.key \
       -in certs/server.csr \
       -req -days 365 -out certs/server.crt
	openssl x509 -text -noout -in certs/server.crt
