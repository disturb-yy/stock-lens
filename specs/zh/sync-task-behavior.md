# 同步任务行为

第一阶段同步任务由手动触发、全局互斥，并由单个服务实例的 goroutine 执行。

## 并发

同一时间只允许一个 active sync task 存在。Active task 是任何状态为 `PENDING` 或 `RUNNING` 的任务。

当已有其他 active task 时收到同步创建请求，API 返回 `409 Conflict` 和 `MARKET_SYNC_TASK_CONFLICT`。如果能找到该 active task，响应应包含它的 `task_uid`。

第一阶段不支持任务取消、定时同步、队列、分布式 worker 或重启后的任务恢复。

## 状态流转

允许的任务状态流转：

```text
PENDING -> RUNNING -> SUCCESS
PENDING -> RUNNING -> FAILED
PENDING -> RUNNING -> PARTIAL_SUCCESS
```

终态不再变更。

同步创建端点先持久化 `PENDING` 任务并立即返回。后台 goroutine 在开始执行时将任务标记为 `RUNNING`。

服务启动时，将遗留的 `PENDING` 和 `RUNNING` 任务标记为 `FAILED`，原因可以是 `service restarted before task completed`。

如果任务执行 panic，goroutine 会 recover、写同步日志、将任务标记为 `FAILED`，并在 `error_msg` 中保存 panic 摘要。

## 前置校验

无效客户端输入会在任务创建前被拒绝。这些错误不会创建失败任务。

示例：

- invalid symbol
- invalid enum value
- invalid date format
- `start_date > end_date`
- date range larger than the allowed maximum

日期范围违规返回 `MARKET_DATE_RANGE_TOO_LARGE`。

依赖现有本地数据的同步任务在缺少前置条件时失败：

- missing trade calendar：`FAILED`，错误为 `MARKET_TRADE_CALENDAR_NOT_INITIALIZED`
- full-market daily K sync without stock master data：`FAILED`，错误为 `MARKET_STOCKS_NOT_INITIALIZED`
- single-stock daily K sync for a missing stock：`FAILED`，错误为 `MARKET_STOCK_NOT_FOUND`

## 同步范围

股票主数据同步会获取并存储 `LISTED`、`DELISTED` 和 `PAUSED` 股票。如果任一股票状态获取失败，整个任务失败。

全市场日 K 同步默认只同步 `LISTED` 股票。单股票日 K 同步可对 `LISTED`、`PAUSED` 或 `DELISTED` 股票运行。

日 K 同步缺少 `end_date` 时默认使用最新开市交易日。缺少 `start_date` 时默认使用 `end_date` 往前三年。

全市场日 K 同步最多允许 5 年。单股票日 K 同步最多允许 20 年。

交易日历同步缺少范围时，默认使用最近 5 年加未来 1 年。

第一阶段使用 `market=CN` 和 `exchange=SSE` 作为默认 A 股交易日历。日 K 同步不按股票交易所动态选择交易日历。

最近开市日从某个 `market + exchange` 下 `is_open = true` 的最大 `cal_date` 推导。它不等于 today，也不等于股票本地最新日 K 日期。

服务业务时区为 `Asia/Shanghai`。日期边界、默认范围和 `today` 都按该时区解释。日 K 同步默认值使用最近开市日，而不是系统当前日期。

交易日历同步范围是闭区间，可以包含未来日期；超过 30 年的范围返回 `MARKET_DATE_RANGE_TOO_LARGE`。

## Provider 空结果

Provider 空结果处理：

- single-stock daily K empty result：`SUCCESS`，`affected_rows = 0`
- one stock in full-market daily K empty result：股票级 `SUCCESS`，`affected_rows = 0`
- trade-calendar empty result：`FAILED`，错误为 `MARKET_PROVIDER_ERROR`
- stock-list empty result：`FAILED`，错误为 `MARKET_PROVIDER_ERROR`

空 K 线数据对某个日期范围可能是有效结果。空股票列表和空交易日历在第一阶段被视为 provider 或映射失败。

## 失败语义

全市场日 K 同步按股票隔离失败。每只股票的 provider 或 upsert 失败会被记录，处理继续进入剩余股票。

全市场日 K 同步最终状态规则：

- 所有股票成功：`SUCCESS`
- 至少一只股票成功且至少一只失败：`PARTIAL_SUCCESS`
- 所有股票失败：`FAILED`

对于股票主数据同步、交易日历同步和单股票日 K 同步，provider 错误、网络错误和 upsert 失败都会使任务失败。

第一阶段不执行自动重试。Provider 实现以后可以添加小型内部重试，但任务模型不依赖重试语义。

## 进度计数

任务计数器使用对象级进度：

```text
total_items      = expected number of objects to process
processed_items  = completed objects, whether successful or failed
success_items    = successful objects
failed_items     = failed objects
```

对于全市场日 K 同步，对象是股票，不是单条 K 线。

## 同步日志

同步日志是任务级和股票级摘要。它们不为每条 K 线记录一行。

推荐日志点：

- task start
- preflight check
- provider fetch summary
- batch upsert summary
- per-stock daily K success or failure summary
- task end

日志状态为 `SUCCESS`、`FAILED` 和 `WARNING`。

`affected_rows` 含义：

- stock sync：upsert 的股票行数
- trade-calendar sync：upsert 的日历行数
- single-stock daily K sync：upsert 的 K 线行数
- full-market daily K sync：该股票级日志中 upsert 的 K 线行数
