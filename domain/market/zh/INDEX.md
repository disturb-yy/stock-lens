# Market 领域

`domain/market` 是第一阶段的市场数据领域包。它拥有股票主数据、原始日 K、交易日历、同步任务和同步日志。

## 文件

```text
model.go              领域模型、枚举、常量和小型判断方法
port.go               repository、provider、事务接口，以及 service 查询/请求结构
service.go            QueryService 和 SyncService 用例编排
handler.go            Gin HTTP 边界和路由 handler
dto.go                HTTP request/response DTO 和转换 helper
errors.go             market 领域错误和错误码
repository_mysql.go   MySQL/GORM repository 和事务实现
repository_mock.go    用于开发和测试的内存 repository 实现
provider_tushare.go   Tushare provider 适配器和 provider 到领域模型的映射
provider_mock.go      用于开发和测试的 mock provider 适配器
INDEX.md              包导览和边界规则
```

## 入口

- `Handler` 是 HTTP 入口。
- `QueryService` 处理读侧市场查询。
- `SyncService` 创建并运行市场数据同步任务。

公开的 `QueryService` 和 `SyncService` 方法对应 API 用例。Handler 解析 HTTP 输入并映射响应，但不从更底层的领域操作中组合业务流程。

## 边界规则

- 第一阶段保持为一个 Go package。
- 按文件区分职责，不按子 package 区分。
- 将 Gin 使用限制在 `handler.go`。
- 将 HTTP DTO 限制在 `dto.go`。
- 将 GORM 使用限制在 `repository_mysql.go`。
- 将原始 Tushare 协议处理限制在 `pkg/tushare` 和 `provider_tushare.go`。
- Service 依赖 `port.go` 中的接口，不依赖具体基础设施类型。

## 命名规则

- Go identifier 使用 `CamelCase` 或 `camelCase`，例如 `TaskUID`、`PageSize` 和 `LatestDailyKLine`。
- JSON 字段、HTTP query/body 参数和 SQL identifier 使用 `snake_case`，例如 `task_uid`、`page_size` 和 `latest_daily_k_line`。
- 枚举值保持大写英文 token，例如 `CN`、`STOCK`、`SYNC_DAILY_K_LINES` 和 `PARTIAL_SUCCESS`。
- Go DTO 字段使用 Go 命名并添加 `snake_case` JSON tag。

## Repository 接口

Repository 接口按数据职责拆分：

- `StockRepository`
- `KLineRepository`
- `TradeCalendarRepository`
- `SyncTaskRepository`

第一阶段不要把它们合并为一个 `MarketRepository`。拆分接口能让 service 依赖保持明确，同时仍然保留扁平的 `domain/market` package 形状。

## 接口和边界约定

- Provider 接口保持拆分为 `InstrumentProvider`、`MarketDataProvider` 和 `CalendarProvider`。
- Repository 接收并返回领域模型，而不是 GORM record 或 HTTP DTO。
- Provider 返回领域模型，而不是原始 Tushare 协议结构。
- 将 `ListStocksQuery`、`FindStockQuery` 等 repository 查询结构放在 `port.go`。
- 将 `FetchStocksRequest`、`FetchDailyKLinesRequest` 等 provider 请求结构放在 `port.go`。
- 不要把 Gin context 或 HTTP DTO 传入 service。
- 为复杂用例定义 service input struct，并让它们靠近使用它们的 service 方法。
- Service 返回 market 领域错误；handler 将这些错误映射为 HTTP status code 和响应体。
- 将 GORM record struct 保持为 `repository_mysql.go` 内的私有类型。
- 将具体 repository 实现保持为私有类型，并暴露返回接口的 constructor。
- 优先使用私有 provider 实现结构体，并通过 constructor 返回 provider 接口。
- 为 mock repository 和 provider 提供 constructor；不要把内部 map 作为公开状态暴露。
- 将 `TxManager` 限制为 `WithTx(ctx, fn)`。
- `TxRepositories` 只暴露 repository 接口，不暴露原始数据库 handle。
- `QueryService` 不依赖 `TxManager`；`SyncService` 可在同步写入中使用它。
- `Handler` 只持有 service，不持有 repository 或 provider。
- market 路由从 market package 注册，外层 router 拥有 `/api/v1` 前缀。
- 将 Admin Token 检查放在 middleware 和受保护 route group 中，不放在 `SyncService`。
- 枚举响应从领域常量生成，不使用数据库字典表。
- DTO 转换只从领域模型到 HTTP response 单向进行；领域模型不依赖 DTO。

