# Phase 1 Table Structure

This document freezes the Phase 1 MySQL table design for the market-data backend.

## Design Rules

The schema follows an identity-first, upsert-oriented design:

- Internal database identity uses `BIGINT UNSIGNED AUTO_INCREMENT`.
- Public opaque identity uses ULID only where a public resource ID is needed, such as sync tasks.
- Market data identity uses business unique keys.
- Sync writes are idempotent and use batch upsert.
- Tables do not use database foreign key constraints in Phase 1.
- Lifecycle state is represented by status fields, not soft delete.
- Trade dates use `DATE`; audit timestamps use `DATETIME(3)`.
- Market numeric values use `DECIMAL`, not floating-point types.
- Stored market-data units follow the data source's raw units, with API DTOs exposing unit fields.

## stocks

`stocks` stores stock master data.

Business identity:

```text
market + asset_type + symbol
```

Columns:

```text
id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
market       VARCHAR(16) NOT NULL
asset_type   VARCHAR(32) NOT NULL
symbol       VARCHAR(32) NOT NULL
ts_code      VARCHAR(32) NOT NULL
name         VARCHAR(128) NOT NULL
exchange     VARCHAR(32) NOT NULL
board        VARCHAR(32) NOT NULL DEFAULT 'UNKNOWN'
area         VARCHAR(64) NOT NULL DEFAULT ''
industry     VARCHAR(128) NOT NULL DEFAULT ''
status       VARCHAR(32) NOT NULL
list_date    DATE NULL
delist_date  DATE NULL
data_source  VARCHAR(32) NOT NULL
synced_at    DATETIME(3) NOT NULL
created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
updated_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

Indexes:

```text
UNIQUE uk_stock_identity (market, asset_type, symbol)
INDEX  idx_stocks_filter (market, asset_type, exchange, status)
INDEX  idx_stocks_symbol (symbol)
INDEX  idx_stocks_name (name)
```

Rules:

- Stocks do not have a public `uid`.
- API lookup uses `symbol` plus optional `market` and `asset_type`.
- Local rows are not deleted when a stock delists.
- `status` expresses lifecycle state.
- `ts_code` is stored as data-source identity, not as the system identity.

## daily_k_lines

`daily_k_lines` stores raw, unadjusted daily OHLCV facts.

Business identity:

```text
market + asset_type + symbol + trade_date
```

Columns:

```text
id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
market       VARCHAR(16) NOT NULL
asset_type   VARCHAR(32) NOT NULL
symbol       VARCHAR(32) NOT NULL
trade_date   DATE NOT NULL
open_price   DECIMAL(20,4) NOT NULL
high_price   DECIMAL(20,4) NOT NULL
low_price    DECIMAL(20,4) NOT NULL
close_price  DECIMAL(20,4) NOT NULL
pre_close    DECIMAL(20,4) NOT NULL
change_amt   DECIMAL(20,4) NOT NULL
pct_change   DECIMAL(20,4) NOT NULL
volume       DECIMAL(24,4) NOT NULL
amount       DECIMAL(24,4) NOT NULL
data_source  VARCHAR(32) NOT NULL
synced_at    DATETIME(3) NOT NULL
created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
updated_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

Indexes:

```text
UNIQUE uk_daily_kline_identity (market, asset_type, symbol, trade_date)
INDEX  idx_daily_kline_query (market, asset_type, symbol, trade_date)
```

Rules:

- Do not store `ts_code`, stock `name`, or `exchange`.
- Do not store adjustment type in Phase 1.
- Store only raw unadjusted daily K lines.
- `pct_change` stores percentage units, so `2.35` means `2.35%`.
- `volume` stores lots, following Tushare raw units.
- `amount` stores thousand CNY, following Tushare raw units.

## trade_calendars

`trade_calendars` stores open and closed trading dates.

Business identity:

```text
market + exchange + cal_date
```

Columns:

