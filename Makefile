export GOPRIVATE := github.com/philips-labs/*

NOTARY_REPO ?= $(CURDIR)/notary
SANDBOX_COMPOSE ?= $(NOTARY_REPO)/docker-compose.sandbox.yml
SANDBOX_HEALTH ?= https://localhost:4443/_notary_server/health

VAULT_COMPOSE ?= $(CURDIR)/vault/docker-compose.dev.yml

VERSION ?= $(shell git describe --tags --abbrev 2> /dev/null || echo v0.0.0-dev)
MAJOR ?= $(word 1,$(subst ., ,$(VERSION)))
MINOR ?= $(word 2,$(subst ., ,$(VERSION)))
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

DOCKER_HUB_REPO_WEB := philipssoftware/dctna-web
GHCR_REPO_WEB := ghcr.io/philips-labs/dctna-web
DOCKER_HUB_REPO_SERVER := philipssoftware/dctna-server
GHCR_REPO_SERVER := ghcr.io/philips-labs/dctna-server

.PHONY: help all run build-sandbox clean-dangling-images run-sandbox check-sandbox bootstrap-sandbox sandbox-logs stop-sandbox reset-sandbox download test coverage coverage-out coverage-html build build-static certs dockerize outdated

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

all: build test ## Build and test

run: build ## Run dctna server
	@bin/dctna-server --config .notary/config.json

build-sandbox: ## build Docker images for notary sandbox
	@(cd $(NOTARY_REPO) ; make cross ; docker-compose -f docker-compose.sandbox.yml build)
	@docker-compose build

clean-dangling-images: ## Clean dangling Docker images
	@docker rmi $$(docker images -qf dangling=true)

run-sandbox: ## Run notary sandbox in Docker
	@docker-compose -f $(SANDBOX_COMPOSE) up -d
	@echo
	@echo Too get logs:
	@echo "  make sandbox-logs"
	@echo
	@echo Too enter the sandbox:
	@echo "  docker-compose -f $(SANDBOX_COMPOSE) exec sandbox sh"

run-vault: ## Run hashicorp vault
	@vault/prepare.sh dev

run-dctna: run-sandbox run-vault ## Run dctna server including dependencies
	@docker-compose up -d

check-sandbox: ## Check if the notary sandbox is up and running
	@while [[ "$$(curl --insecure -sLSo /dev/null -w ''%{http_code}'' $(SANDBOX_HEALTH))" != "200" ]]; \
	do echo "Waiting for $(SANDBOX_HEALTH)" && sleep 1; \
	done
	@echo $(SANDBOX_HEALTH)
	@curl -X GET -IL --insecure ${SANDBOX_HEALTH}

bootstrap-sandbox: ## Bootstrap the notary sandbox with some certificates for content trust
	@docker cp bootstrap-sandbox.sh notary_sandbox_1:/root/
	@docker-compose -f $(SANDBOX_COMPOSE) exec sandbox ./bootstrap-sandbox.sh

sandbox-logs: ## Tail the sandbox Docker logs
	@docker-compose -f $(SANDBOX_COMPOSE) logs -f

stop-sandbox: ## Stop the notary sandbox environment
	@echo Shutting down sandbox
	@docker-compose -f $(SANDBOX_COMPOSE) down

stop-vault: ## Stop hashicorp vault
	@echo Shutting down vault
	@docker-compose -f $(VAULT_COMPOSE) down

stop-dctna: ## Stop hashicorp vault
	@echo Shutting down dctna
	@docker-compose down

stop-all: stop-dctna stop-vault stop-sandbox ## Stop all docker containers

reset-sandbox: stop-sandbox ## Reset the Notary sandbox
	@echo Cleaning volumes
	@docker volume rm $$(docker-compose -f $(SANDBOX_COMPOSE) config --volumes | sed 's/^/notary_/g') 2> /dev/null || true
	@docker volume rm $$(docker-compose config --volumes | sed 's/^/dct-notary-admin_/g') 2> /dev/null || true

download: ## Download go dependencies
	@echo Downloading dependencies
	@go mod download

$(GO_PATH)/bin/goimports:
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: lint
lint: $(GO_PATH)/bin/goimports ## runs linting
	@echo Linting imports
	@goimports -d -e -local github.com/philips-labs/dct-notary-admin $(shell go list -f '{{ .Dir }}' ./...)

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

build-static: download ## Build the static binary
	@echo Building binary
	@go build -a -installsuffix cgo ${GO_LDFLAGS_STATIC} -o bin/static/dctna-server ./cmd/dctna-server

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

dockerize-web: ## builds docker images
	docker build -t $(DOCKER_HUB_REPO_WEB) web
	docker rmi $$(docker images -qf dangling=true)

docker-publish-web: ## publishes the image to the hsdp registry
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(DOCKER_HUB_REPO_WEB):$(VERSION)
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(DOCKER_HUB_REPO_WEB):$(MAJOR).$(MINOR)
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(DOCKER_HUB_REPO_WEB):$(MAJOR)
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(GHCR_REPO_WEB):latest
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(GHCR_REPO_WEB):$(VERSION)
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(GHCR_REPO_WEB):$(MAJOR).$(MINOR)
	docker tag $(DOCKER_HUB_REPO_WEB):latest $(GHCR_REPO_WEB):$(MAJOR)
	docker push $(DOCKER_HUB_REPO_WEB):latest
	docker push $(DOCKER_HUB_REPO_WEB):$(VERSION)
	docker push $(DOCKER_HUB_REPO_WEB):$(MAJOR).$(MINOR)
	docker push $(DOCKER_HUB_REPO_WEB):$(MAJOR)
	docker push $(GHCR_REPO_WEB):latest
	docker push $(GHCR_REPO_WEB):$(VERSION)
	docker push $(GHCR_REPO_WEB):$(MAJOR).$(MINOR)
	docker push $(GHCR_REPO_WEB):$(MAJOR)

.PHONY: container-digest
container-digest: ## retrieves the container digest from the given tag
	@:$(call check_defined, GITHUB_REF)
	@docker inspect $(GHCR_REPO_WEB):$(subst refs/tags/,,$(GITHUB_REF)) --format '{{ index .RepoDigests 0 }}' | cut -d '@' -f 2

.PHONY: container-tags
container-tags: ## retrieves the container tags applied to the image with a given digest
	@:$(call check_defined, CONTAINER_DIGEST)
	@docker inspect $(GHCR_REPO_WEB)@$(CONTAINER_DIGEST) --format '{{ join .RepoTags "\n" }}' | sed 's/.*://' | awk '!_[$$0]++'

.PHONY: container-repos
container-repos: ## retrieves the container tags applied to the image with a given digest
	@:$(call check_defined, CONTAINER_DIGEST)
	@docker inspect $(GHCR_REPO_WEB)@$(CONTAINER_DIGEST) --format '{{ join .RepoTags "\n" }}' | sed 's/:.*//' | awk '!_[$$0]++'

outdated: ## Checks for outdated dependencies
	go list -u -m -json all | go-mod-outdated -update
