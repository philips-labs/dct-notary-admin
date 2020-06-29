export GOPRIVATE := github.com/philips-labs/*

NOTARY_REPO ?= $(CURDIR)/notary
SANDBOX_COMPOSE ?= $(NOTARY_REPO)/docker-compose.sandbox.yml
SANDBOX_HEALTH ?= https://localhost:4443/_notary_server/health

VERSION := 0.0.0-dev
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X main.commit=$(GITCOMMIT) -X main.version=$(VERSION) -X main.date=$(shell date +%FT%TZ)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"
PLANTUML_JAR_URL = https://sourceforge.net/projects/plantuml/files/plantuml.jar/download
DIAGRAMS_SRC := $(wildcard docs/diagrams/*.plantuml)
DIAGRAMS_PNG := $(addsuffix .png, $(basename $(DIAGRAMS_SRC)))
DIAGRAMS_SVG := $(addsuffix .svg, $(basename $(DIAGRAMS_SRC)))

help:
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build test ## Build and test

.PHONY: run
run: build ## Run dctna server
	@bin/dctna --config .notary/config.json

.PHONY: build-sandbox
build-sandbox: ## build Docker images for notary sandbox
	@(cd $(NOTARY_REPO) ; make cross ; docker-compose -f docker-compose.sandbox.yml build)
	@docker-compose build

.PHONY: clean-dangling-images
clean-dangling-images: ## Clean dangling Docker images
	@docker rmi $$(docker images -qf dangling=true)

.PHONY: run-sandbox
run-sandbox: ## Run notary sandbox in Docker
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml up -d
	@echo
	@echo Too get logs:
	@echo "  make sandbox-logs"
	@echo
	@echo Too enter the sandbox:
	@echo "  docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox sh"

.PHONY: check-sandbox
check-sandbox: ## Check if the notary sandbox is up and running
	@while [[ "$$(curl --insecure -sLSo /dev/null -w ''%{http_code}'' $(SANDBOX_HEALTH))" != "200" ]]; \
	do echo "Waiting for $(SANDBOX_HEALTH)" && sleep 1; \
	done
	@echo $(SANDBOX_HEALTH)
	@curl -X GET -IL --insecure ${SANDBOX_HEALTH}

.PHONY: bootstrap-sandbox
bootstrap-sandbox: ## Bootstrap the notary sandbox with some certificates for content trust
	@docker cp bootstrap-sandbox.sh notary_sandbox_1:/root/
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox ./bootstrap-sandbox.sh

.PHONY: sandbox-logs
sandbox-logs: ## Tail the Docker logs
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml logs -f

.PHONY: sandbox-logs
stop-sandbox: ## Stop the vault notary sandbox environment
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml down

.PHONY: reset-sandbox
reset-sandbox: ## Reset the Notary sandbox
	@echo Shutting down sandbox
	@docker-compose -f $(SANDBOX_COMPOSE) down &> /dev/null
	@echo Cleaning volumes
	@docker volume rm $$(docker-compose -f $(SANDBOX_COMPOSE) config --volumes | sed 's/^/notary_/g') 2> /dev/null || true

.PHONY: download
download: ## Download go dependencies
	@echo Downloading dependencies
	@go mod download

.PHONY: test
test: reset-sandbox ## Run the tests
	@echo Testing
	@docker-compose -f $(SANDBOX_COMPOSE) up -d
	@make check-sandbox
	@go test -race -v -count=1 ./...

.PHONY: coverage
coverage: reset-sandbox ## Run the tests with coverage
	@echo Testing with code coverage
	@docker-compose -f $(SANDBOX_COMPOSE) up -d
	@make check-sandbox
	@go test -race -v -count=1 -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: coverage-out
coverage-out: coverage ## Output code coverage at the CLI
	@echo Coverage details
	@go tool cover -func=coverage.out

.PHONY: coverage-htlm
coverage-html: coverage ## Output code coverage as HTML
	@go tool cover -html=coverage.out

.PHONY: build
build: download ## Build the binary
	@echo Building binary
	@go build -a ${GO_LDFLAGS} -o bin/dctna .

build-static: download ## Build the static binary
	@echo Building binary
	@go build -a -installsuffix cgo ${GO_LDFLAGS_STATIC} -o bin/static/dctna .

.PHONY: certs
certs: ## Creates selfsigned TLS certificates
	@echo Create TLS certificates
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

clean-diagrams: ## Cleans plantuml.jar and generated diagrams
	@rm -f plantuml.jar $(DIAGRAMS_PNG) $(DIAGRAMS_SVG)

diagrams: svg-diagrams png-diagrams ## Generate diagrams in SVG and PNG format
svg-diagrams: plantuml.jar $(DIAGRAMS_SVG) ## Generate diagrams in SVG format
png-diagrams: plantuml.jar $(DIAGRAMS_PNG) ## Generate diagrams in PNG format

plantuml.jar:
	@echo Downloading $@....
	@curl -sSfL $(PLANTUML_JAR_URL) -o $@

docs/diagrams/%.svg: docs/diagrams/%.plantuml
	@echo Generating $@ from plantuml....
	@java -jar plantuml.jar -tsvg $^

docs/diagrams/%.png: docs/diagrams/%.plantuml
	@echo Generating $@ from plantuml....
	@java -jar plantuml.jar -tpng $^

dockerize: ## builds docker images
	docker build -t dctna-web web
	docker build -t dctna-server .
	docker rmi $$(docker images -qf dangling=true)