## 同步任务规则

- 第一阶段同一时间只允许一个 `PENDING` 或 `RUNNING` 同步任务存在。
- 同步创建会持久化 `PENDING` 并立即返回；后台 goroutine 将任务标记为 `RUNNING`。
- 终态任务状态为 `SUCCESS`、`FAILED` 和 `PARTIAL_SUCCESS`。
- 无效请求输入在任务创建前被拒绝。
- 缺少同步前置条件会让已创建的任务失败。
- 全市场日 K 同步可以以 `PARTIAL_SUCCESS` 结束；股票主数据同步、交易日历同步和单股票日 K 同步要么成功，要么失败。
- 全市场日 K 进度按股票计数，而不是按 K 线行数计数。
- 同步日志是任务级和股票级摘要，不记录每条 K 线。

## 错误映射

- Market 参数错误映射到 HTTP `400`。
- 资源不存在映射到 HTTP `404`。
- 缺少本地前置条件和 active task 冲突映射到 HTTP `409`。
- Provider 失败映射到 HTTP `502`。
- Store 和未分类任务失败映射到 HTTP `500`。
- 无效参数、active task 冲突和 Admin Token 失败不创建同步任务。
- 同步执行期间发现缺少前置条件时，创建失败的同步任务。

## API 形状规则

- 成功响应使用 `code = "OK"` 和 `message = "ok"`。
- 错误 message 是稳定英文短语；客户端用错误码处理逻辑。
- API 响应永不暴露内部数据库自增 ID。
- 股票通过 `symbol` 标识；同步任务通过 `task_uid` 标识。
- 日期字段使用 `YYYY-MM-DD`；datetime 字段使用带时区的 ISO 8601。
- 日 K API 字段使用 `open`、`high`、`low`、`close` 和 `change`；领域和数据库字段保留明确的 price/amount 名称。
- HTTP market 数值字段使用字符串。
- 日 K 响应包含涨跌幅、成交量和成交额的显式单位字段。
- 股票列表 item 不包含最新日 K；股票详情包含 `latest_daily_k_line`。
- 最新日 K 端点保留为专用读取端点。
- 枚举 API 响应来自领域常量，第一阶段不包含多语言 label。

## 日历和日期规则

- 第一阶段交易日历行为默认使用 `market=CN` 和 `exchange=SSE`。
- 最近开市日从 `trade_calendars` 中某个 `market + exchange` 下最大的开市 `cal_date` 推导。
- 最近开市日不同于 today，也不同于股票本地最新日 K 日期。
- 日 K 同步缺少 `end_date` 时默认使用最近开市日。
- 日 K 查询缺少 `end_date` 时默认使用该股票本地最新 `trade_date`；如果没有本地日 K，返回空数组。
- 服务业务时区为 `Asia/Shanghai`；日期边界和 `today` 使用该时区。
- 第一阶段日 K 同步不按股票交易所动态选择交易日历。

## Provider 规则

- 原始 Tushare 协议处理位于 `pkg/tushare`；market 适配位于 `provider_tushare.go`。
- Provider 返回 market 领域模型，而不是原始 Tushare response 结构。
- Tushare provider 输出使用 `DataSource = TUSHARE`；mock provider 输出使用 `DataSource = MOCK`。
- `stocks.ts_code` 存储 Tushare 身份；日 K 不存储 `ts_code`。
- Tushare `ts_code` 映射为系统 `symbol`；无效 provider 标识是 mapping error。
- Tushare 数值解析为 `decimal.Decimal`；无效数字字段是 mapping error。
- Provider 日志不得包含 Tushare token 或完整的含 token 请求。
- 第一阶段不运行真实 Tushare 网络测试。

## 战略上下文

第一阶段是市场数据底座，不是完整股票分析产品。Market 领域应优先考虑数据正确性、同步可恢复性和 API 稳定性，再考虑分析功能或基础设施扩展。

## 运行时和测试默认值

- 默认 HTTP 服务端口为 `30078`。
- `/healthz` 表示进程存活。
- `/readyz` 表示服务已经就绪，包括必要的启动校验和 MySQL 连通性。
- 单元测试应优先使用表格驱动写法，并选择最小有用测试范围。
- 当接口替换不实际时，单元测试可以使用 `gomonkey`。
- 依赖 monkey patch 的测试必须考虑 Go 内联，可能需要 `go test ./... -gcflags=all="-N -l"`。
