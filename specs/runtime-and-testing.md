# Runtime And Testing

This document freezes Phase 1 runtime defaults, readiness semantics, and unit-test conventions.

## Server

The default HTTP server port is `30078`.

Local default base URL:

```text
http://localhost:30078
```

Public API endpoints are served under `/api/v1`, with market endpoints under `/api/v1/market`.

Health endpoints:

```text
GET /healthz
GET /readyz
```

`/healthz` means the process is alive and can respond to HTTP.

`/readyz` means the service is ready to serve traffic. In Phase 1, readiness should include successful startup configuration validation and MySQL connectivity.

## Configuration Expectations

The service should expose explicit configuration for:

- server address and port, defaulting to port `30078`
- business timezone, defaulting to `Asia/Shanghai`
- MySQL connection settings
- Admin Token
- market provider selection, `mock` or `tushare`
- Tushare token and base URL when using the Tushare provider
- sync batch size
- log level

Mock provider mode may start without a Tushare token. Tushare provider mode requires a Tushare token.

Default configuration path:

```text
configs/config.yaml
```

Test configuration path:

```text
configs/config.test.yaml
```

Recommended configuration shape:

```yaml
server:
  addr: "0.0.0.0"
  port: 30078
  timezone: "Asia/Shanghai"

database:
  driver: "mysql"
  dsn: "${MYSQL_DSN}"

auth:
  admin_token: "${ADMIN_TOKEN}"

market:
  provider: "mock"
  batch_size: 500

tushare:
  base_url: "https://api.tushare.pro"
  token: "${TUSHARE_TOKEN}"

log:
  level: "info"
```

YAML values support environment-variable expansion, such as `${TUSHARE_TOKEN}`. Missing required values fail startup.

Configuration path priority:

```text
explicit config path flag > configs/config.yaml
```

Environment variables are used for YAML expansion only. Do not scatter direct environment lookups through the codebase.

Validation rules:

- `server.addr` defaults to `0.0.0.0`.
- `server.port` defaults to `30078`.
- `server.timezone` defaults to `Asia/Shanghai`; invalid timezones fail startup.
- `database.dsn` is required.
- `auth.admin_token` is required because sync APIs require Admin Token.
- `market.provider` defaults to `mock` and must be `mock` or `tushare`.
- `market.batch_size` defaults to `500`; non-positive values fail startup.
- `market.provider=mock` does not require `tushare.token`.
- `market.provider=tushare` requires `tushare.token`.
- `log.level` defaults to `info` and must be `debug`, `info`, `warn`, or `error`.

Admin-protected sync APIs use:

```text
Authorization: Bearer <token>
```

## Startup And Shutdown

Startup order:

1. Load configuration from the explicit config path or `configs/config.yaml`.
2. Expand environment variables in YAML values.
3. Validate configuration.
4. Initialize logging.
5. Connect to MySQL and ping it.
6. Mark stale `PENDING` and `RUNNING` sync tasks as `FAILED`.
7. Start the HTTP server.

Configuration validation failures stop startup before the HTTP server starts.

MySQL ping failure stops startup.

The service does not run migrations automatically on startup. Migrations are executed explicitly through a command or Makefile target.

`/readyz` returns ready only after configuration validation, MySQL connectivity, and stale task recovery have completed. Otherwise it returns a non-200 response.

`/healthz` does not check MySQL. It returns `200` when the process can respond over HTTP.

On `SIGINT` or `SIGTERM`, the service should stop accepting new requests and perform HTTP graceful shutdown with a timeout.

Phase 1 does not support cancellation of a running sync task during graceful shutdown. If the process exits while a task is running, the next startup marks the task `FAILED`.

## Unit Tests

Unit tests are preferred over broad integration tests for Phase 1.

Use table-driven tests by default when a function has multiple input/output cases.

Prefer the smallest useful unit test scope. Test the function or method behavior directly, and avoid pulling in handlers, repositories, providers, and databases unless that boundary is the behavior under test.

When a dependency function must be intercepted and ordinary interface substitution is not practical, `gomonkey` may be used in unit tests.

Tests that use `gomonkey` must consider Go inlining. If a patch depends on replacing an inlinable function, run that test with inlining disabled:

```sh
go test ./... -gcflags=all="-N -l"
```

Default test commands should not disable inlining. The no-inline command is reserved for tests or local debugging flows that actually need monkey patching.

Provider unit tests must use constructed raw responses rather than real Tushare network calls.

Repository integration tests, when added, should remain behind the `integration` build tag.
