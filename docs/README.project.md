# Stock Lens Phase 1

Stock Lens Phase 1 is a market-data backend foundation for A-share stock data. It provides configuration-driven runtime startup, MySQL persistence, market-data sync tasks, and `/api/v1/market` query APIs.

## Local Runtime

Start local MySQL:

```sh
make dev-up
```

Apply migrations:

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' make migrate-up
```

Run the server:

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' \
ADMIN_TOKEN='local-admin-token' \
make run
```

The default HTTP port is `30078`.

## API And Config

- Health: `GET /healthz`
- Readiness: `GET /readyz`
- Market API prefix: `/api/v1/market`
- OpenAPI contract: `specs/openapi.yaml`
- Default config: `configs/config.yaml`
- Test config: `configs/config.test.yaml`

Admin-protected sync APIs use:

```text
Authorization: Bearer <admin-token>
```

## Validation

Run unit tests:

```sh
make test
```

Run formatting and vet checks:

```sh
make fmt-check
make vet
```

Run integration tests after applying migrations to a real MySQL database:

```sh
MYSQL_DSN='stock_lens:stock_lens@tcp(127.0.0.1:33306)/stock_lens?parseTime=true&loc=Asia%2FShanghai' make test-integration
```
