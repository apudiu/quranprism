# apps/api/Makefile

# Tooling
GO ?= go
GOLANGCI_LINT ?= golangci-lint
SQLC ?= sqlc
GOOSE ?= goose

# Database
DATABASE_URL ?= postgres://qp:qp_dev@localhost:5432/qp?sslmode=disable

.PHONY: help
help:
	@echo "Common targets:"
	@echo "  make run-api         - run the HTTP API server"
	@echo "  make run-worker      - run the background job worker"
	@echo "  make run-cron        - run the scheduler"
	@echo "  make build           - build all three binaries to ./bin/"
	@echo "  make test            - run unit tests"
	@echo "  make test-integration - run integration tests (requires running Postgres)"
	@echo "  make lint            - run golangci-lint"
	@echo "  make sqlc            - regenerate sqlc code from queries"
	@echo "  make migrate-up      - apply pending app migrations"
	@echo "  make migrate-down    - roll back the last app migration"
	@echo "  make river-migrate-up - apply River internal migrations"

.PHONY: run-api
run-api:
	$(GO) run ./cmd/api

.PHONY: run-worker
run-worker:
	$(GO) run ./cmd/worker

.PHONY: run-cron
run-cron:
	$(GO) run ./cmd/cron

.PHONY: build
build:
	mkdir -p ./bin
	$(GO) build -o ./bin/api ./cmd/api
	$(GO) build -o ./bin/worker ./cmd/worker
	$(GO) build -o ./bin/cron ./cmd/cron

.PHONY: test
test:
	$(GO) test -race -short ./...

.PHONY: test-integration
test-integration:
	$(GO) test -race ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: sqlc
sqlc:
	$(SQLC) generate

.PHONY: migrate-up
migrate-up:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

.PHONY: migrate-status
migrate-status:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" status

.PHONY: river-migrate-up
river-migrate-up:
	$(GO) run ./cmd/river-migrate up
