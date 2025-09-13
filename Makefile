SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

MKCERT_INSTALL_PATH ?= $(HOME)/.local/bin
MKCERT_VERSION := v1.4.4
MKCERT_ASSET := mkcert-$(MKCERT_VERSION)-linux-amd64
MKCERT_URL := https://github.com/FiloSottile/mkcert/releases/latest/download/$(MKCERT_ASSET)
CERT_DIR ?= ./certs
HOSTS ?= localhost 127.0.0.1 ::1


.PHONY: setup
setup: install-mkcert mkcert-generate mkcert-generate mkcert-install-ca


.PHONY: check
check: ## Run linters
	@go tool golangci-lint run

.PHONY: fix
fix: ## Auto-fix lint issues where possible
	@go tool golangci-lint run --fix

.PHONY: fieldalign
fieldalign: ## Apply field alignment fixes
	@go tool fieldalignment -fix ./pkg/...

.PHONY: fmt
fmt: ## Run project formatting helpers
	@go tool golangci-lint fmt

.PHONY: tidy
tidy: ## Tidy go.mod and download modules
	go mod tidy
	go mod download

.PHONY: test
test: ## Run unit tests with coverage (JSON output piped through gotestfmt)
	@go test -covermode=atomic -gcflags='all=-N -l' -tags testing -coverprofile=coverage.txt -timeout 5m -json -v ./... 2>&1 | go tool gotestfmt -showteststatus

.PHONY: security
security: ## Run basic security checks (gosec)
	go tool gosec ./...


# --- Mkcert  -------------------------------
install-mkcert: ## Install mkcert binary to $(MKCERT_INSTALL_PATH) if missing
	@if [ ! -f "$(MKCERT_INSTALL_PATH)/mkcert" ]; then \
	  mkdir -p $(MKCERT_INSTALL_PATH) 2>/dev/null || true; \
	  echo "Downloading mkcert from $(MKCERT_URL)..."; \
	  tmp=$$(mktemp -d); \
	  curl -L --fail -o $$tmp/$(MKCERT_ASSET) "$(MKCERT_URL)"; \
	  chmod +x $$tmp/$(MKCERT_ASSET); \
	  mv $$tmp/$(MKCERT_ASSET) $(MKCERT_INSTALL_PATH)/mkcert || { echo "Failed to move mkcert to $(MKCERT_INSTALL_PATH)"; exit 1; }; \
	  rm -rf $$tmp; \
	  echo "mkcert installed to $(MKCERT_INSTALL_PATH)/mkcert"; \
	fi

mkcert-install-ca: install-mkcert ## Install mkcert CA into system trust stores
	@echo "Installing mkcert CA..."
	@$(MKCERT_INSTALL_PATH)/mkcert -install

mkcert-generate: mkcert-install-ca ## Generate cert/key for HOSTS into $(CERT_DIR) (usage: make mkcert-generate HOSTS='example.test localhost')
	@mkdir -p $(CERT_DIR)
	@echo "Generating cert for: $(HOSTS)"
	@$(MKCERT_INSTALL_PATH)/mkcert -cert-file $(CERT_DIR)/smtp.crt -key-file $(CERT_DIR)/smtp.key $(HOSTS)
	@echo "Created $(CERT_DIR)/smtp.crt and $(CERT_DIR)/smtp.key"
	@$(MKCERT_INSTALL_PATH)/mkcert -cert-file $(CERT_DIR)/local.crt -key-file $(CERT_DIR)/local.key $(HOSTS)
	@echo "Created $(CERT_DIR)/local.crt and $(CERT_DIR)/local.key"

mkcert-uninstall-ca: install-mkcert ## Uninstall mkcert CA from system trust stores
	@echo "Uninstalling mkcert CA..."
	@$(MKCERT_INSTALL_PATH)/mkcert -uninstall || echo "mkcert -uninstall failed (maybe not installed)"


.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/
	go clean

.PHONY: help
help:
	@echo 'UTILS - Available commands:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
