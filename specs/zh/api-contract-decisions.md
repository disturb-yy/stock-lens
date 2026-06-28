# API 合约决策

本文档记录第一阶段 API 合约决策，并在决策完成时更新。

## 命名约定

Go identifier 使用 Go 风格的 `CamelCase` 或 `camelCase`，例如 `TaskUID`、`PageSize`、`StartDate`、`AssetType`、`DailyKLine`、`LatestDailyKLine`、`PctChange` 和 `AffectedRows`。

HTTP JSON 字段、query 参数和 request body 字段使用 `snake_case`，例如 `task_uid`、`page_size`、`start_date`、`asset_type`、`daily_k_line`、`latest_daily_k_line`、`pct_change` 和 `affected_rows`。

SQL 表名、字段名和索引名使用 `snake_case`，例如 `sync_tasks`、`task_uid`、`trade_date`、`asset_type` 和 `idx_sync_tasks_created_at`。

枚举值保持大写英文 token，需要时使用下划线分隔，例如 `CN`、`STOCK`、`TUSHARE`、`SYNC_DAILY_K_LINES` 和 `PARTIAL_SUCCESS`。

Go struct 可以通过 JSON tag 暴露 API 字段：

```go
type SyncTaskResponse struct {
    TaskUID  string `json:"task_uid"`
    TaskType string `json:"task_type"`
    PageSize int    `json:"page_size,omitempty"`
}
```

## 分页

列表端点在 `data` 中返回分页信息，结构如下：

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

规则：

- `page` 默认值为 `1`。
- `page` 必须大于或等于 `1`。
- `page_size` 默认值为 `20`。
- `page_size` 必须大于或等于 `1`。
- `page_size` 必须小于或等于 `100`。
- `total` 是匹配过滤条件的总记录数。
- `total_pages` 为 `ceil(total / page_size)`。
- 没有结果时 `items` 是空数组，不是 `null`。
- 分页列表端点始终返回 `pagination`。

## 日期范围

HTTP API 日期参数使用 `YYYY-MM-DD` 和闭区间。

规则：

- 日期范围同时包含 `start_date` 和 `end_date`。
- `start_date > end_date` 返回日期范围错误。
- 无效日期格式返回 invalid argument 错误。
- 查询 API 可以接受未来日期；如果没有数据，返回空结果。
- 同步 API 在适用时将缺失的 `end_date` 默认设为最近开市交易日。
- 业务日期边界和 `today` 使用服务业务时区 `Asia/Shanghai`。
- 日期格式错误返回 `INVALID_ARGUMENT`。
- 日期逻辑范围错误返回 `MARKET_INVALID_DATE_RANGE`。
- 日期范围超过端点限制时返回 `MARKET_DATE_RANGE_TOO_LARGE`。

适用于：

- `GET /api/v1/market/stocks/{symbol}/daily-k-lines`
- `GET /api/v1/market/trade-calendars`
- `POST /api/v1/market/sync/daily-k-lines`
- `POST /api/v1/market/sync/trade-calendars`

## Market 和 Asset Type

Market API 根据端点不同，将 `market` 和 `asset_type` 作为可选 query 参数或 body 参数接收。

规则：

- `market` 默认值为 `CN`。
- `asset_type` 默认值为 `STOCK`。
- 第一阶段只接受 `market=CN`。
- 第一阶段只接受 `asset_type=STOCK`。
- 枚举值大小写敏感，必须为大写。
- 无效值返回 invalid argument 错误。

## Symbol

第一阶段股票 symbol 是 6 位 A 股代码。

规则：

- 接受 `600519` 这样的 symbol。
- 不接受 `600519.SH` 这样的 provider 专用代码。
- 无效 symbol 返回 `MARKET_INVALID_SYMBOL`。
- `ts_code` 属于 provider 边界，不属于公共股票 API。

## 空结果

规则：

- 列表端点在没有匹配记录时返回 `items: []`。
- 股票不存在时，股票详情返回 `404` 和 `MARKET_STOCK_NOT_FOUND`。
- 股票存在但没有同步日 K 时，股票详情返回 `latest_daily_k_line: null`。
- 资源不存在返回 `404` 和具体 not-found 错误。
- 资源存在但可选子数据缺失时，该子数据返回 `null`。

## 响应体

成功和错误响应使用统一响应 envelope：

```json
{
  "code": "OK",
  "message": "ok",
  "request_id": "req_xxx",
  "data": {}
}
```

成功响应使用 `code = "OK"` 和 `message = "ok"`。

第一阶段错误 message 使用稳定英文短语。客户端必须使用 `code` 而不是 `message` 进行程序化处理。

公共 API 不暴露内部数据库自增 ID。股票使用 `symbol`；同步任务使用 `task_uid`。

`trade_date`、`list_date`、`delist_date`、`cal_date` 和 `pretrade_date` 等日期字段使用 `YYYY-MM-DD`。

