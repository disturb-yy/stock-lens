# Market 领域包结构

第一阶段为 market 领域使用一个 Go package：

```text
domain/market/
  model.go
  port.go
  service.go
  handler.go
  dto.go
  errors.go
  repository_mysql.go
  repository_mock.go
  provider_tushare.go
  provider_mock.go
  INDEX.md
```

## 文件职责

`model.go` 包含 market 领域模型、枚举、常量和小型领域判断方法。它不包含 HTTP DTO、GORM struct 或 Tushare 原始协议 struct。

`port.go` 包含领域 service 使用的接口和 query/request struct，包括 repository、provider 和事务管理。它不包含外部协议类型。

`service.go` 通过 `QueryService` 和 `SyncService` 包含用例编排。它不直接依赖 Gin、GORM 或原始 Tushare client。

公开 service 方法对应 API 用例。`handler.go` 不应从较低层操作中组合业务流程；它应解析 HTTP 输入、调用一个 service 方法，并把结果映射为 HTTP 响应。

`handler.go` 包含 Gin HTTP 边界：路由绑定、请求解析、service 调用和 HTTP 响应/错误映射。它不包含同步业务逻辑，也不直接访问 repository/provider。

`dto.go` 包含 HTTP request/response DTO 和 DTO 转换 helper。Decimal 市场数值以字符串暴露，市场数据单位作为响应字段暴露。

`errors.go` 包含 market 领域错误、错误码，以及用于分类领域错误的 helper。

`repository_mysql.go` 包含 MySQL/GORM repository 实现、表 struct、upsert 逻辑、查询逻辑和事务处理。GORM 类型在此文件中保持私有。

`repository_mock.go` 包含用于开发和测试的内存 repository 实现。Mock 行为应尽量贴近 MySQL 实现，包括 upsert、分页、过滤和唯一键语义。

`provider_tushare.go` 包含 Tushare provider 适配器，包括状态转换、`ts_code` 到 `symbol` 的转换、board 推导、日期/decimal/单位映射和 provider 错误包装。它可以调用原始 `pkg/tushare` client。

`provider_mock.go` 包含用于本地开发、单元测试和无网络测试流程的 mock provider 适配器。

`INDEX.md` 包含简短的包指南：目的、文件映射、主要入口和边界规则。它不应重复完整 API 或表结构规格。

## Repository 接口拆分

Repository 接口按数据职责保持拆分：

```text
StockRepository
KLineRepository
TradeCalendarRepository
SyncTaskRepository
```

第一阶段不将它们合并成单一 `MarketRepository`。拆分接口让 `QueryService` 和 `SyncService` 的依赖保持明确，使测试更容易限定范围，并避免出现方法并不总是一起使用的宽接口。

## 边界规则

Market 领域将外部实现细节保持在边缘：

- Gin 限制在 `handler.go`。
- HTTP DTO 限制在 `dto.go`。
- GORM 限制在 `repository_mysql.go`。
- 原始 Tushare 协议处理限制在 `pkg/tushare` 和 `provider_tushare.go`。
- 领域 service 依赖 `port.go` 中的接口，而不是具体基础设施 package。

## 命名规则

Go identifier 使用 `CamelCase` 或 `camelCase`。JSON 字段、HTTP query/body 参数和 SQL identifier 使用 `snake_case`。枚举值保持大写英文 token，例如 `PARTIAL_SUCCESS`。

当 Go DTO 字段表示 API 字段时，字段使用 Go 命名并添加 `snake_case` JSON tag，例如 `TaskUID string `json:"task_uid"`。

## 接口约定

Provider 接口按外部能力保持拆分：

```text
InstrumentProvider
MarketDataProvider
CalendarProvider
```

第一阶段不将它们合并成单一 `MarketProvider`。

Repository 接收并返回领域模型，而不是 GORM record 或 HTTP DTO。Provider 返回领域模型，而不是原始 Tushare 协议 struct。Repository query struct 和 provider request struct 属于 `port.go`。

Handler 不把 Gin context 或 HTTP DTO 传入 service。Service 返回 market 领域错误，handler 将这些错误映射为 HTTP status code 和响应体。

具体 repository 实现、provider 实现和 GORM record struct 应保持私有。Constructor 暴露 `StockRepository`、`MarketDataProvider` 或 `TxManager` 等接口。

Mock repository 和 provider 也应使用 constructor。测试可以通过 option 或显式方法注入 mock 数据，而不是修改公开的内部 map。

`TxManager` 只暴露 `WithTx(ctx, fn)`。`TxRepositories` 只暴露 repository 接口，绝不暴露原始数据库 handle。`QueryService` 不依赖 `TxManager`；同步写入流程可以通过 `SyncService` 使用它。

`Handler` 只持有 service。Market 端点的路由注册位于 market package，外层 router 拥有 `/api/v1` 前缀。Admin Token 检查属于 middleware 和受保护 route group，而不是 `SyncService`。

枚举响应从领域常量生成，而不是从数据库字典表生成。DTO 转换从领域模型到 HTTP response 单向进行；领域模型不依赖 DTO。
