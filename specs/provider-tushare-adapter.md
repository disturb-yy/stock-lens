# Tushare Provider Adapter

This document freezes Phase 1 Tushare provider adaptation rules.

## Boundaries

The raw Tushare HTTP client lives in `pkg/tushare`. It owns Tushare protocol details such as token handling, `api_name`, explicit `fields`, raw request/response structures, provider error codes, rate limits, and optional low-level retry.

The market-domain Tushare adapter lives in `domain/market/provider_tushare.go`. It converts raw Tushare data into market domain models and returns those models through provider interfaces.

Services do not handle raw Tushare response structures.

## Identity Mapping

`stocks.ts_code` stores the Tushare identifier, such as `600519.SH`.

`daily_k_lines` does not store `ts_code`.

The provider extracts the system `symbol` from `ts_code`, for example `600519.SH` -> `600519`. Invalid `ts_code` values are provider mapping errors.

Exchange mapping uses the `ts_code` suffix or a Tushare exchange field:

```text
.SH -> SSE
.SZ -> SZSE
.BJ -> BSE
```

Unrecognized exchanges are provider mapping errors for stock master data.

## Stock Mapping

Board inference follows Phase 1 rules:

```text
688xxx          -> STAR
300xxx          -> GEM
8xxxxx / 4xxxxx -> BSE
other           -> MAIN
unknown         -> UNKNOWN
```

Stock status mapping:

```text
L -> LISTED
D -> DELISTED
P -> PAUSED
```

Unknown stock statuses are provider mapping errors.

Tushare provider output uses `DataSource = TUSHARE`.

## Date And Decimal Mapping

At the Tushare boundary, dates use `YYYYMMDD`.

Domain models use date/time types. HTTP APIs use `YYYY-MM-DD`. MySQL stores trade dates as `DATE`.

Tushare numeric values are converted to `decimal.Decimal`. Empty strings, non-numeric values, and precision parse failures are provider mapping errors.

Stored market-data units keep Tushare raw units:

```text
volume     = lot
amount     = thousand_cny
pct_change = percent
```

## Daily K Lines

Phase 1 fetches and stores only raw, unadjusted daily K lines.

The provider must not request front-adjusted or back-adjusted data. If the Tushare API requires an adjustment parameter, the provider must explicitly request raw unadjusted data.

## Empty Results

Provider empty-result handling follows sync task behavior:

- empty stock list: `MARKET_PROVIDER_ERROR`
- empty trade calendar: `MARKET_PROVIDER_ERROR`
- empty single-stock daily K line result: success with zero rows
- empty per-stock result during full-market daily K sync: stock-level success with zero rows

## Error Handling

Provider errors are wrapped as market domain errors, normally `MARKET_PROVIDER_ERROR`.

Internal error text may be written to sync logs and application logs, but sensitive values such as the Tushare token must not be logged.

Tushare token values, request headers containing tokens, and complete request bodies containing tokens must not appear in logs.

## Tests And Retry

Phase 1 does not run real Tushare network tests.

Provider tests use constructed raw responses rather than live Tushare calls.

Phase 1 task semantics do not rely on automatic retry. If lightweight retry is implemented, it belongs inside `pkg/tushare` and must not change sync task state semantics.

Raw client requests should explicitly list Tushare `fields` so Tushare default field changes do not break parsing.

## Mock Provider

Mock providers return market domain models and use `DataSource = MOCK`.

Mock providers should support controlled failure scenarios so tests can cover sync task failure and partial-success behavior.
