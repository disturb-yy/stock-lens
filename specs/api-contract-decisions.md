# API Contract Decisions

This document records Phase 1 API contract decisions as they are finalized.

## Naming Conventions

Go identifiers use Go-style `CamelCase` or `camelCase`, such as `TaskUID`, `PageSize`, `StartDate`, `AssetType`, `DailyKLine`, `LatestDailyKLine`, `PctChange`, and `AffectedRows`.

HTTP JSON fields, query parameters, and request body fields use `snake_case`, such as `task_uid`, `page_size`, `start_date`, `asset_type`, `daily_k_line`, `latest_daily_k_line`, `pct_change`, and `affected_rows`.

SQL table names, column names, and index names use `snake_case`, such as `sync_tasks`, `task_uid`, `trade_date`, `asset_type`, and `idx_sync_tasks_created_at`.

Enum values remain uppercase English tokens, including underscore-separated values when needed, such as `CN`, `STOCK`, `TUSHARE`, `SYNC_DAILY_K_LINES`, and `PARTIAL_SUCCESS`.

Go structs may expose API fields through JSON tags:

```go
type SyncTaskResponse struct {
    TaskUID  string `json:"task_uid"`
    TaskType string `json:"task_type"`
    PageSize int    `json:"page_size,omitempty"`
}
```

## Pagination

List endpoints return pagination inside `data` with this structure:

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

Rules:

- `page` defaults to `1`.
- `page` must be greater than or equal to `1`.
- `page_size` defaults to `20`.
- `page_size` must be greater than or equal to `1`.
- `page_size` must be less than or equal to `100`.
- `total` is the total number of records matching the filters.
- `total_pages` is `ceil(total / page_size)`.
- `items` is an empty array when there are no results, not `null`.
- `pagination` is always returned for paginated list endpoints.

## Date Ranges

HTTP API date parameters use `YYYY-MM-DD` and closed intervals.

Rules:

- Date ranges include both `start_date` and `end_date`.
- `start_date > end_date` returns a date range error.
- Invalid date formats return an invalid argument error.
- Query APIs may accept future dates; if no data exists, they return an empty result.
- Sync APIs default missing `end_date` to the latest open trading day when applicable.
- Business date boundaries and `today` use the service business timezone, `Asia/Shanghai`.
- Date format errors return `INVALID_ARGUMENT`.
- Logical date range errors return `MARKET_INVALID_DATE_RANGE`.
- Date ranges that exceed endpoint limits return `MARKET_DATE_RANGE_TOO_LARGE`.

Applies to:

- `GET /api/v1/market/stocks/{symbol}/daily-k-lines`
- `GET /api/v1/market/trade-calendars`
- `POST /api/v1/market/sync/daily-k-lines`
- `POST /api/v1/market/sync/trade-calendars`

## Market And Asset Type

Market APIs accept `market` and `asset_type` as optional query or body parameters depending on the endpoint.

Rules:

- `market` defaults to `CN`.
- `asset_type` defaults to `STOCK`.
- Phase 1 only accepts `market=CN`.
- Phase 1 only accepts `asset_type=STOCK`.
- Enum values are case-sensitive and must be uppercase.
- Invalid values return an invalid argument error.

## Symbol

Phase 1 stock symbols are 6-digit A-share codes.

Rules:

- Accept symbols such as `600519`.
- Do not accept provider-specific codes such as `600519.SH`.
- Invalid symbols return `MARKET_INVALID_SYMBOL`.
- `ts_code` belongs to provider boundaries, not public stock APIs.

## Empty Results

Rules:

- List endpoints return `items: []` when there are no matching records.
- Stock detail returns `404` with `MARKET_STOCK_NOT_FOUND` when the stock does not exist.
- Stock detail returns `latest_daily_k_line: null` when the stock exists but no daily K line has been synchronized.
- Missing resources return `404` with a specific not-found error.
- Existing resources with missing optional child data return `null` for that child data.

## Response Body

Success and error responses use the uniform response envelope:

```json
{
  "code": "OK",
  "message": "ok",
  "request_id": "req_xxx",
  "data": {}
}
```

Success responses use `code = "OK"` and `message = "ok"`.

Error messages use stable English phrases in Phase 1. Clients must use `code`, not `message`, for programmatic handling.

