SHELL := /bin/bash

GO ?= go
VERSION ?= 0.0.0-dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/Exonical/kubevirt-management/internal/version.Version=$(VERSION) \
	-X github.com/Exonical/kubevirt-management/internal/version.Commit=$(COMMIT) \
	-X github.com/Exonical/kubevirt-management/internal/version.BuildDate=$(BUILD_DATE)

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_.-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

GO_PKGS := ./cmd/... ./internal/...

.PHONY: vet
vet: ## go vet
	$(GO) vet $(GO_PKGS)

.PHONY: test
test: ## go test
	$(GO) test $(GO_PKGS) -race -count=1

.PHONY: build
build: web ## Build the server binary with the embedded SPA
	rm -rf internal/static/dist && cp -r web/dist internal/static/dist
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o bin/kubevirt-management ./cmd/server

.PHONY: run
run: ## Run the server (assumes web/dist already built or use placeholder)
	$(GO) run ./cmd/server

.PHONY: web
web: ## Build the SPA
	cd web && npm ci && npm run build

.PHONY: web-dev
web-dev: ## Run the SPA dev server (Vite)
	cd web && npm run dev

.PHONY: web-lint
web-lint: ## Lint and typecheck the SPA
	cd web && npm run lint && npm run typecheck

.PHONY: docker
docker: ## Build the container image
	docker build -t kubevirt-management:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) .

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint deploy/helm/kubevirt-management

.PHONY: dev-keys
dev-keys: ## Print a working set of dev session/OIDC env vars
	@echo "export SESSION_HASH_KEY=$$(openssl rand -hex 64)"
	@echo "export SESSION_BLOCK_KEY=$$(openssl rand -hex 32)"
	@echo "export SESSION_SECURE=false"
	@echo "export OIDC_ISSUER=http://localhost:5556/dex"
	@echo "export OIDC_CLIENT_ID=kubevirt-management"
	@echo "export OIDC_CLIENT_SECRET=local-dev-secret"
	@echo "export OIDC_REDIRECT_URL=http://localhost:8080/api/auth/callback"
	@echo "export AUTH_MODE=impersonation"