`synced_at`、`started_at`、`finished_at` 和 `created_at` 等 datetime 字段使用带时区的 ISO 8601，例如 `2026-06-27T10:00:00.123+08:00`。

Datetime 响应使用服务业务时区 `Asia/Shanghai`，因此通常携带 `+08:00`。

## 日 K 查询

规则：

- 默认范围为最近 1 年。
- 缺少 `end_date` 时默认使用该股票本地已同步的最新 `trade_date`。
- 缺少 `start_date` 时默认使用 `end_date` 往前 1 年。
- 如果该股票没有本地日 K，默认查询返回空数组。
- 最大范围为 5 年。
- 第一阶段结果不分页。
- 结果按 `trade_date ASC` 排序。
- 超过 5 年的范围返回 `MARKET_DATE_RANGE_TOO_LARGE`。

日 K API 字段使用市场数据短名：

```text
open
high
low
close
change
```

领域模型和数据库字段保留明确名称，例如 `OpenPrice`、`HighPrice`、`LowPrice`、`ClosePrice`、`ChangeAmt`、`open_price` 和 `change_amt`。

HTTP 响应中的市场数值字段是字符串，不是 JSON number。

日 K 响应暴露单位字段：

```text
pct_change_unit = percent
volume_unit     = lot
amount_unit     = thousand_cny
```

## 同步任务创建

同步创建端点会创建后台任务并立即返回。

响应 data：

```json
{
  "task_uid": "01J2Z3ABCDEF123456789XYZAB",
  "status": "PENDING"
}
```

规则：

- 同步创建不等待任务完成。
- 客户端通过查询任务详情端点确认完成。
- 第一阶段不提供 WebSocket、SSE、webhook callback 或阻塞等待端点。

任务状态流：

```text
PENDING -> RUNNING -> SUCCESS
PENDING -> RUNNING -> FAILED
PENDING -> RUNNING -> PARTIAL_SUCCESS
```

任务详情响应 data：

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

客户端轮询：

- 每 1 到 3 秒轮询 `GET /api/v1/market/sync/tasks/{task_uid}`。
- 当状态为 `SUCCESS`、`FAILED` 或 `PARTIAL_SUCCESS` 时停止轮询。
- 状态为 `FAILED` 或 `PARTIAL_SUCCESS` 时查询任务日志。

## 错误码

规则：

- Market 领域错误使用 `MARKET_` 前缀。
- 通用错误包括 `INVALID_ARGUMENT`、`UNAUTHORIZED` 和 `INTERNAL_ERROR`。
- HTTP status code 仍然承载协议语义。
- 除非操作确实成功，错误响应不会包装成 HTTP 200。

## 股票 API

### 股票列表

端点：

```http
GET /api/v1/market/stocks
```

Query 参数：

```text
page       optional, default 1
page_size  optional, default 20, max 100
keyword    optional, matches symbol or name
exchange   optional
status     optional
market     optional, default CN
asset_type optional, default STOCK
```

响应 item：

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

规则：

- 第一阶段列表端点不包含最新日 K。
- `keyword` 对 `symbol` 和 `name` 执行模糊匹配。
- `exchange` 和 `status` 是精确枚举过滤。

### 获取股票详情

端点：

```http
GET /api/v1/market/stocks/{symbol}
```

Query 参数：

```text
market     optional, default CN
asset_type optional, default STOCK
```

响应 data：

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

规则：

- 股票存在但没有同步日 K 时，`latest_daily_k_line` 为 `null`。
- 股票缺失返回 `404` 和 `MARKET_STOCK_NOT_FOUND`。
- 股票详情包含 `latest_daily_k_line`；股票列表 item 不包含。

## 日 K API

### 日 K 列表

端点：

```http
GET /api/v1/market/stocks/{symbol}/daily-k-lines
```

Query 参数：

```text
start_date optional, default latest 1 year
end_date   optional, default latest available date
market     optional, default CN
asset_type optional, default STOCK
```

响应 item：

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

规则：

- 响应 data 是数组，不是分页对象。
- 结果按 `trade_date ASC` 排序。
- 股票缺失返回 `404` 和 `MARKET_STOCK_NOT_FOUND`。
- 没有 K 线数据时返回空数组。

### 获取最新日 K

端点：

```http
GET /api/v1/market/stocks/{symbol}/latest-daily-k-line
```

规则：

- 股票缺失返回 `404` 和 `MARKET_STOCK_NOT_FOUND`。
- 已存在股票没有日 K 时返回 `data: null`。
- 已存在股票有数据时返回与日 K 列表相同的 item 形状。
- 即使股票详情也嵌入 `latest_daily_k_line`，最新日 K 端点仍然保留。

## 交易日历 API

### 交易日历列表

端点：

```http
GET /api/v1/market/trade-calendars
```

Query 参数：