```text
id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
market         VARCHAR(16) NOT NULL
exchange       VARCHAR(32) NOT NULL
cal_date       DATE NOT NULL
is_open        TINYINT(1) NOT NULL
pretrade_date  DATE NULL
data_source    VARCHAR(32) NOT NULL
synced_at      DATETIME(3) NOT NULL
created_at     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
updated_at     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

Indexes:

```text
UNIQUE uk_trade_calendar_identity (market, exchange, cal_date)
INDEX  idx_trade_calendar_open_day (market, exchange, is_open, cal_date)
```

Rules:

- Phase 1 defaults to `market=CN` and `exchange=SSE`.
- Latest open day is derived from rows where `is_open = 1`.

## sync_tasks

`sync_tasks` stores user-triggered sync task lifecycle and progress.

Public identity:

```text
uid ULID CHAR(26)
```

Columns:

```text
id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
uid              CHAR(26) NOT NULL
task_type        VARCHAR(64) NOT NULL
market           VARCHAR(16) NOT NULL
asset_type       VARCHAR(32) NOT NULL DEFAULT 'STOCK'
data_source      VARCHAR(32) NOT NULL
status           VARCHAR(32) NOT NULL
total_items      BIGINT NOT NULL DEFAULT 0
processed_items  BIGINT NOT NULL DEFAULT 0
success_items    BIGINT NOT NULL DEFAULT 0
failed_items     BIGINT NOT NULL DEFAULT 0
request_id       VARCHAR(128) NOT NULL DEFAULT ''
started_at       DATETIME(3) NULL
finished_at      DATETIME(3) NULL
error_msg        TEXT NULL
created_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
updated_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

Indexes:

```text
UNIQUE uk_sync_tasks_uid (uid)
INDEX  idx_sync_tasks_status (status)
INDEX  idx_sync_tasks_created_at (created_at)
```

Rules:

- API exposes `uid`, not internal `id`.
- Phase 1 supports `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, and `PARTIAL_SUCCESS`.
- Phase 1 does not support cancellation states.
- Service startup marks stale `PENDING` and `RUNNING` tasks as `FAILED`.

## sync_logs

`sync_logs` stores task-level and stock-level sync observations.

Columns:

```text
id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
task_id        BIGINT UNSIGNED NOT NULL
task_uid       CHAR(26) NOT NULL
step           VARCHAR(64) NOT NULL
status         VARCHAR(32) NOT NULL
market         VARCHAR(16) NOT NULL DEFAULT ''
asset_type     VARCHAR(32) NOT NULL DEFAULT ''
symbol         VARCHAR(32) NOT NULL DEFAULT ''
data_source    VARCHAR(32) NOT NULL DEFAULT ''
message        TEXT NULL
error_detail   TEXT NULL
affected_rows  BIGINT NOT NULL DEFAULT 0
created_at     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
```

Indexes:

```text
INDEX idx_sync_logs_task_id (task_id)
INDEX idx_sync_logs_task_uid (task_uid)
INDEX idx_sync_logs_created_at (created_at)
```

Rules:

- Do not record one log row per K line.
- Record task step summaries and stock-level success or failure summaries.
- `task_id` is an internal reference and has no database foreign key constraint.
- `task_uid` is duplicated intentionally for API and troubleshooting queries.

## Upsert Rules

All sync writes use batch upsert. Default batch size is `500`.

`stocks` conflicts update:

```text
name, exchange, board, area, industry, status, list_date, delist_date,
data_source, synced_at, updated_at
```

`daily_k_lines` conflicts update:

```text
open_price, high_price, low_price, close_price, pre_close, change_amt,
pct_change, volume, amount, data_source, synced_at, updated_at
```

`trade_calendars` conflicts update:

```text
is_open, pretrade_date, data_source, synced_at, updated_at
```

Conflict updates never change internal primary keys, business identity columns, or `created_at`.