Internal database auto-increment IDs are not exposed by public APIs. Stocks use `symbol`; sync tasks use `task_uid`.

Date fields such as `trade_date`, `list_date`, `delist_date`, `cal_date`, and `pretrade_date` use `YYYY-MM-DD`.

Datetime fields such as `synced_at`, `started_at`, `finished_at`, and `created_at` use ISO 8601 with timezone, such as `2026-06-27T10:00:00.123+08:00`.

Datetime responses use the service business timezone, `Asia/Shanghai`, and therefore normally carry `+08:00`.

## Daily K Line Query

Rules:

- Default range is the latest 1 year.
- Missing `end_date` defaults to the stock's latest locally synchronized `trade_date`.
- Missing `start_date` defaults to 1 year before `end_date`.
- If no local daily K line exists for the stock, the default query returns an empty array.
- Maximum range is 5 years.
- Results are not paginated in Phase 1.
- Results are sorted by `trade_date ASC`.
- Ranges larger than 5 years return `MARKET_DATE_RANGE_TOO_LARGE`.

Daily K line API fields use short market-data names:

```text
open
high
low
close
change
```

Domain models and database columns keep explicit names such as `OpenPrice`, `HighPrice`, `LowPrice`, `ClosePrice`, `ChangeAmt`, `open_price`, and `change_amt`.

Market numeric fields in HTTP responses are strings, not JSON numbers.

Daily K line responses expose unit fields:

```text
pct_change_unit = percent
volume_unit     = lot
amount_unit     = thousand_cny
```

## Sync Task Creation

Sync creation endpoints create a background task and return immediately.

Response data:

```json
{
  "task_uid": "01J2Z3ABCDEF123456789XYZAB",
  "status": "PENDING"
}
```

Rules:

- Sync creation does not wait for task completion.
- Clients confirm completion by querying the task detail endpoint.
- Phase 1 does not provide WebSocket, SSE, webhook callbacks, or blocking wait endpoints.

Task status flow:

```text
PENDING -> RUNNING -> SUCCESS
PENDING -> RUNNING -> FAILED
PENDING -> RUNNING -> PARTIAL_SUCCESS
```

Task detail response data:

```json
{
  "task_uid": "01J2Z3ABCDEF123456789XYZAB",
  "task_type": "SYNC_DAILY_K_LINES",
  "status": "RUNNING",
  "market": "CN",
  "asset_type": "STOCK",
  "data_source": "TUSHARE",
  "total_items": 5000,
  "processed_items": 120,
  "success_items": 118,
  "failed_items": 2,
  "request_id": "req_xxx",
  "started_at": "2026-06-27T10:00:00.123+08:00",
  "finished_at": null,
  "error_msg": ""
}
```

Client polling:

- Poll `GET /api/v1/market/sync/tasks/{task_uid}` every 1 to 3 seconds.
- Stop polling when status is `SUCCESS`, `FAILED`, or `PARTIAL_SUCCESS`.
- Query task logs when status is `FAILED` or `PARTIAL_SUCCESS`.

## Error Codes

Rules:

- Market-domain errors use the `MARKET_` prefix.
- Common errors include `INVALID_ARGUMENT`, `UNAUTHORIZED`, and `INTERNAL_ERROR`.
- HTTP status codes still carry protocol semantics.
- Error responses are not wrapped as HTTP 200 unless the operation actually succeeded.

## Stock APIs

### List Stocks

Endpoint:

```http
GET /api/v1/market/stocks
```

Query parameters:

```text
page       optional, default 1
page_size  optional, default 20, max 100
keyword    optional, matches symbol or name
exchange   optional
status     optional
market     optional, default CN
asset_type optional, default STOCK
```

Response item:

```json
{
  "symbol": "600519",
  "name": "贵州茅台",
  "market": "CN",
  "asset_type": "STOCK",
  "exchange": "SSE",
  "board": "MAIN",
  "area": "贵州",
  "industry": "白酒",
  "status": "LISTED",
  "list_date": "2001-08-27",
  "delist_date": null,
  "data_source": "TUSHARE",
  "synced_at": "2026-06-27T10:00:00.123+08:00"
}
```

Rules:

- The list endpoint does not include latest daily K line in Phase 1.
- `keyword` performs fuzzy matching against `symbol` and `name`.
- `exchange` and `status` are exact enum filters.

### Get Stock Detail

Endpoint:

```http
GET /api/v1/market/stocks/{symbol}
```

Query parameters:

```text
market     optional, default CN
asset_type optional, default STOCK
```

Response data:

```json
{
  "symbol": "600519",
  "name": "贵州茅台",
  "market": "CN",
  "asset_type": "STOCK",
  "exchange": "SSE",
  "board": "MAIN",
  "area": "贵州",
  "industry": "白酒",
  "status": "LISTED",
  "list_date": "2001-08-27",
  "delist_date": null,
  "data_source": "TUSHARE",
  "synced_at": "2026-06-27T10:00:00.123+08:00",
  "latest_daily_k_line": {
    "trade_date": "2026-06-25",
    "open": "1500.0000",
    "high": "1510.0000",
    "low": "1490.0000",
    "close": "1505.0000",
    "pre_close": "1498.0000",
    "change": "7.0000",
    "pct_change": "0.4673",
    "pct_change_unit": "percent",
    "volume": "123456.0000",
    "volume_unit": "lot",
    "amount": "987654321.0000",
    "amount_unit": "thousand_cny"
  }
}
```

Rules:

- `latest_daily_k_line` is `null` when the stock exists but no daily K line has been synchronized.
- Missing stock returns `404` with `MARKET_STOCK_NOT_FOUND`.
- Stock detail includes `latest_daily_k_line`; stock list items do not.

## Daily K Line APIs

### List Daily K Lines

Endpoint:

```http
GET /api/v1/market/stocks/{symbol}/daily-k-lines
```

Query parameters:

```text
start_date optional, default latest 1 year
end_date   optional, default latest available date
market     optional, default CN
asset_type optional, default STOCK
```

Response item:

```json
{
  "symbol": "600519",
  "market": "CN",
  "asset_type": "STOCK",
  "trade_date": "2026-06-25",
  "open": "1500.0000",
  "high": "1510.0000",
  "low": "1490.0000",
  "close": "1505.0000",
  "pre_close": "1498.0000",
  "change": "7.0000",
  "pct_change": "0.4673",
  "pct_change_unit": "percent",
  "volume": "123456.0000",
  "volume_unit": "lot",
  "amount": "987654321.0000",
  "amount_unit": "thousand_cny",
  "data_source": "TUSHARE",
  "synced_at": "2026-06-27T10:00:00.123+08:00"
}
```

Rules:

- Response data is an array, not a paginated object.
- Results are sorted by `trade_date ASC`.
- Missing stock returns `404` with `MARKET_STOCK_NOT_FOUND`.
- No K line data returns an empty array.

### Get Latest Daily K Line

Endpoint:

```http
GET /api/v1/market/stocks/{symbol}/latest-daily-k-line
```

Rules:

- Missing stock returns `404` with `MARKET_STOCK_NOT_FOUND`.
- Existing stock without daily K line returns `data: null`.
- Existing stock with data returns the same item shape as list daily K lines.
- The latest daily K line endpoint remains available even though stock detail also embeds `latest_daily_k_line`.

## Trade Calendar APIs

### List Trade Calendars

Endpoint:

```http
GET /api/v1/market/trade-calendars
```

Query parameters:

```text
start_date optional
end_date   optional
market     optional, default CN
exchange   optional, default SSE
```

Response item:

```json
{
  "market": "CN",
  "exchange": "SSE",
  "cal_date": "2026-06-25",
  "is_open": true,
  "pretrade_date": "2026-06-24",
  "data_source": "TUSHARE",
  "synced_at": "2026-06-27T10:00:00.123+08:00"
}
```

Rules:

- Response data is an array.
- Results are sorted by `cal_date ASC`.
- Missing `start_date` and `end_date` default to the latest 1 year.
- Date ranges are closed intervals.
- Future dates are accepted; dates without rows simply do not appear in the list.
- Very large ranges above 30 years return `MARKET_DATE_RANGE_TOO_LARGE`.

### Get Latest Open Day

Endpoint:

```http
GET /api/v1/market/trade-calendars/latest-open-day
```

Query parameters:

