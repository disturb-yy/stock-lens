# 第一阶段表结构

本文档冻结 market-data 后端第一阶段 MySQL 表设计。

## 设计规则

Schema 遵循 identity-first、upsert-oriented 的设计：

- 内部数据库身份使用 `BIGINT UNSIGNED AUTO_INCREMENT`。
- 公开不透明身份只在需要公开资源 ID 的地方使用 ULID，例如同步任务。
- 市场数据身份使用业务唯一键。
- 同步写入是幂等的，并使用批量 upsert。
- 第一阶段表不使用数据库外键约束。
- 生命周期状态通过 status 字段表示，而不是 soft delete。
- 交易日期使用 `DATE`；审计时间戳使用 `DATETIME(3)`。
- 市场数值使用 `DECIMAL`，不使用浮点类型。
- 存储的市场数据单位遵循数据源原始单位，API DTO 暴露单位字段。

## stocks

`stocks` 存储股票主数据。

业务身份：

```text
market + asset_type + symbol
```

字段：

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

索引：

```text
UNIQUE uk_stock_identity (market, asset_type, symbol)
INDEX  idx_stocks_filter (market, asset_type, exchange, status)
INDEX  idx_stocks_symbol (symbol)
INDEX  idx_stocks_name (name)
```

规则：

- 股票没有公开 `uid`。
- API lookup 使用 `symbol` 加可选的 `market` 和 `asset_type`。
- 股票退市时不删除本地行。
- `status` 表示生命周期状态。
- `ts_code` 存储为数据源身份，不是系统身份。

## daily_k_lines

`daily_k_lines` 存储原始、不复权的日 OHLCV 事实。

业务身份：

```text
market + asset_type + symbol + trade_date
```

字段：

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

索引：

```text
UNIQUE uk_daily_kline_identity (market, asset_type, symbol, trade_date)
INDEX  idx_daily_kline_query (market, asset_type, symbol, trade_date)
```

规则：

- 不存储 `ts_code`、股票 `name` 或 `exchange`。
- 第一阶段不存储复权类型。
- 只存储原始不复权日 K。
- `pct_change` 存储百分比单位，因此 `2.35` 表示 `2.35%`。
- `volume` 按 Tushare 原始单位存储为手。
- `amount` 按 Tushare 原始单位存储为千元人民币。

## trade_calendars

`trade_calendars` 存储开市和休市交易日期。

业务身份：

```text
market + exchange + cal_date
```

字段：

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

索引：

```text
UNIQUE uk_trade_calendar_identity (market, exchange, cal_date)
INDEX  idx_trade_calendar_open_day (market, exchange, is_open, cal_date)
```

规则：

- 第一阶段默认 `market=CN` 和 `exchange=SSE`。
- 最近开市日从 `is_open = 1` 的行推导。

## sync_tasks

`sync_tasks` 存储用户触发的同步任务生命周期和进度。

公开身份：

```text
uid ULID CHAR(26)
```

字段：

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

索引：

```text
UNIQUE uk_sync_tasks_uid (uid)
INDEX  idx_sync_tasks_status (status)
INDEX  idx_sync_tasks_created_at (created_at)
```

规则：

- API 暴露 `uid`，不暴露内部 `id`。
- 第一阶段支持 `PENDING`、`RUNNING`、`SUCCESS`、`FAILED` 和 `PARTIAL_SUCCESS`。
- 第一阶段不支持取消状态。
- 服务启动时将 stale `PENDING` 和 `RUNNING` 任务标记为 `FAILED`。

## sync_logs

`sync_logs` 存储任务级和股票级同步观察记录。

字段：

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

索引：

```text
INDEX idx_sync_logs_task_id (task_id)
INDEX idx_sync_logs_task_uid (task_uid)
INDEX idx_sync_logs_created_at (created_at)
```

规则：

- 不为每条 K 线记录一行日志。
- 记录任务步骤摘要和股票级成功或失败摘要。
- `task_id` 是内部引用，并且没有数据库外键约束。
- `task_uid` 为 API 和排障查询有意冗余存储。

## Upsert 规则

所有同步写入都使用批量 upsert。默认批大小为 `500`。

`stocks` 冲突时更新：

```text
name, exchange, board, area, industry, status, list_date, delist_date,
data_source, synced_at, updated_at
```

`daily_k_lines` 冲突时更新：

```text
open_price, high_price, low_price, close_price, pre_close, change_amt,
pct_change, volume, amount, data_source, synced_at, updated_at
```

`trade_calendars` 冲突时更新：

```text
is_open, pretrade_date, data_source, synced_at, updated_at
```

冲突更新永不改变内部主键、业务身份字段或 `created_at`。
