# Market 错误码

本文档冻结第一阶段 market 错误码和 HTTP status 映射。

## 响应形状

错误使用统一 API 响应体：

```json
{
  "code": "MARKET_STOCK_NOT_FOUND",
  "message": "stock not found",
  "request_id": "req_xxx",
  "data": null
}
```

当同步任务与 active task 冲突时，`data` 可以包含 active task：

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

## 通用错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `INVALID_ARGUMENT` | 400 | 没有更具体 market 错误适用时的通用参数错误 |
| `UNAUTHORIZED` | 401 | Admin Token 缺失或无效 |
| `NOT_FOUND` | 404 | 没有更具体 market 错误适用时的通用资源不存在 |
| `INTERNAL_ERROR` | 500 | 未分类内部错误 |

## Market 参数错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `MARKET_INVALID_SYMBOL` | 400 | Symbol 不是 6 位 A 股代码，或传入了 `600519.SH` 这样的 provider 专用代码 |
| `MARKET_INVALID_MARKET` | 400 | 第一阶段不支持该 market；只接受 `CN` |
| `MARKET_INVALID_ASSET_TYPE` | 400 | 第一阶段不支持该 asset type；只接受 `STOCK` |
| `MARKET_INVALID_EXCHANGE` | 400 | Exchange 不是 `SSE`、`SZSE` 或 `BSE` |
| `MARKET_INVALID_STATUS` | 400 | Stock status 不是 `LISTED`、`DELISTED` 或 `PAUSED` |
| `MARKET_INVALID_DATE_RANGE` | 400 | `start_date` 晚于 `end_date` |
| `MARKET_DATE_RANGE_TOO_LARGE` | 400 | 查询或同步日期范围超过允许的最大值 |
| `MARKET_INVALID_TASK_STATUS` | 400 | Task 或 log status 过滤条件无效 |

## Market 资源错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `MARKET_STOCK_NOT_FOUND` | 404 | 股票不存在 |
| `MARKET_SYNC_TASK_NOT_FOUND` | 404 | 同步任务不存在 |
| `MARKET_TRADE_CALENDAR_NOT_FOUND` | 404 | 查询的指定交易日历日期不存在 |

## Market 前置条件错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `MARKET_TRADE_CALENDAR_NOT_INITIALIZED` | 409 | 最新开市日不可用，或日 K 同步需要的本地交易日历数据尚未初始化 |
| `MARKET_STOCKS_NOT_INITIALIZED` | 409 | 全市场日 K 同步需要的本地股票主数据尚未初始化 |

第一阶段不定义单独的 `MARKET_LATEST_OPEN_DAY_NOT_FOUND`。使用 `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`。

## Market 冲突错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `MARKET_SYNC_TASK_CONFLICT` | 409 | 已存在 `PENDING` 或 `RUNNING` 同步任务 |

## Provider 和 Store 错误

| Error Code | HTTP | 含义 |
|---|---:|---|
| `MARKET_PROVIDER_ERROR` | 502 | Tushare 网络错误、限流、响应格式异常、空股票列表或空交易日历 |
| `MARKET_STORE_ERROR` | 500 | Repository、query、upsert 或 transaction 失败 |
| `MARKET_SYNC_TASK_FAILED` | 500 | 同步任务失败且没有更具体错误码 |

## 空结果规则

| 场景 | 结果 |
|---|---|
| 查询未来日期且无数据 | `200`，返回空数组 |
| 已存在股票没有最新日 K | `200`，返回 `data: null` |
| 股票列表没有匹配项 | `200`，返回 `items: []` |
| 交易日历列表没有匹配项 | `200`，返回 `[]` |
| 查询的股票不存在 | `404`，返回 `MARKET_STOCK_NOT_FOUND` |
| 无法从本地日历数据推导最新开市日 | `409`，返回 `MARKET_TRADE_CALENDAR_NOT_INITIALIZED` |

## 同步任务创建规则

这些错误会在任务创建前返回，且不会创建失败的同步任务：

- invalid argument
- invalid date
- date range too large
- active task conflict
- missing or invalid Admin Token

这些错误会先创建任务，然后将任务标记为 `FAILED`：

- missing trade-calendar prerequisite
- missing stock master prerequisite
- missing stock for single-stock daily K sync
- provider error during execution
- store error during execution

如果 store error 在任务持久化前发生，则不创建任务。
