# Market Domain Package Structure

Phase 1 uses one Go package for the market domain:

```text
domain/market/
  model.go
  port.go
  service.go
  handler.go
  dto.go
  errors.go
  repository_mysql.go
  repository_mock.go
  provider_tushare.go
  provider_mock.go
  INDEX.md
```

## File Responsibilities

`model.go` contains market domain models, enums, constants, and small domain predicates. It does not contain HTTP DTOs, GORM structs, or Tushare raw protocol structs.

`port.go` contains the interfaces and query/request structs used by domain services, including repositories, providers, and transaction management. It does not contain external protocol types.

`service.go` contains use-case orchestration through `QueryService` and `SyncService`. It does not directly depend on Gin, GORM, or the raw Tushare client.

Public service methods correspond to API use cases. `handler.go` should not compose business workflows from lower-level operations; it should parse HTTP input, call one service method, and map the result to an HTTP response.

`handler.go` contains the Gin HTTP boundary: route binding, request parsing, service calls, and HTTP response/error mapping. It does not contain sync business logic or direct repository/provider access.

`dto.go` contains HTTP request and response DTOs and DTO conversion helpers. Decimal market values are exposed as strings, and market-data units are exposed as response fields.

`errors.go` contains market domain errors, error codes, and helpers for classifying domain errors.

`repository_mysql.go` contains MySQL/GORM repository implementations, table structs, upsert logic, query logic, and transaction handling. GORM types stay private to this file.

`repository_mock.go` contains in-memory repository implementations for development and tests. Mock behavior should stay close to the MySQL implementation for upsert, pagination, filtering, and unique-key semantics.

`provider_tushare.go` contains Tushare provider adapters, including status conversion, `ts_code` to `symbol` conversion, board inference, date/decimal/unit mapping, and provider error wrapping. It may call the raw `pkg/tushare` client.

`provider_mock.go` contains mock provider adapters for local development, unit tests, and no-network test flows.

`INDEX.md` contains a short package guide: purpose, file map, main entry points, and boundary rules. It should not duplicate the full API or table specifications.

## Repository Interface Split

Repository interfaces stay split by data responsibility:

```text
StockRepository
KLineRepository
TradeCalendarRepository
SyncTaskRepository
```

Phase 1 does not merge these into a single `MarketRepository`. Split interfaces keep `QueryService` and `SyncService` dependencies explicit, make tests easier to scope, and avoid a broad interface whose methods are not used together.

## Boundary Rules

The market domain keeps external implementation details at the edge:

- Gin is limited to `handler.go`.
- HTTP DTOs are limited to `dto.go`.
- GORM is limited to `repository_mysql.go`.
- Raw Tushare protocol handling is limited to `pkg/tushare` and `provider_tushare.go`.
- Domain services depend on interfaces in `port.go`, not on concrete infrastructure packages.

## Naming Rules

Go identifiers use `CamelCase` or `camelCase`. JSON fields, HTTP query/body parameters, and SQL identifiers use `snake_case`. Enum values remain uppercase English tokens such as `PARTIAL_SUCCESS`.

When a Go DTO field represents an API field, use Go naming for the field and a `snake_case` JSON tag, for example `TaskUID string `json:"task_uid"``.

## Interface Conventions

Provider interfaces stay split by external capability:

```text
InstrumentProvider
MarketDataProvider
CalendarProvider
```

They are not merged into a single `MarketProvider` in Phase 1.

Repositories accept and return domain models, not GORM records or HTTP DTOs. Providers return domain models, not raw Tushare protocol structs. Repository query structs and provider request structs belong in `port.go`.

Handlers do not pass Gin context or HTTP DTOs into services. Services return market domain errors, and handlers map those errors to HTTP status codes and response bodies.

Concrete repository implementations, provider implementations, and GORM record structs should stay private. Constructors expose interfaces such as `StockRepository`, `MarketDataProvider`, or `TxManager`.

Mock repositories and providers should also use constructors. Tests may seed mock data through options or explicit methods, not by mutating public internal maps.

`TxManager` exposes only `WithTx(ctx, fn)`. `TxRepositories` exposes repository interfaces only and never a raw database handle. `QueryService` does not depend on `TxManager`; sync write workflows may use it through `SyncService`.

`Handler` holds services only. Route registration for market endpoints lives in the market package, while the outer router owns the `/api/v1` prefix. Admin Token checks belong in middleware and protected route groups, not in `SyncService`.

Enum responses are generated from domain constants, not database dictionary tables. DTO conversion is one-way from domain model to HTTP response; domain models do not depend on DTOs.
