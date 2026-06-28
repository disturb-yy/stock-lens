# Stock Lens

Stock Lens is a Phase 1 market-data backend foundation for A-share stock data. It synchronizes stock master data, trade calendars, and raw unadjusted daily K lines, then exposes stable query and sync-task APIs.

## Current Scope

Phase 1 focuses on:

- A-share stock master data sync
- A-share trade calendar sync
- raw unadjusted daily K line sync
- stock list and stock detail queries
- daily K line and latest daily K line queries
- sync task and sync log queries
- health checks and enum metadata

Phase 1 does not include frontend pages, user login, RBAC, watchlists, alerts, technical indicators, AI analysis, backtesting, real-time data, scheduled sync, or distributed task execution.

## Runtime Defaults

Default local server:

```text
http://localhost:30078
```

Health endpoints:

```text
GET /healthz
GET /readyz
```

`/healthz` means the HTTP process is alive. `/readyz` means the service is ready, including required startup validation and MySQL connectivity.

API prefix:

```text
/api/v1
```

Market API prefix:

```text
/api/v1/market
```

## Configuration

The implementation should use explicit configuration for:

- server address and port, defaulting to `30078`
- business timezone, defaulting to `Asia/Shanghai`
- MySQL connection settings
- Admin Token for sync APIs
- market provider, `mock` or `tushare`
- Tushare token and base URL when using the Tushare provider
- sync batch size
- log level

Mock provider mode can run without a Tushare token. Tushare provider mode requires a Tushare token.

## Main APIs

Read APIs:

```text
GET /api/v1/market/stocks
GET /api/v1/market/stocks/{symbol}
GET /api/v1/market/stocks/{symbol}/daily-k-lines
GET /api/v1/market/stocks/{symbol}/latest-daily-k-line
GET /api/v1/market/trade-calendars
GET /api/v1/market/trade-calendars/latest-open-day
GET /api/v1/market/trade-calendars/is-open
GET /api/v1/market/meta/enums
```

Sync APIs require Admin Token:

```text
POST /api/v1/market/sync/stocks
POST /api/v1/market/sync/trade-calendars
POST /api/v1/market/sync/daily-k-lines
GET  /api/v1/market/sync/tasks/{task_uid}
GET  /api/v1/market/sync/tasks/{task_uid}/logs
```

## Testing

Use fast unit tests as the default test suite:

```sh
go test ./...
```

Unit tests should prefer table-driven style and the smallest useful test scope.

Tests that rely on `gomonkey` may need inlining disabled:

```sh
go test ./... -gcflags=all="-N -l"
```

Do not disable inlining for default test runs unless the specific test flow requires monkey patching.

Repository integration tests, when added, should use the `integration` build tag.

## Documentation

English documentation is the source of truth. Chinese documents under `zh/` directories are translation mirrors and should be updated whenever the corresponding English document changes.

Key documents:

- `specs/phase1-strategy.md`
- `specs/api-contract-decisions.md`
- `specs/phase1-table-structure.md`
- `specs/sync-task-behavior.md`
- `specs/provider-tushare-adapter.md`
- `specs/runtime-and-testing.md`
- `specs/implementation-plan.md`
- `docs/git-conventions.md`
- `domain/market/INDEX.md`
