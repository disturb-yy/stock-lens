# Stock Lens Model Index

This file is the model-facing project entry point. Models and automation agents should read this file instead of the root `README.md`.

## Read Rules

- Do not read the root `README.md`; it is user-facing documentation.
- Use this `INDEX.md` for project orientation.
- When this `INDEX.md` changes, update the root `README.md` in the same change so the user-facing documentation stays aligned.
- English documentation remains the source of truth for technical decisions unless a project rule explicitly says otherwise.
- Chinese documentation under `zh/` directories is a translation mirror and should not be used as decision context.

## Project Purpose

Stock Lens Phase 1 is a market-data backend foundation for A-share stock data. It synchronizes stock master data, trade calendars, and raw unadjusted daily K lines, then exposes stable query and sync-task APIs.

Phase 1 is not a complete stock analysis product. Prioritize data correctness, sync recoverability, and API stability before analysis features or infrastructure expansion.

## Runtime Defaults

- Default server port: `30078`
- Business timezone: `Asia/Shanghai`
- Health endpoint: `GET /healthz`
- Readiness endpoint: `GET /readyz`
- API prefix: `/api/v1`
- Market API prefix: `/api/v1/market`

`/healthz` means the HTTP process is alive. `/readyz` means the service is ready, including required startup validation and MySQL connectivity.

## Key Phase 1 Scope

Included:

- stock master data sync
- trade calendar sync
- raw unadjusted daily K line sync
- stock list and stock detail queries
- daily K line and latest daily K line queries
- sync task and sync log queries
- health checks
- enum metadata

Excluded:

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

## Documentation Map

- `AGENTS.md`: project rules for models and agents
- `CONTEXT.md`: domain glossary
- `domain/market/INDEX.md`: market package guide
- `specs/phase1-strategy.md`: strategic scope and priorities
- `specs/api-contract-decisions.md`: API contract decisions
- `specs/phase1-table-structure.md`: MySQL table design
- `specs/sync-task-behavior.md`: sync task behavior
- `specs/provider-tushare-adapter.md`: Tushare provider mapping
- `specs/runtime-and-testing.md`: runtime, config, readiness, and test rules
- `specs/implementation-plan.md`: staged implementation plan
- `docs/git-conventions.md`: Git workflow and commit conventions

## Build And Test

Default unit test command:

```sh
go test ./...
```

If the default Go build cache is not writable:

```sh
GOCACHE=/tmp/stock-lens-go-cache go test ./...
```

Tests that depend on `gomonkey` may need inlining disabled:

```sh
go test ./... -gcflags=all="-N -l"
```

Do not make the no-inline command the default unless a specific test flow requires monkey patching.
