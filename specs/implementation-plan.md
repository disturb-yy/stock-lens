# Implementation Plan

This plan sequences Phase 1 implementation from infrastructure foundations to market-domain behavior and API documentation.

## Principles

- Keep Phase 1 focused on `domain/market`.
- Build in thin vertical slices, but establish shared infrastructure first.
- Prefer mock provider/repository flows before real Tushare integration.
- Keep tests close to each layer and use table-driven unit tests by default.
- Keep English documentation as the source of truth and update Chinese mirrors when English docs change.

## Milestone 1: Project Skeleton And Runtime

Deliver:

- `go.mod` using Go 1.26
- command entry point under `cmd/server`
- explicit config loader with YAML and environment expansion
- config validation for server, database, auth, market, Tushare, and log settings
- `slog` logger setup
- HTTP router with `/healthz` and `/readyz`
- default server port `30078`
- graceful shutdown on `SIGINT` and `SIGTERM`

Tests:

- table-driven config loading and validation tests
- health/readiness handler tests

## Milestone 2: Shared HTTP And Middleware

Deliver:

- uniform response helper with `code`, `message`, `request_id`, and `data`
- request ID middleware
- access log middleware
- Admin Token middleware using `Authorization: Bearer <token>`
- error mapping helpers for common and market errors

Tests:

- response shape tests
- request ID propagation tests
- Admin Token middleware tests
- error mapping table tests

## Milestone 3: Market Domain Model And Ports

Deliver:

- `domain/market/model.go`
- `domain/market/port.go`
- `domain/market/errors.go`
- domain constants and enums
- repository interfaces
- provider interfaces
- transaction manager interface
- service input/query structs

Tests:

- enum validation tests
- symbol/date/range validation tests
- domain error classification tests

## Milestone 4: Mock Repositories And Providers

Deliver:

- in-memory mock repositories
- mock providers with `DataSource = MOCK`
- controllable failure scenarios for sync task testing
- upsert-like behavior for mock repositories
- pagination and filtering behavior close to MySQL semantics

Tests:

- repository upsert and query table tests
- mock provider success, empty-result, and failure tests

## Milestone 5: Query Service

Deliver:

- stock list query
- stock detail query with `latest_daily_k_line`
- daily K line list query
- latest daily K line query
- trade calendar list query
- latest open day query
- is-open query
- enum metadata query

Tests:

- table-driven service tests using mock repositories
- empty-result and not-found behavior tests
- default date range tests

## Milestone 6: Sync Service

Deliver:

- sync task creation and global active-task conflict check
- background goroutine execution
- stale task recovery on startup
- sync stock master data
- sync trade calendars
- sync daily K lines, single-stock and full-market
- task status transitions
- progress counters
- sync logs
- panic recovery to failed task

Tests:

- task conflict tests
- status transition tests
- prerequisite failure tests
- partial-success tests for full-market daily K sync
- panic recovery tests
- gomonkey may be used only where interface substitution is not practical; no-inline test command must be considered for those tests

## Milestone 7: HTTP Handlers And Routes

Deliver:

- market handler and route registration
- read endpoints under `/api/v1/market`
- Admin Token protected sync endpoints
- DTO conversion with snake_case JSON tags
- pagination response shape
- date/datetime formatting
- error-to-HTTP mapping

Tests:

- handler tests for success and error responses
- DTO formatting tests
- Admin-protected route tests

## Milestone 8: MySQL Persistence

Deliver:

- SQL migrations under `migrations/`
- MySQL repository implementations using GORM only inside repositories
- MySQL transaction manager
- batch upsert behavior
- indexes matching the table structure spec
- Makefile migration targets

Tests:

- repository integration tests behind the `integration` build tag
- migration smoke checks where practical

## Milestone 9: Tushare Integration

Deliver:

- raw `pkg/tushare` client
- explicit Tushare `fields` requests
- Tushare provider adapters in `domain/market/provider_tushare.go`
- `ts_code` to `symbol` mapping
- exchange, board, status, date, decimal, and unit mapping
- provider error wrapping without logging tokens

Tests:

- raw response parsing tests using constructed responses
- provider mapping table tests
- invalid mapping error tests
- no real Tushare network tests

## Milestone 10: API Contract And Local Developer Experience

Deliver:

- `specs/openapi.yaml`
- human API guide if needed
- `configs/config.yaml`
- `configs/config.test.yaml`
- `docker-compose.yml` for local MySQL
- `Makefile` targets for run, test, integration test, migration up/down, and formatting
- README updates
- minimal CI with formatting, vet, and unit tests

Tests:

- `go test ./...`
- `go vet ./...`
- formatting check

## Recommended Build Order

```text
runtime/config/logger/router
-> HTTP response and middleware
-> market models/errors/ports
-> mock repositories/providers
-> query service
-> sync service
-> handlers/routes
-> MySQL repositories/migrations
-> Tushare raw client/provider
-> OpenAPI/docs/CI polish
```
