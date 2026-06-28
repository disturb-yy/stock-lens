# Phase 1 Strategy

Stock Lens Phase 1 is a market-data foundation, not a complete stock analysis product. Its success standard is stable synchronization, storage, and querying of A-share market data.

## Priorities

Phase 1 priorities are:

```text
data correctness > sync recoverability > API stability > performance optimization > analysis capability
```

The primary users are downstream frontends, analysis services, and developers, not end investors. API clarity and data contracts are more important than page-level user experience in Phase 1.

## Scope

Phase 1 implements only the market-data backend foundation:

- stock master data sync
- trade calendar sync
- raw unadjusted daily K line sync
- stock list and detail queries
- daily K line and latest daily K line queries
- sync task and sync log queries
- health checks
- enum metadata

Phase 1 explicitly excludes:

- frontend pages
- login, RBAC, user, account, and tenant systems
- watchlists
- alerts
- technical indicators
- AI analysis
- backtesting
- minute-level data
- real-time market data
- scheduled sync
- distributed task execution
- multi-source arbitration

Do not introduce `domain/analysis`, `domain/strategy`, or `domain/indicator` in Phase 1. Raw data foundation stability comes first.

## Data Source Strategy

Tushare is the Phase 1 real data source, but the domain model must not be coupled to Tushare semantics.

Provider boundaries exist so future sources such as AkShare, JoinQuant, or internal data providers can be added without rewriting market services.

## Storage Strategy

MySQL is the single Phase 1 primary store.

Do not introduce ClickHouse, Elasticsearch, Redis, or other storage systems until concrete query, throughput, search, cache, or analytical pressure justifies them.

## Sync Strategy

Phase 1 sync is manually triggered, single-instance, and globally mutually exclusive.

Recommended evolution path:

```text
manual sync -> scheduled sync -> cancelable tasks -> queue/workers -> distributed scheduling
```

## Market Data Strategy

Phase 1 stores only raw, unadjusted daily K lines.

Future expansion may include:

- adjustment factors
- front-adjusted and back-adjusted views
- minute-level lines
- indexes, ETFs, funds, and other asset types
- real-time market data

## API Strategy

`/api/v1` must remain stable and should not change for internal implementation convenience.

Backward-incompatible public API changes belong in a future `/api/v2`.

## Documentation Strategy

English documentation is the source of truth.

Chinese documentation is a translation mirror and must not be used as the basis for decisions.

When English documentation changes, update the corresponding Chinese translation.

## Test Strategy

Phase 1 uses fast unit tests as the default test suite.

Repository integration tests may exist behind the `integration` build tag. Real Tushare network tests are not part of Phase 1.

## Performance Strategy

Phase 1 focuses on reasonable batch upsert behavior and query indexes.

Do not add cache systems, search engines, or columnar stores before there is concrete pressure.

## Observability Strategy

Phase 1 uses `request_id`, structured `slog`, and `sync_logs`.

OpenTelemetry, metrics stacks, and distributed tracing are deferred until service decomposition or production troubleshooting pressure justifies them.

## Security Strategy

Phase 1 uses Admin Token protection for write-like sync operations.

Query APIs are open by default. If the service becomes public-facing, authentication, authorization, and rate limiting can be introduced later.

## Deployment Strategy

Phase 1 targets a single Go service plus MySQL.

Docker Compose is for local dependencies only. Do not design Kubernetes or distributed deployment topology before it is needed.

## Domain Boundary Strategy

Phase 1 implements only `domain/market`.

New domains should be created only when real use cases and independent domain language appear. Do not create domains from technical speculation.

## Deferred Capability Triggers

Deferred capabilities require concrete triggers:

```text
scheduled sync: unattended daily refresh is needed
Redis: hot queries create measurable MySQL pressure
ClickHouse: historical K line analysis slows OLTP workloads
user system: multi-user permission requirements appear
indicator domain: explicit technical indicator APIs or calculations are required
```

## Done Criteria

Phase 1 is complete when the system can:

- sync stock master data
- sync trade calendars
- sync daily K lines
- query stock lists and stock details
- query daily K lines and latest daily K lines
- query sync tasks and sync logs
- expose stable health checks
- expose enum metadata
- maintain stable error codes, API documentation, and table structure documentation
- run basic tests and CI
