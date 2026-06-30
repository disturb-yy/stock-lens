GO ?= go
GOOSE_VERSION ?= v3.26.0
GOOSE := $(GO) run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
MIGRATIONS_DIR ?= migrations
MYSQL_DSN ?=

.PHONY: run test test-integration vet fmt-check dev-up dev-down migrate-up migrate-down migrate-status

run:
	$(GO) run ./cmd/server

test:
	GOCACHE=/tmp/stock-lens-go-cache $(GO) test ./... -count=1

test-integration:
	GOCACHE=/tmp/stock-lens-go-cache $(GO) test -tags=integration ./... -count=1

vet:
	$(GO) vet ./...

fmt-check:
	test -z "$$($(GO)fmt -l $$(find . -name '*.go' -not -path './.git/*'))"

dev-up:
	docker compose up -d mysql

dev-down:
	docker compose down

migrate-up:
	@test -n "$(MYSQL_DSN)" || (echo "MYSQL_DSN is required" && exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(MYSQL_DSN)" up

migrate-down:
	@test -n "$(MYSQL_DSN)" || (echo "MYSQL_DSN is required" && exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(MYSQL_DSN)" down

migrate-status:
	@test -n "$(MYSQL_DSN)" || (echo "MYSQL_DSN is required" && exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(MYSQL_DSN)" status
