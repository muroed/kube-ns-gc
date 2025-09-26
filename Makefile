# Makefile for kube-ns-gc

# Variables
APP_NAME := kube-ns-gc
VERSION := $(shell git describe --tags --always --dirty)
REGISTRY := ghcr.io
IMAGE_NAME := $(REGISTRY)/$(APP_NAME)

# Go variables
GO_VERSION := 1.24
GOOS := linux
GOARCH := amd64

# Docker variables
DOCKER_BUILDX := docker buildx build --platform $(GOOS)/$(GOARCH)

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the Go binary
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -installsuffix cgo -o $(APP_NAME) ./src

.PHONY: test
test: ## Run tests
	go test -v ./src/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./src/...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./src/...

.PHONY: vet
vet: ## Run go vet
	go vet ./src/...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	go mod tidy

.PHONY: docker-build
docker-build: ## Build Docker image
	$(DOCKER_BUILDX) -t $(IMAGE_NAME):$(VERSION) .
	$(DOCKER_BUILDX) -t $(IMAGE_NAME):latest .

.PHONY: docker-push
docker-push: ## Push Docker image
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

.PHONY: helm-package
helm-package: ## Package Helm chart
	helm package deploy/$(APP_NAME) --destination ./packages

.PHONY: helm-lint
helm-lint: ## Lint Helm chart
	helm lint deploy/$(APP_NAME)

.PHONY: helm-template
helm-template: ## Template Helm chart
	helm template $(APP_NAME) deploy/$(APP_NAME)

.PHONY: helm-install
helm-install: ## Install Helm chart locally
	helm upgrade --install $(APP_NAME) deploy/$(APP_NAME) --namespace $(APP_NAME) --create-namespace

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm chart
	helm uninstall $(APP_NAME) --namespace $(APP_NAME)

.PHONY: clean
clean: ## Clean build artifacts
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	rm -rf packages/

.PHONY: dev
dev: ## Run in development mode
	go run ./src

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: check
check: fmt vet lint test ## Run all checks

.PHONY: all
all: clean check build docker-build helm-package ## Run all build steps
