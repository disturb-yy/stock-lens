# Market Domain

`domain/market` is the Phase 1 market-data domain package. It owns stock master data, raw daily K lines, trade calendars, sync tasks, and sync logs.

## Files

```text
model.go              domain models, enums, constants, and small predicates
port.go               repository, provider, transaction interfaces, and service query/request structs
service.go            QueryService and SyncService use-case orchestration
handler.go            Gin HTTP boundary and route handlers
dto.go                HTTP request/response DTOs and conversion helpers
errors.go             market domain errors and error codes
repository_mysql.go   MySQL/GORM repository and transaction implementations
repository_mock.go    in-memory repository implementations for development and tests
provider_tushare.go   Tushare provider adapters and provider-to-domain mapping
provider_mock.go      mock provider adapters for development and tests
INDEX.md              package guide and boundary rules
```

## Entry Points

- `Handler` is the HTTP entry point.
- `QueryService` handles read-side market queries.
- `SyncService` creates and runs market-data sync tasks.

Public `QueryService` and `SyncService` methods correspond to API use cases. Handlers parse HTTP input and map responses, but they do not compose business workflows from lower-level domain operations.

## Boundary Rules

- Keep this as one Go package in Phase 1.
- Separate responsibilities by file, not by subpackage.
- Keep Gin usage in `handler.go`.
- Keep HTTP DTOs in `dto.go`.
- Keep GORM usage in `repository_mysql.go`.
- Keep raw Tushare protocol handling in `pkg/tushare` and `provider_tushare.go`.
- Keep services dependent on `port.go` interfaces, not concrete infrastructure types.

## Naming Rules

- Go identifiers use `CamelCase` or `camelCase`, such as `TaskUID`, `PageSize`, and `LatestDailyKLine`.
- JSON fields, HTTP query/body parameters, and SQL identifiers use `snake_case`, such as `task_uid`, `page_size`, and `latest_daily_k_line`.
- Enum values remain uppercase English tokens, such as `CN`, `STOCK`, `SYNC_DAILY_K_LINES`, and `PARTIAL_SUCCESS`.
- Go DTO fields use Go naming plus `snake_case` JSON tags.

## Repository Interfaces

Repository interfaces stay split by data responsibility:

- `StockRepository`
- `KLineRepository`
- `TradeCalendarRepository`
- `SyncTaskRepository`

Do not merge them into a single `MarketRepository` in Phase 1. Split interfaces keep service dependencies explicit while still preserving the flat `domain/market` package shape.

## Interface And Boundary Conventions

- Keep provider interfaces split as `InstrumentProvider`, `MarketDataProvider`, and `CalendarProvider`.
- Repositories accept and return domain models, not GORM records or HTTP DTOs.
- Providers return domain models, not raw Tushare protocol structs.
- Put repository query structs such as `ListStocksQuery` and `FindStockQuery` in `port.go`.
- Put provider request structs such as `FetchStocksRequest` and `FetchDailyKLinesRequest` in `port.go`.
- Do not pass Gin context or HTTP DTOs into services.
- Define service input structs for complex use cases; keep them close to the service methods that use them.
- Services return market domain errors; handlers map them to HTTP status codes and response bodies.
- Keep GORM record structs private in `repository_mysql.go`.
- Keep concrete repository implementations private and expose constructors returning interfaces.
- Prefer private provider implementation structs with constructors returning provider interfaces.
- Provide constructors for mock repositories and providers; do not expose internal maps as public state.
- Keep `TxManager` limited to `WithTx(ctx, fn)`.
- `TxRepositories` exposes repository interfaces only, not a raw database handle.
- `QueryService` does not depend on `TxManager`; `SyncService` may use it for sync writes.
- `Handler` holds services only, not repositories or providers.
- Register market routes from the market package, while the outer router owns the `/api/v1` prefix.
- Keep Admin Token checks in middleware and protected route groups, not in `SyncService`.
- Generate enum responses from domain constants, not database dictionary tables.
- Keep DTO conversion one-way from domain model to HTTP response; domain models do not depend on DTOs.

