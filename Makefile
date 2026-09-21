SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

.PHONY: help doctor bootstrap format format-check lint contracts unit integration e2e test build smoke smoke-backend smoke-frontend generate generate-check migrate-up migrate-down migrate-version dev down check ci

help: ## Show the stable repository command surface.
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

doctor: ## Validate required local toolchains and print actionable failures.
	@./scripts/doctor.sh

bootstrap: doctor ## Install locked Go and JavaScript dependencies.
	@cd backend && go mod download
	@pnpm install --frozen-lockfile

format: ## Format all authored backend and frontend files.
	@cd backend && gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')
	@pnpm --dir frontend format

format-check: ## Check formatting without modifying files.
	@./scripts/check-go-format.sh
	@pnpm --dir frontend format:check

lint: ## Run backend and frontend static analysis.
	@cd backend && go vet ./...
	@pnpm --dir frontend lint

contracts: ## Validate OpenAPI and JSON Schema contracts.
	@pnpm contracts:lint

unit: ## Run deterministic unit tests.
	@cd backend && go test ./...
	@pnpm --dir frontend test

integration: ## Run integration tests against isolated dependencies.
	@cd backend && go test -tags=integration ./...

e2e: ## Run critical browser journeys against built applications.
	@pnpm --dir frontend test:e2e

test: unit integration ## Run unit and integration tests.

build: ## Build production backend and frontend artifacts.
	@mkdir -p bin
	@cd backend && go build -trimpath -o ../bin/manoreck-api ./cmd/api
	@pnpm --dir frontend build

smoke: smoke-backend smoke-frontend ## Start each built application and probe its public surface.

smoke-backend: ## Start the backend process and probe its health endpoint.
	@./scripts/smoke-backend.sh

smoke-frontend: ## Start the frontend process and probe its shell.
	@./scripts/smoke-frontend.sh

generate: ## Regenerate contract-derived and database code.
	@cd backend && go generate ./...
	@pnpm --dir frontend generate

generate-check: generate ## Fail when committed generated files are stale.
	@git diff --exit-code -- contracts backend frontend

migrate-up: ## Apply all pending local database migrations.
	@docker compose --env-file .env -f deploy/compose.yaml run --rm migrate up

migrate-down: ## Revert one local database migration.
	@docker compose --env-file .env -f deploy/compose.yaml run --rm migrate down 1

migrate-version: ## Print the current local database migration version.
	@docker compose --env-file .env -f deploy/compose.yaml run --rm migrate version

dev: ## Build and start the complete local stack.
	@docker compose --env-file .env -f deploy/compose.yaml up --build --wait

down: ## Stop the local stack without deleting persisted volumes.
	@docker compose --env-file .env -f deploy/compose.yaml down

check: format-check lint contracts unit build ## Run fast local quality checks.

ci: check integration generate-check ## Run the non-browser continuous-integration gate.
