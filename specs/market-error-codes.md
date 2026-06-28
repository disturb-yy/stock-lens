# Market Error Codes

This document freezes Phase 1 market error codes and HTTP status mapping.

## Response Shape

Errors use the uniform API response body:

```json
{
  "code": "MARKET_STOCK_NOT_FOUND",
  "message": "stock not found",
  "request_id": "req_xxx",
  "data": null
}
```

When a sync task conflicts with an active task, `data` may include the active task:

```json
{
  "code": "MARKET_SYNC_TASK_CONFLICT",
  "message": "sync task already running",
  "request_id": "req_xxx",
  "data": {
    "task_uid": "01J2Z3ABCDEF123456789XYZAB"
  }
}
```

## Common Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `INVALID_ARGUMENT` | 400 | Generic argument error when no more specific market error applies |
| `UNAUTHORIZED` | 401 | Missing or invalid Admin Token |
| `NOT_FOUND` | 404 | Generic resource not found when no more specific market error applies |
| `INTERNAL_ERROR` | 500 | Unclassified internal error |

## Market Argument Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `MARKET_INVALID_SYMBOL` | 400 | Symbol is not a 6-digit A-share code, or a provider-specific code such as `600519.SH` was passed |
| `MARKET_INVALID_MARKET` | 400 | Market is not supported in Phase 1; only `CN` is accepted |
| `MARKET_INVALID_ASSET_TYPE` | 400 | Asset type is not supported in Phase 1; only `STOCK` is accepted |
| `MARKET_INVALID_EXCHANGE` | 400 | Exchange is not one of `SSE`, `SZSE`, or `BSE` |
| `MARKET_INVALID_STATUS` | 400 | Stock status is not one of `LISTED`, `DELISTED`, or `PAUSED` |
| `MARKET_INVALID_DATE_RANGE` | 400 | `start_date` is later than `end_date` |
| `MARKET_DATE_RANGE_TOO_LARGE` | 400 | Query or sync date range exceeds the allowed maximum |
| `MARKET_INVALID_TASK_STATUS` | 400 | Task or log status filter is invalid |

## Market Resource Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `MARKET_STOCK_NOT_FOUND` | 404 | Stock does not exist |
| `MARKET_SYNC_TASK_NOT_FOUND` | 404 | Sync task does not exist |
| `MARKET_TRADE_CALENDAR_NOT_FOUND` | 404 | A specific queried trade-calendar date does not exist |

## Market Prerequisite Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `MARKET_TRADE_CALENDAR_NOT_INITIALIZED` | 409 | Latest open day is unavailable, or daily K sync requires local trade-calendar data that has not been initialized |
| `MARKET_STOCKS_NOT_INITIALIZED` | 409 | Full-market daily K sync requires local stock master data that has not been initialized |

Phase 1 does not define a separate `MARKET_LATEST_OPEN_DAY_NOT_FOUND`. Use `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`.

## Market Conflict Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `MARKET_SYNC_TASK_CONFLICT` | 409 | A `PENDING` or `RUNNING` sync task already exists |

## Provider And Store Errors

| Error Code | HTTP | Meaning |
|---|---:|---|
| `MARKET_PROVIDER_ERROR` | 502 | Tushare network error, rate limit, malformed response, empty stock list, or empty trade calendar |
| `MARKET_STORE_ERROR` | 500 | Repository, query, upsert, or transaction failure |
| `MARKET_SYNC_TASK_FAILED` | 500 | Sync task failed without a more specific error code |

## Empty Result Rules

| Scenario | Result |
|---|---|
| Query future dates with no data | `200` with an empty array |
| Existing stock has no latest daily K line | `200` with `data: null` |
| Stock list has no matches | `200` with `items: []` |
| Trade-calendar list has no matches | `200` with `[]` |
| Queried stock does not exist | `404` with `MARKET_STOCK_NOT_FOUND` |
| Latest open day cannot be derived from local calendar data | `409` with `MARKET_TRADE_CALENDAR_NOT_INITIALIZED` |

## Sync Task Creation Rules

These errors are returned before task creation and do not create failed sync tasks:

- invalid argument
- invalid date
- date range too large
- active task conflict
- missing or invalid Admin Token

These errors create a task first and then mark the task `FAILED`:

- missing trade-calendar prerequisite
- missing stock master prerequisite
- missing stock for single-stock daily K sync
- provider error during execution
- store error during execution

If a store error occurs before a task can be persisted, no task is created.