## Sync Task Rules

- Only one `PENDING` or `RUNNING` sync task may exist at a time in Phase 1.
- Sync creation persists `PENDING` and returns immediately; the background goroutine marks the task `RUNNING`.
- Terminal task states are `SUCCESS`, `FAILED`, and `PARTIAL_SUCCESS`.
- Invalid request input is rejected before task creation.
- Missing sync prerequisites fail the created task.
- Full-market daily K sync can finish as `PARTIAL_SUCCESS`; stock master sync, trade-calendar sync, and single-stock daily K sync either succeed or fail.
- Full-market daily K progress counts stocks, not K line rows.
- Sync logs are task-level and stock-level summaries, not per-K-line records.

## Error Mapping

- Market argument errors map to HTTP `400`.
- Missing resources map to HTTP `404`.
- Missing local prerequisites and active task conflicts map to HTTP `409`.
- Provider failures map to HTTP `502`.
- Store and unclassified task failures map to HTTP `500`.
- Invalid arguments, active task conflicts, and Admin Token failures do not create sync tasks.
- Missing prerequisites discovered during sync execution create a failed sync task.

## API Shape Rules

- Success responses use `code = "OK"` and `message = "ok"`.
- Error messages are stable English phrases; clients use error codes for logic.
- API responses never expose internal database auto-increment IDs.
- Stocks are identified by `symbol`; sync tasks are identified by `task_uid`.
- Date fields use `YYYY-MM-DD`; datetime fields use ISO 8601 with timezone.
- Daily K line API fields use `open`, `high`, `low`, `close`, and `change`; domain and database fields keep explicit price/amount names.
- HTTP market numeric fields are strings.
- Daily K line responses include explicit unit fields for percentage change, volume, and amount.
- Stock list items do not include latest daily K line; stock detail includes `latest_daily_k_line`.
- The latest daily K line endpoint remains available as a dedicated read endpoint.
- Enum API responses come from domain constants and do not include multilingual labels in Phase 1.

## Calendar And Date Rules

- Phase 1 defaults to `market=CN` and `exchange=SSE` for trade-calendar behavior.
- Latest open day is derived from the maximum open `cal_date` in `trade_calendars` for `market + exchange`.
- Latest open day is distinct from today and from a stock's latest local daily K line date.
- Daily K sync defaults missing `end_date` to latest open day.
- Daily K query defaults missing `end_date` to the stock's latest local `trade_date`; if no local daily K line exists, it returns an empty array.
- Service business timezone is `Asia/Shanghai`; date boundaries and `today` use that timezone.
- Phase 1 does not dynamically choose a trade calendar by stock exchange for daily K sync.

## Provider Rules

- Raw Tushare protocol handling lives in `pkg/tushare`; market adaptation lives in `provider_tushare.go`.
- Providers return market domain models, not raw Tushare response structures.
- Tushare provider output uses `DataSource = TUSHARE`; mock provider output uses `DataSource = MOCK`.
- `stocks.ts_code` stores Tushare identity; daily K lines do not store `ts_code`.
- Tushare `ts_code` is mapped to system `symbol`; invalid provider identifiers are mapping errors.
- Tushare numeric values are parsed as `decimal.Decimal`; invalid numeric fields are mapping errors.
- Provider logs must not include Tushare tokens or complete token-bearing requests.
- Phase 1 does not run real Tushare network tests.

## Strategic Context

Phase 1 is a market-data foundation, not a complete stock analysis product. The market domain should prioritize data correctness, sync recoverability, and API stability before analysis features or infrastructure expansion.

## Runtime And Test Defaults

- Default HTTP server port is `30078`.
- `/healthz` means the process is alive.
- `/readyz` means the service is ready, including required startup validation and MySQL connectivity.
- Unit tests should prefer table-driven style and the smallest useful test scope.
- `gomonkey` may be used in unit tests when interface substitution is not practical.
- Tests that rely on monkey patching must consider Go inlining and may need `go test ./... -gcflags=all="-N -l"`.
