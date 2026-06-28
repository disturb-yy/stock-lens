# Sync Task Behavior

Phase 1 sync tasks are manually triggered, globally mutually exclusive, and executed by a single service instance goroutine.

## Concurrency

Only one active sync task may exist at a time. An active task is any task with status `PENDING` or `RUNNING`.

When a sync creation request arrives while another task is active, the API returns `409 Conflict` with `MARKET_SYNC_TASK_CONFLICT`. If the active task can be found, the response should include its `task_uid`.

Phase 1 does not support task cancellation, scheduled sync, queues, distributed workers, or task recovery after restart.

## State Flow

Allowed task status flows:

```text
PENDING -> RUNNING -> SUCCESS
PENDING -> RUNNING -> FAILED
PENDING -> RUNNING -> PARTIAL_SUCCESS
```

Terminal states are not changed.

Sync creation endpoints first persist a `PENDING` task and return immediately. The background goroutine marks the task `RUNNING` when execution starts.

On service startup, stale `PENDING` and `RUNNING` tasks are marked `FAILED` with a reason such as `service restarted before task completed`.

If task execution panics, the goroutine recovers, writes a sync log, marks the task `FAILED`, and stores a panic summary in `error_msg`.

## Preflight Validation

Invalid client input is rejected before task creation. These errors do not create failed tasks.

Examples:

- invalid symbol
- invalid enum value
- invalid date format
- `start_date > end_date`
- date range larger than the allowed maximum

Date range violations return `MARKET_DATE_RANGE_TOO_LARGE`.

Sync tasks that depend on existing local data fail when prerequisites are missing:

- missing trade calendar: `FAILED` with `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`
- full-market daily K sync without stock master data: `FAILED` with `MARKET_STOCKS_NOT_INITIALIZED`
- single-stock daily K sync for a missing stock: `FAILED` with `MARKET_STOCK_NOT_FOUND`

## Sync Scope

Stock master sync fetches and stores `LISTED`, `DELISTED`, and `PAUSED` stocks. If fetching any stock status fails, the whole task fails.

Full-market daily K sync defaults to `LISTED` stocks only. Single-stock daily K sync may run for `LISTED`, `PAUSED`, or `DELISTED` stocks.

Missing `end_date` for daily K sync defaults to the latest open trading day. Missing `start_date` defaults to three years before `end_date`.

Full-market daily K sync allows at most 5 years. Single-stock daily K sync allows at most 20 years.

Trade-calendar sync defaults missing range to the latest 5 years plus the next 1 year.

Phase 1 uses `market=CN` and `exchange=SSE` as the default A-share trading calendar. Daily K sync does not dynamically choose a trade calendar by stock exchange.

The latest open day is derived from the maximum `cal_date` where `is_open = true` for a `market + exchange`. It is not the same as today and not the same as a stock's latest local daily K line date.

The service business timezone is `Asia/Shanghai`. Date boundaries, default ranges, and `today` are interpreted in that timezone. Daily K sync defaults use latest open day rather than the system current date.

Trade-calendar sync ranges are closed intervals, may include future dates, and reject ranges above 30 years with `MARKET_DATE_RANGE_TOO_LARGE`.

## Empty Provider Results

Provider empty-result handling:

- single-stock daily K empty result: `SUCCESS`, with `affected_rows = 0`
- one stock in full-market daily K empty result: stock-level `SUCCESS`, with `affected_rows = 0`
- trade-calendar empty result: `FAILED` with `MARKET_PROVIDER_ERROR`
- stock-list empty result: `FAILED` with `MARKET_PROVIDER_ERROR`

Empty K line data may be valid for a date range. Empty stock lists and empty trade calendars are treated as provider or mapping failures in Phase 1.

## Failure Semantics

Full-market daily K sync isolates failures by stock. Per-stock provider or upsert failures are logged and processing continues for the remaining stocks.

Final status rules for full-market daily K sync:

- all stocks succeed: `SUCCESS`
- at least one stock succeeds and at least one fails: `PARTIAL_SUCCESS`
- all stocks fail: `FAILED`

For stock master sync, trade-calendar sync, and single-stock daily K sync, provider errors, network errors, and upsert failures fail the task.

Phase 1 does not perform automatic retry. A provider implementation may add a small internal retry later, but the task model does not rely on retry semantics.

## Progress Counters

Task counters use object-level progress:

```text
total_items      = expected number of objects to process
processed_items  = completed objects, whether successful or failed
success_items    = successful objects
failed_items     = failed objects
```

For full-market daily K sync, the object is a stock, not an individual K line.

## Sync Logs

Sync logs are task-level and stock-level summaries. They do not record one row per K line.

Recommended log points:

- task start
- preflight check
- provider fetch summary
- batch upsert summary
- per-stock daily K success or failure summary
- task end

Log statuses are `SUCCESS`, `FAILED`, and `WARNING`.

`affected_rows` means:

- stock sync: number of upserted stock rows
- trade-calendar sync: number of upserted calendar rows
- single-stock daily K sync: number of upserted K line rows
- full-market daily K sync: number of upserted K line rows for that stock-level log
