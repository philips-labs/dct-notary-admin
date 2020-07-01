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

.PHONY: help all run build-sandbox clean-dangling-images run-sandbox check-sandbox bootstrap-sandbox sandbox-logs stop-sandbox reset-sandbox download test coverage coverage-out coverage-html build build-static certs dockerize outdated

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

all: build test ## Build and test

run: build ## Run dctna server
	@bin/dctna --config .notary/config.json

build-sandbox: ## build Docker images for notary sandbox
	@(cd $(NOTARY_REPO) ; make cross ; docker-compose -f docker-compose.sandbox.yml build)
	@docker-compose build

clean-dangling-images: ## Clean dangling Docker images
	@docker rmi $$(docker images -qf dangling=true)

run-sandbox: ## Run notary sandbox in Docker
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml up -d
	@echo
	@echo Too get logs:
	@echo "  make sandbox-logs"
	@echo
	@echo Too enter the sandbox:
	@echo "  docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox sh"

check-sandbox: ## Check if the notary sandbox is up and running
	@while [[ "$$(curl --insecure -sLSo /dev/null -w ''%{http_code}'' $(SANDBOX_HEALTH))" != "200" ]]; \
	do echo "Waiting for $(SANDBOX_HEALTH)" && sleep 1; \
	done
	@echo $(SANDBOX_HEALTH)
	@curl -X GET -IL --insecure ${SANDBOX_HEALTH}

bootstrap-sandbox: ## Bootstrap the notary sandbox with some certificates for content trust
	@docker cp bootstrap-sandbox.sh notary_sandbox_1:/root/
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml exec sandbox ./bootstrap-sandbox.sh

sandbox-logs: ## Tail the Docker logs
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml logs -f

stop-sandbox: ## Stop the vault notary sandbox environment
	@docker-compose -f $(SANDBOX_COMPOSE) -f docker-compose.yml down

reset-sandbox: ## Reset the Notary sandbox
	@echo Shutting down sandbox
	@docker-compose -f $(SANDBOX_COMPOSE) down &> /dev/null
	@echo Cleaning volumes
	@docker volume rm $$(docker-compose -f $(SANDBOX_COMPOSE) config --volumes | sed 's/^/notary_/g') 2> /dev/null || true

download: ## Download go dependencies
	@echo Downloading dependencies
	@go mod download

test: reset-sandbox ## Run the tests
	@echo Testing
	@docker-compose -f $(SANDBOX_COMPOSE) up -d
	@make check-sandbox
	@go test -race -v -count=1 ./...

coverage: reset-sandbox ## Run the tests with coverage
	@echo Testing with code coverage
	@docker-compose -f $(SANDBOX_COMPOSE) up -d
	@make check-sandbox
	@go test -race -v -count=1 -covermode=atomic -coverprofile=coverage.out ./...

coverage-out: coverage ## Output code coverage at the CLI
	@echo Coverage details
	@go tool cover -func=coverage.out

coverage-html: coverage ## Output code coverage as HTML
	@go tool cover -html=coverage.out

build: download ## Build the binary
	@echo Building binary
	@go build -a ${GO_LDFLAGS} -o bin/dctna-server ./cmd/dctna-server

build-static: download
	@echo Building binaries
	@go build -a -installsuffix cgo ${GO_LDFLAGS_STATIC} -o bin/static/dctna ./cmd/dctna
	@go build -a -installsuffix cgo ${GO_LDFLAGS_STATIC} -o bin/static/dctna-server ./cmd/dctna-server

certs: ## Creates selfsigned TLS certificates
	@echo Create TLS certificates
	@mkdir -p certs
	@openssl req \
       -newkey rsa:2048 -nodes -keyout certs/server.key \
	   -subj "/C=NL/O=Philips Labs/CN=localhost" \
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
	docker build --pull --force-rm -t dctna-web web
	docker build --pull --force-rm -f server.Dockerfile -t dctna-server .
	docker build --pull --force-rm -t dctna .

docker-publish-hsdp: ## publishes the image to the hsdp registry
ifndef HSDP_DOCKER_REGISTRY_USER
	$(error HSDP_DOCKER_REGISTRY_USER is undefined)
endif
ifndef HSDP_DOCKER_REGISTRY_PASSWD
	$(error HSDP_DOCKER_REGISTRY_PASSWD is undefined)
endif
ifndef HSDP_DOCKER_REGISTRY
	$(error HSDP_DOCKER_REGISTRY is undefined)
endif
ifndef HSDP_DOCKER_REGISTRY_NS
	$(error HSDP_DOCKER_REGISTRY_NS is undefined)
endif
	docker tag dctna-web:latest $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna-web:latest
	docker tag dctna-server:latest $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna-server:latest
	docker tag dctna:latest $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna:latest
	@echo $(HSDP_DOCKER_REGISTRY_PASSWD) > .pass
	cat .pass | docker login $(HSDP_DOCKER_REGISTRY) -u $(HSDP_DOCKER_REGISTRY_USER) --password-stdin
	@rm .pass
	docker push $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna-web:latest
	docker push $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna-server:latest
	docker push $(HSDP_DOCKER_REGISTRY)/$(HSDP_DOCKER_REGISTRY_NS)/dctna:latest
	docker logout $(HSDP_DOCKER_REGISTRY)

outdated: ## Checks for outdated dependencies
	go list -u -m -json all | go-mod-outdated -update