```text
market   optional, default CN
exchange optional, default SSE
```

Rules:

- Returns the latest row where `is_open = true`.
- If calendar data is missing, returns `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`.
- The latest open day is derived from the maximum `cal_date` where `is_open = true` for the requested `market + exchange`.
- The default latest open day uses `market=CN` and `exchange=SSE`.
- Latest open day is not the same concept as today or a stock's latest local daily K line date.

### Check Is Open

Endpoint:

```http
GET /api/v1/market/trade-calendars/is-open
```

Query parameters:

```text
date     optional, default today
market   optional, default CN
exchange optional, default SSE
```

Response data:

```json
{
  "market": "CN",
  "exchange": "SSE",
  "date": "2026-06-25",
  "is_open": true,
  "pretrade_date": "2026-06-24"
}
```

## Sync APIs

All sync creation endpoints require Admin Token.

### Sync Trade Calendars

Endpoint:

```http
POST /api/v1/market/sync/trade-calendars
```

Request body:

```json
{
  "market": "CN",
  "exchange": "SSE",
  "start_date": "2021-01-01",
  "end_date": "2027-12-31"
}
```

Rules:

- Missing range defaults to latest 5 years plus next 1 year.
- The default range is anchored to the current date in the service business timezone.
- Sync ranges are closed intervals.
- Trade-calendar sync accepts future dates.
- Very large ranges above 30 years return `MARKET_DATE_RANGE_TOO_LARGE`.
- Returns sync task creation response.

### Sync Stocks

Endpoint:

```http
POST /api/v1/market/sync/stocks
```

Request body:

```json
{
  "market": "CN",
  "asset_type": "STOCK"
}
```

Rules:

- Syncs `LISTED`, `DELISTED`, and `PAUSED` stocks.
- Any status fetch failure makes the task `FAILED`.
- Returns sync task creation response.

### Sync Daily K Lines

Endpoint:

```http
POST /api/v1/market/sync/daily-k-lines
```

Request body:

```json
{
  "market": "CN",
  "asset_type": "STOCK",
  "symbol": "600519",
  "start_date": "2023-01-01",
  "end_date": "2026-06-25"
}
```

Rules:

- Missing `symbol` means full-market sync.
- Present `symbol` means single-stock sync.
- Missing `start_date` defaults to latest 3 years.
- Missing `end_date` defaults to latest open day.
- Full-market sync maximum range is 5 years.
- Single-stock sync maximum range is 20 years.
- Full-market sync defaults to `LISTED` stocks only.
- Full-market per-stock failure does not block the entire task and may produce `PARTIAL_SUCCESS`.
- Missing trade calendar returns `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`.
- Full-market sync with no stocks returns `MARKET_STOCKS_NOT_INITIALIZED`.
- Single-stock sync with missing stock returns `MARKET_STOCK_NOT_FOUND`.
- Daily K sync does not default directly to the system current date.
- Phase 1 uses `CN + SSE` as the default A-share trading calendar for daily K sync and does not dynamically choose calendars by stock exchange.

### Get Sync Task

Endpoint:

```http
GET /api/v1/market/sync/tasks/{task_uid}
```

Rules:

- Missing task returns `404` with `MARKET_SYNC_TASK_NOT_FOUND`.

### List Sync Logs

Endpoint:

```http
GET /api/v1/market/sync/tasks/{task_uid}/logs
```

Query parameters:

```text
page      optional, default 1
page_size optional, default 20, max 100
status    optional
symbol    optional
```

Rules:

- Response uses standard pagination.
- Logs are sorted by `created_at ASC`.
- Missing task returns `404` with `MARKET_SYNC_TASK_NOT_FOUND`.

## Enum API

Endpoint:

```http
GET /api/v1/market/meta/enums
```

Recommended response data shape:

```json
{
  "markets": ["CN"],
  "asset_types": ["STOCK"],
  "exchanges": ["SSE", "SZSE", "BSE"],
  "stock_statuses": ["LISTED", "DELISTED", "PAUSED"],
  "data_sources": ["TUSHARE", "MOCK"]
}
```

Rules:

- Values are returned from domain constants.
- Phase 1 does not use database dictionary tables.
- Phase 1 does not support multi-language enum labels.
