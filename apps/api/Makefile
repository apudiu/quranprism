# apps/api/Makefile
#
# The whole CLI is the single `qp` binary (`cmd/qp`). Run any subcommand
# via `go run ./cmd/qp <subcommand>` in dev or the built binary in prod.

# Tooling
GO ?= go
GOLANGCI_LINT ?= golangci-lint
SQLC ?= sqlc

# Database (used by tooling that talks directly to postgres outside the binary)
DATABASE_URL ?= postgres://qp:qp_dev@localhost:5432/qp?sslmode=disable

.PHONY: help
help:
	@echo "Common targets:"
	@echo "  make run-api          - run the HTTP API (qp serve:api)"
	@echo "  make run-worker       - run the worker (qp serve:worker)"
	@echo "  make run-cron         - run the scheduler (qp serve:cron)"
	@echo "  make build            - build the qp binary to ./bin/qp"
	@echo "  make test             - run unit tests"
	@echo "  make test-integration - run integration tests (requires running Postgres)"
	@echo "  make lint             - run golangci-lint"
	@echo "  make sqlc             - regenerate sqlc code from queries"
	@echo "  make migrate-up       - apply pending goose migrations"
	@echo "  make migrate-down     - roll back the last goose migration"
	@echo "  make migrate-status   - list goose migration state"
	@echo "  make river-migrate-up - apply River's internal migrations (qp migrate:queue)"

.PHONY: run-api
run-api:
	$(GO) run ./cmd/qp serve:api

.PHONY: run-worker
run-worker:
	$(GO) run ./cmd/qp serve:worker

.PHONY: run-cron
run-cron:
	$(GO) run ./cmd/qp serve:cron

.PHONY: build
build:
	mkdir -p ./bin
	$(GO) build -o ./bin/qp ./cmd/qp

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
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/qp migrate:up

.PHONY: migrate-down
migrate-down:
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/qp migrate:down

.PHONY: migrate-status
migrate-status:
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/qp migrate:status

.PHONY: river-migrate-up
river-migrate-up:
	DATABASE_URL=$(DATABASE_URL) $(GO) run ./cmd/qp migrate:queue