```text
start_date optional
end_date   optional
market     optional, default CN
exchange   optional, default SSE
```

响应 item：

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

规则：

- 响应 data 是数组。
- 结果按 `cal_date ASC` 排序。
- 同时缺少 `start_date` 和 `end_date` 时默认最近 1 年。
- 日期范围是闭区间。
- 接受未来日期；没有行的日期不会出现在列表中。
- 超过 30 年的超大范围返回 `MARKET_DATE_RANGE_TOO_LARGE`。

### 获取最近开市日

端点：

```http
GET /api/v1/market/trade-calendars/latest-open-day
```

Query 参数：

```text
market   optional, default CN
exchange optional, default SSE
```

规则：

- 返回 `is_open = true` 的最新行。
- 如果日历数据缺失，返回 `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`。
- 最近开市日从请求的 `market + exchange` 下 `is_open = true` 的最大 `cal_date` 推导。
- 默认最近开市日使用 `market=CN` 和 `exchange=SSE`。
- 最近开市日不是 today，也不是股票本地最新日 K 日期。

### 检查是否开市

端点：

```http
GET /api/v1/market/trade-calendars/is-open
```

Query 参数：

```text
date     optional, default today
market   optional, default CN
exchange optional, default SSE
```

响应 data：

```json
{
  "market": "CN",
  "exchange": "SSE",
  "date": "2026-06-25",
  "is_open": true,
  "pretrade_date": "2026-06-24"
}
```

## 同步 API

所有同步创建端点都需要 Admin Token。

### 同步交易日历

端点：

```http
POST /api/v1/market/sync/trade-calendars
```

Request body：

```json
{
  "market": "CN",
  "exchange": "SSE",
  "start_date": "2021-01-01",
  "end_date": "2027-12-31"
}
```

规则：

- 缺少范围时默认最近 5 年加未来 1 年。
- 默认范围锚定服务业务时区中的当前日期。
- 同步范围是闭区间。
- 交易日历同步接受未来日期。
- 超过 30 年的超大范围返回 `MARKET_DATE_RANGE_TOO_LARGE`。
- 返回同步任务创建响应。

### 同步股票

端点：

```http
POST /api/v1/market/sync/stocks
```

Request body：

```json
{
  "market": "CN",
  "asset_type": "STOCK"
}
```

规则：

- 同步 `LISTED`、`DELISTED` 和 `PAUSED` 股票。
- 任一状态 fetch 失败都会使任务 `FAILED`。
- 返回同步任务创建响应。

### 同步日 K

端点：

```http
POST /api/v1/market/sync/daily-k-lines
```

Request body：

```json
{
  "market": "CN",
  "asset_type": "STOCK",
  "symbol": "600519",
  "start_date": "2023-01-01",
  "end_date": "2026-06-25"
}
```

规则：

- 缺少 `symbol` 表示全市场同步。
- 存在 `symbol` 表示单股票同步。
- 缺少 `start_date` 默认最近 3 年。
- 缺少 `end_date` 默认最近开市日。
- 全市场同步最大范围为 5 年。
- 单股票同步最大范围为 20 年。
- 全市场同步默认只同步 `LISTED` 股票。
- 全市场单股票失败不阻塞整个任务，可能产生 `PARTIAL_SUCCESS`。
- 缺少交易日历返回 `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`。
- 全市场同步无股票时返回 `MARKET_STOCKS_NOT_INITIALIZED`。
- 单股票同步股票缺失时返回 `MARKET_STOCK_NOT_FOUND`。
- 日 K 同步不会直接默认使用系统当前日期。
- 第一阶段使用 `CN + SSE` 作为默认 A 股交易日历，不按股票交易所动态选择日历。

### 获取同步任务

端点：

```http
GET /api/v1/market/sync/tasks/{task_uid}
```

规则：

- 任务缺失返回 `404` 和 `MARKET_SYNC_TASK_NOT_FOUND`。

### 同步日志列表

端点：

```http
GET /api/v1/market/sync/tasks/{task_uid}/logs
```

Query 参数：

```text
page      optional, default 1
page_size optional, default 20, max 100
status    optional
symbol    optional
```

规则：

- 响应使用标准分页。
- 日志按 `created_at ASC` 排序。
- 任务缺失返回 `404` 和 `MARKET_SYNC_TASK_NOT_FOUND`。

## 枚举 API

端点：

```http
GET /api/v1/market/meta/enums
```

推荐响应 data 形状：

```json
{
  "markets": ["CN"],
  "asset_types": ["STOCK"],
  "exchanges": ["SSE", "SZSE", "BSE"],
  "stock_statuses": ["LISTED", "DELISTED", "PAUSED"],
  "data_sources": ["TUSHARE", "MOCK"]
}
```

规则：

- 值从领域常量返回。
- 第一阶段不使用数据库字典表。
- 第一阶段不支持多语言枚举标签。
