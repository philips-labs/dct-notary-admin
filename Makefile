export GOPRIVATE := github.com/philips-labs/*

NOTARY_REPO ?= $(CURDIR)/notary
SANDBOX_COMPOSE ?= $(NOTARY_REPO)/docker-compose.sandbox.yml

VERSION := 0.0.0-dev
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X main.commit=$(GITCOMMIT) -X main.version=$(VERSION) -X main.date=$(shell date +%FT%TZ)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

.PHONY: all
all: build test

.PHONY: run
run: build
	@bin/dctna --config .notary/config.json

.PHONY: build-sandbox
build-sandbox:
	@(cd $(NOTARY_REPO) ; make cross ; docker-compose build)
	@docker-compose build

.PHONY: clean-dangling-images
clean-dangling-images:
	@docker rmi $$(docker images -qf dangling=true)

.PHONY: run-sandbox
run-sandbox: build-sandbox
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml up -d
	@echo
	@echo Too get logs:
	@echo "  make sandbox-logs"
	@echo
	@echo Too enter the sandbox:
	@echo "  docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox sh"

.PHONY: bootstrap-sandbox
bootstrap-sandbox:
	@docker cp bootstrap-sandbox.sh notary_sandbox_1:/root/
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox ./bootstrap-sandbox.sh

.PHONY: sandbox-logs
sandbox-logs:
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml logs -f

.PHONY: sandbox-logs
stop-sandbox:
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml down

.PHONY: download
download:
	@echo Downloading dependencies
	@go mod download

.PHONY: test
test:
	@echo Testing
	@go test -v -count=1 ./...

.PHONY: coverage
coverage:
	@echo Testing with code coverage
	@go test -v -count=1 -covermode=atomic -coverprofile=coverage.out ./...

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
	@go build -a ${GO_LDFLAGS} -o bin/dctna .

build-static: download
	@echo Building binary
	@go build -a -installsuffix cgo ${GO_LDFLAGS_STATIC} -o bin/static/dctna .

.PHONY: certs
certs:
	@echo Create SSL certificates
	@mkdir -p certs
	@openssl req \
       -newkey rsa:2048 -nodes -keyout certs/server.key \
	   -subj "/C=NL/O=Philips Labs/CN=localhost:8443" \
       -new -out certs/server.csr
	@openssl x509 \
       -signkey certs/server.key \
       -in certs/server.csr \
       -req -days 365 -out certs/server.crt
	openssl x509 -text -noout -in certs/server.crt
